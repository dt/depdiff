package glide

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"syscall"

	"github.com/dt/glide-diff-parser/diff"

	yaml "gopkg.in/yaml.v2"
)

type dep struct {
	Name    string
	Version string
}

type lockfile struct {
	Imports []dep
}

// ReadFromGit returns the version info as of a the given tree-ish.
func ReadFromGit(treeish string) (diff.Versions, error) {
	out, err := exec.Command("git", "show", fmt.Sprintf("%s:glide.lock", treeish)).Output()
	if err != nil {
		return nil, err
	}
	return parseGlildeYAML(out)
}

// ReadFromFs returns the version info in the current working directory.
func ReadFromFs() (diff.Versions, error) {
	out, err := ioutil.ReadFile("glide.lock")
	if err != nil {
		return nil, err
	}
	return parseGlildeYAML(out)
}

// LocalFileChanged returns if the file in the current directory has uncommited changes.
func LocalFileChanged() (bool, error) {
	err := exec.Command("git", "diff", "--quiet", "glide.lock").Run()
	if err == nil {
		return false, nil
	}
	if x, ok := err.(*exec.ExitError); ok && x.Sys().(syscall.WaitStatus).ExitStatus() == 1 {
		return true, nil
	}
	return false, err
}

func parseGlildeYAML(content []byte) (diff.Versions, error) {
	l := lockfile{}
	err := yaml.Unmarshal(content, &l)
	if err != nil {
		return nil, err
	}
	m := make(map[string]string, len(l.Imports))
	for _, dep := range l.Imports {
		m[dep.Name] = dep.Version
	}
	return m, nil
}
