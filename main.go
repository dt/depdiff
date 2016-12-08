package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/dt/glide-diff-parser/diff"
	"github.com/dt/glide-diff-parser/glide"
)

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
	verbose := flag.Bool("v", false, "print a detailed summary of added, removed and changed dependencies")
	flag.Usage = func() {
		help := `
     usage: %s [from [to]]

     Parse and summarize changes to glide.lock.

     'from' and 'to' can be any git tree-ish, like 'master' or 'HEAD~3'

      If 'to' is not specified, the current file contents is used.

      If 'from' is not specified:
        'HEAD' is used if 'glide.lock' has uncommited modifications, otherwise
        'HEAD~' is used (effectively comparing 'HEAD~' to 'HEAD').

      Flags:
     `
		fmt.Fprintf(os.Stderr, help, os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	checkErr := func(e error) {
		if e != nil {
			fmt.Fprintf(os.Stderr, "%+v", e)
			os.Exit(1)
		}
	}

	args := flag.Args()

	var from, to diff.Versions
	var err error

	if len(args) > 2 {
		flag.Usage()
		os.Exit(-1)
	}

	file := "glide.lock"
	// The `after` in our comparison is always the current file unless we're given
	// an explicit treeish as a second argument.
	if len(args) == 2 {
		fmt.Fprintf(os.Stderr, "Changes in %s between %s and %s\n", file, args[0], args[1])
		from, err = glide.ReadFromGit(args[0])
		checkErr(err)
		to, err = glide.ReadFromGit(args[1])
		checkErr(err)
	} else {
		to, err = glide.ReadFromFs()
		checkErr(err)

		since := "HEAD~"
		// If we're given a `from` argument, we read deps from that tree-ish.
		if len(args) == 1 {
			since = args[0]
		} else {
			// Otherwise, we compare to HEAD~ unless there are are uncommitted changes
			// in which case we just compare to HEAD.
			changed, err := glide.LocalFileChanged()
			checkErr(err)
			if changed {
				since = "HEAD"
			}
		}

		fmt.Fprintf(os.Stderr, "Changes in %s since %s\n", file, since)
		from, err = glide.ReadFromGit(since)
		checkErr(err)
	}

	d, err := diff.Compare(from, to)
	checkErr(err)

	if *verbose {
		d.Print()
	} else {
		d.Links()
	}

}
