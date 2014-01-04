// +build ignore

/*
Copyright 2013 Brandon Philips

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// This program builds a project and is a copy of path.go. See
// github.com/philips/path.go
//
// $ go run path.go
//
// See the README file for more details.
package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

const (
	DefaultThirdParty = "third_party"
)

// thirdPartyDir creates a string path to the third_party directory based on
// the current working directory.
func thirdPartyDir() string {
	root, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get the current working directory: %v", err)
	}
	return path.Join(root, DefaultThirdParty)
}

func srcDir() string {
	return path.Join(thirdPartyDir(), "src")
}

// binDir creates a string path to the GOBIN directory based on the current
// working directory.
func binDir() string {
	root, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get the current working directory: %v", err)
	}
	return path.Join(root, "bin")
}

// runEnv execs a command like a shell script piping everything to the parent's
// stderr/stdout and uses the given environment.
func runEnv(env []string, name string, arg ...string) {
	cmd := exec.Command(name, arg...)

	cmd.Env = env

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
	err = cmd.Start()
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
	go io.Copy(os.Stdout, stdout)
	go io.Copy(os.Stderr, stderr)
	cmd.Wait()
}

// run calls runEnv with the GOPATH third_party packages.
func run(name string, arg ...string) {
	env := append(os.Environ(),
		"GOPATH="+thirdPartyDir(),
		"GOBIN="+binDir(),
	)

	runEnv(env, name, arg...)
}

// setupProject does the initial setup of the third_party src directory
// including setting up the symlink to the cwd from the src directory.
func setupProject(pkg string) {
	root, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get the current working directory: %v", err)
	}

	src := path.Join(thirdPartyDir(), "src", pkg)
	srcdir := path.Dir(src)

	os.MkdirAll(srcdir, 0755)

	rel, err := filepath.Rel(srcdir, root)
	if err != nil {
		log.Fatalf("creating relative third party path: %v", err)
	}

	err = os.Symlink(rel, src)
	if err != nil && os.IsExist(err) == false {
		log.Fatalf("creating project third party symlink: %v", err)
	}
}

// commit grabs the commit id from hg or git as a string.
func commit(dir string) string {
	base := path.Base(dir)
	switch base {
	case ".git":
		return commitGit(dir)
	case ".hg":
		return commitHg(dir)
	default:
		return ""
	}
}

// commitGit returns the current HEAD commit hash for a given git dir.
func commitGit(dir string) string {
	out, err := exec.Command("git", "--git-dir="+dir, "rev-parse", "HEAD").Output()
	if err != nil {
		return ""
	}
	return string(out)
}

// commitHg returns the current HEAD commit hash for a given git dir.
func commitHg(dir string) string {
	out, err := exec.Command("hg", "id", "-i", "-R", dir).Output()
	if err != nil {
		return ""
	}
	return string(out)
}

// removeVcs removes a .git or .hg directory from the given root if it exists.
func removeVcs(root string) (bool, string) {
	for _, v := range []string{".git", ".hg"} {
		r := path.Join(root, v)
		info, err := os.Stat(r)

		if err != nil {
			continue
		}

		// We didn't find it, next!
		if info.IsDir() == false {
			continue
		}

		// We found it, grab the commit and remove the directory
		c := commit(r)
		err = os.RemoveAll(r)
		if err != nil {
			log.Fatalf("removeVcs: %v", err)
		}
		return true, c
	}

	return false, ""
}

// bump takes care of grabbing a package, getting the package git hash and
// removing all of the version control stuff.
func bump(pkg string) {
	tpd := thirdPartyDir()

	temp, _ := ioutil.TempDir(tpd, "bump")
	defer os.RemoveAll(temp)

	env := append(os.Environ(),
		"GOPATH="+temp,
	)

	runEnv(env, "go", "get", "-u", "-d", pkg)

	for {
		root := path.Join(temp, "src", pkg) // the temp installation root
		home := path.Join(tpd, "src", pkg)  // where the package will end up

		ok, c := removeVcs(root)
		if ok {
			// Create the path leading up to the package
			err := os.MkdirAll(path.Dir(home), 0755)
			if err != nil {
				log.Fatalf("bump: %v", err)
			}

			// Remove anything that might have been there
			err = os.RemoveAll(home)
			if err != nil {
				log.Fatalf("bump: %v", err)
			}

			// Finally move the package
			err = os.Rename(root, home)
			if err != nil {
				log.Fatalf("bump: %v", err)
			}

			fmt.Printf("%s %s\n", pkg, strings.TrimSpace(c))
			break
		}

		// Pop off and try to find this directory!
		pkg = path.Dir(pkg)
		if pkg == "." {
			return
		}
	}
}

// validPkg uses go list to decide if the given path is a valid go package.
// This is used by the bumpAll walk to bump all of the existing packages.
func validPkg(pkg string) bool {
	env := append(os.Environ(),
		"GOPATH="+thirdPartyDir(),
	)
	cmd := exec.Command("go", "list", pkg)
	cmd.Env = env

	out, err := cmd.Output()
	if err != nil {
		return false
	}

	if pkg == strings.TrimSpace(string(out)) {
		return true
	}

	return false
}

// bumpWalk walks the third_party directory and bumps all of the packages that it finds.
func bumpWalk(path string, info os.FileInfo, err error) error {
	if err != nil {
		return nil
	}

	// go packages are always directories
	if info.IsDir() == false {
		return nil
	}

	parts := strings.Split(path, srcDir()+"/")
	if len(parts) == 1 {
		return nil
	}

	pkg := parts[1]

	if validPkg(pkg) == false {
		return nil
	}

	bump(pkg)

	return nil
}

func bumpAll() {
	err := filepath.Walk(srcDir(), bumpWalk)
	if err != nil {
		log.Fatalf(err.Error())
	}
}

func main() {
	log.SetFlags(0)

	if len(os.Args) <= 1 {
		log.Fatalf("No commnad")
	}

	cmd := os.Args[1]

	if cmd == "setup" && len(os.Args) > 2 {
		setupProject(os.Args[2])
		return
	}

	if cmd == "bump" && len(os.Args) > 2 {
		bump(os.Args[2])
		return
	}

	if cmd == "bump-all" && len(os.Args) > 1 {
		bumpAll()
		return
	}

	run("go", os.Args[1:]...)
}
