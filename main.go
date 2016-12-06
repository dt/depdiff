package main

import (
	"fmt"
	"log"
	"os/exec"

	"gopkg.in/yaml.v2"
)

type lockfile struct {
	A string
	B struct {
		RenamedC int   `yaml:"c"`
		D        []int `yaml:",flow"`
	}
}

func readFromGit(treeish string) (lockfile, error) {
	l := lockfile{}
	out, err := exec.Command("git", "show", fmt.Sprintf("%s:glide.lock", treeish)).Output()
	if err != nil {
		return l, err
	}
	err = yaml.Unmarshal(out, &l)
	if err != nil {
		return l, err
	}
	return l, nil
}

func main() {
	before, err := readFromGit("HEAD~")
	if err != nil {
		log.Fatal(err)
	}
	after, err := readFromGit("HEAD")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%v \n\n%v\n\n", before, after)
}
