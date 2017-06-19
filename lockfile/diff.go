package lockfile

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

// Versions a map of dependency to version.
type Versions map[string]string

type update struct {
	name string
	from string
	to   string
}

type updates []update

func (c updates) Len() int           { return len(c) }
func (c updates) Less(i, j int) bool { return c[i].name < c[j].name }
func (c updates) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }

var _ sort.Interface = updates{}

// Changes represents the changes between two sets of versions.
type Changes struct {
	updates updates
	added   []string
	removed []string
}

// GithubPrefixes maps non-github import path prefixe to github repos, for
// generating comparison links.
var GithubPrefixes = map[string]string{
	"golang.org/x":                "github.com/golang",
	"google.golang.org/appengine": "github.com/golang/appengine",
	"google.golang.org/grpc":      "github.com/grpc/grpc-go",
	"gopkg.in/inf.v0":             "github.com/go-inf/inf",
	"gopkg.in/yaml.v2":            "github.com/go-yaml/yaml",
	"honnef.co/go/":               "github.com/dominikh/go-",
	"cloud.google.com/go":         "github.com/GoogleCloudPlatform/google-cloud-go",
	"google.golang.org/api":       "github.com/google/google-api-go-client",
}

func (u update) compareLink() string {
	repo := u.name
	for prefix, github := range GithubPrefixes {
		if strings.HasPrefix(repo, prefix) {
			repo = strings.Replace(repo, prefix, github, 1)
		}
	}
	if strings.HasPrefix(repo, "github.com") {
		return fmt.Sprintf("https://%s/compare/%s...%s", repo, u.from[:8], u.to[:8])
	}
	return ""
}

// Print writes a summary of added, removed and changed dependencies to stdout.
func (c Changes) Print() {
	if c.added != nil {
		fmt.Println("Added:")
		for _, i := range c.added {
			fmt.Printf(" - %s\n", i)
		}
	}
	if c.updates != nil {
		fmt.Println("Changed:")
		for _, i := range c.updates {
			fmt.Printf(" - %s: %s -> %s\n", i.name, i.from[:8], i.to[:8])
			if link := i.compareLink(); link != "" {
				fmt.Printf("   - %s\n", link)
			}
		}
	}
	if c.removed != nil {
		fmt.Println("Removed:")
		for _, d := range c.removed {
			fmt.Printf(" - %s\n", d)
		}
	}
}

func (c Changes) Links() {
	for _, i := range c.updates {
		if link := i.compareLink(); link != "" {
			fmt.Println(link)
		} else {
			fmt.Fprintf(os.Stderr, "cannot get link for %s (%s -> %s)\n", i.name, i.from[:8], i.to[:8])
		}
	}
}

// Compare computes a Changes describing the differences between before and after.
func Compare(before, after Versions) (Changes, error) {
	d := Changes{}
	for name, new := range after {
		old, ok := before[name]
		if !ok {
			d.added = append(d.added, name)
		} else if new != old {
			d.updates = append(d.updates, update{name: name, from: old, to: new})
		}
	}
	for name := range before {
		if _, ok := after[name]; !ok {
			d.removed = append(d.removed, name)
		}
	}
	sort.Strings(d.added)
	sort.Strings(d.removed)
	sort.Sort(d.updates)
	return d, nil
}
