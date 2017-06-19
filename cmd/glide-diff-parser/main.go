package main

import (
	"flag"
	"log"
	"os"

	yaml "gopkg.in/yaml.v2"

	"github.com/dt/glide-diff-parser/lockfile"
)

const lockfileName = "glide.lock"

func parseGlildeYAML(content []byte) (lockfile.Versions, error) {
	l := struct {
		Imports []struct {
			Name    string
			Version string
		}
	}{}
	if err := yaml.Unmarshal(content, &l); err != nil {
		return nil, err
	}
	m := make(map[string]string, len(l.Imports))
	for _, dep := range l.Imports {
		m[dep.Name] = dep.Version
	}
	return m, nil
}

func main() {
	verbose := flag.Bool("v", false, "print a detailed summary of added, removed and changed dependencies")
	flag.Usage = lockfile.Usage(lockfileName)
	flag.Parse()

	args := flag.Args()
	if len(args) > 2 {
		flag.Usage()
		os.Exit(-1)
	}
	if err := lockfile.SummarizeDiff(args, *verbose, lockfileName, parseGlildeYAML); err != nil {
		log.Fatal(err)
	}
}
