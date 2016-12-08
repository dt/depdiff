package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"
	"syscall"

	"gopkg.in/yaml.v2"
)

var rewrites = map[string]string{
	"golang.org/x":                "github.com/golang",
	"google.golang.org/appengine": "github.com/golang/appengine",
	"google.golang.org/grpc":      "github.com/grpc/grpc-go",
	"gopkg.in/inf.v0":             "github.com/go-inf/inf",
	"honnef.co/go/":               "github.com/dominikh/go-",
}

// Versions a map of dependency to version.
type Versions map[string]string

type dep struct {
	Name    string
	Version string
}

type lockfile struct {
	Imports []dep
}

func readFromGit(treeish string) (Versions, error) {
	out, err := exec.Command("git", "show", fmt.Sprintf("%s:glide.lock", treeish)).Output()
	if err != nil {
		return nil, err
	}
	return parseGlildeYAML(out)
}

func readFromFs() (Versions, error) {
	fp, err := os.Open("glide.lock")
	if err != nil {
		return nil, err
	}
	defer fp.Close()
	out, err := ioutil.ReadAll(fp)
	if err != nil {
		return nil, err
	}
	return parseGlildeYAML(out)
}

func parseGlildeYAML(content []byte) (Versions, error) {
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

type change struct {
	name string
	from string
	to   string
}

type changes []change

func (c changes) Len() int           { return len(c) }
func (c changes) Less(i, j int) bool { return c[i].name < c[j].name }
func (c changes) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }

var _ sort.Interface = changes{}

// Diff represents the changes between two sets of versions.
type Diff struct {
	changed changes
	added   []string
	removed []string
}

func compareLink(i change) string {
	repo := i.name
	for prefix, github := range rewrites {
		if strings.HasPrefix(repo, prefix) {
			repo = strings.Replace(repo, prefix, github, 1)
		}
	}
	if strings.HasPrefix(repo, "github.com") {
		return fmt.Sprintf("https://%s/compare/%s...%s", repo, i.from[:8], i.to[:8])
	}
	return ""
}

func (d diff) Print() {
	if d.added != nil {
		fmt.Println("Added:")
		for i := range d.added {
			fmt.Printf(" - %s\n", i)
		}
	}
	if d.changed != nil {
		fmt.Println("Changed:")
		for _, i := range d.changed {
			fmt.Printf(" - %s: %s -> %s\n", i.name, i.from[:8], i.to[:8])
			if link := compareLink(i); link != "" {
				fmt.Printf("   - %s\n", link)
			}
		}
	}
	if d.removed != nil {
		fmt.Println("Removed:")
		for d := range d.removed {
			fmt.Printf(" - %s\n", d)
		}
	}
}

func (d diff) Open() {
	for _, i := range d.changed {
		if link := compareLink(i); link != "" {
			exec.Command("open", link).Run()
		}
	}
}

func getDiff(before, after map[string]string) (diff, error) {
	d := diff{}
	for name, new := range after {
		old, ok := before[name]
		if !ok {
			d.added = append(d.added, name)
		} else if new != old {
			d.changed = append(d.changed, change{name: name, from: old, to: new})
		}
	}
	for name := range before {
		if _, ok := after[name]; !ok {
			d.removed = append(d.removed, name)
		}
	}
	sort.Strings(d.added)
	sort.Strings(d.removed)
	sort.Sort(d.changed)
	return d, nil
}

func locallyChanged() (bool, error) {
	err := exec.Command("git", "diff", "--quiet", "glide.lock").Run()
	if err == nil {
		return false, nil
	}
	if x, ok := err.(*exec.ExitError); ok && x.Sys().(syscall.WaitStatus).ExitStatus() == 1 {
		return true, nil
	}
	return false, err
}

func main() {
	open := flag.Bool("open", false, "attempt to open github in browser for each changed dependency")

	flag.Parse()

	args := flag.Args()
	from, to := "HEAD~", "HEAD"
	if len(args) > 0 {
		from = args[0]
	}
	if len(args) > 1 {
		to = args[1]
	}

	diff, err := getDiff(from, to)
	if err != nil {
		log.Fatal(err)
	}

	diff.Print()

	if *open {
		diff.Open()
	}
}
