package lockfile

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"syscall"
)

type Parser func([]byte) (Versions, error)

// ReadFromGit returns the version info as of a the given tree-ish.
func ReadFromGit(treeish string, filename string, handler Parser) (Versions, error) {
	out, err := exec.Command("git", "show", fmt.Sprintf("%s:%s", treeish, filename)).Output()
	if err != nil {
		return nil, err
	}
	return handler(out)
}

// ReadFromFs returns the version info in the current working directory.
func ReadFromFs(filename string, handler Parser) (Versions, error) {
	out, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return handler(out)
}

// LocalFileChanged returns if the file in the current directory has uncommited changes.
func LocalFileChanged(filename string) (bool, error) {
	err := exec.Command("git", "diff", "--quiet", filename).Run()
	if err == nil {
		return false, nil
	}
	if x, ok := err.(*exec.ExitError); ok && x.Sys().(syscall.WaitStatus).ExitStatus() == 1 {
		return true, nil
	}
	return false, err
}

func Usage(lockfileName string) func() {
	return func() {
		help := `
		usage: %s [from [to]]

		Parse and summarize changes to %[1]s.

		'from' and 'to' can be any git tree-ish, like 'master' or 'HEAD~3'

		If 'to' is not specified, the current file contents is used.

		If 'from' is not specified:
			'HEAD' is used if '%[1]s' has uncommited modifications, otherwise
			'HEAD~' is used (effectively comparing 'HEAD~' to 'HEAD').

		Flags:
		`
		fmt.Fprintf(os.Stderr, help, os.Args[0], lockfileName)
		flag.PrintDefaults()
	}
}

func SummarizeDiff(compare []string, verbose bool, lockfileName string, handler Parser) error {
	var from, to Versions
	var err error
	// The `after` in our comparison is always the current file unless we're given
	// an explicit treeish as a second argument.
	if len(compare) == 2 {
		fmt.Fprintf(os.Stderr, "Changes in %s between %s and %s\n", lockfileName, compare[0], compare[1])
		from, err = ReadFromGit(compare[0], lockfileName, handler)
		if err != nil {
			return err
		}
		to, err = ReadFromGit(compare[1], lockfileName, handler)
		if err != nil {
			return err
		}
	} else {
		to, err = ReadFromFs(lockfileName, handler)
		if err != nil {
			return err
		}

		since := "HEAD~"
		// If we're given a `from` argument, we read deps from that tree-ish.
		if len(compare) == 1 {
			since = compare[0]
		} else {
			// Otherwise, we compare to HEAD~ unless there are are uncommitted changes
			// in which case we just compare to HEAD.
			changed, err := LocalFileChanged(lockfileName)
			if err != nil {
				return err
			}
			if changed {
				since = "HEAD"
			}
		}

		fmt.Fprintf(os.Stderr, "Changes in %s since %s\n", lockfileName, since)
		from, err = ReadFromGit(since, lockfileName, handler)
		if err != nil {
			return err
		}
	}

	d, err := Compare(from, to)
	if err != nil {
		return err
	}

	if verbose {
		d.Print()
	} else {
		d.Links()
	}
	return nil
}
