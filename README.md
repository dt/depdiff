# Dependency Resolution Change Summary
A utility for summarizing and inspecting changes in Gopkg.lock/glide.lock files.

# Installation
To summmarize `dep` lockfile changes:
- `go get -u github.com/dt/depdiff/cmd/dep-diff-parser`

For the `glide` version:
- `go get -u github.com/dt/depdiff/cmd/glide-diff-parser`


# Example Usage
```
$ dep-diff-parser
Changes in Gopkg.lock since HEAD~
https://github.com/Sirupsen/logrus/compare/9b48ece7...3f603f49
https://github.com/golang/sys/compare/d75a5265...7a6e5648
https://github.com/golang/text/compare/44f4f658...dafb3384
https://github.com/golang/tools/compare/354f9f8b...19c96be7
```
Note, log messages are written to stderr, so the output is just white-space separated links, thus:
```
dep-diff-parser | xargs open # open all diffs in browser
```

# Usage and Features
`dep-diff-parser [-v] [from] [to]`

- `from` and `to` can be any git tree-ish, like `origin/master`, `HEAD~3`, `a3de42`
  - If `to` is unspecified, the default is to use the current `Gopkg.lock` content.
  - If `from` is not specified, the basis for comparision is `HEAD` if `Gopkg.lock` has uncommited
    modifications or `HEAD~` otherwise.
- Some libraries with non-github import paths (e.g. some `golang.org` paths) are mapped to their
  github mirrors to enable linking to change comparisions.
  - Pull Requests with additional mappings happily accepted -- the current list is just what I've
    happened to hit so far.
- `-v` enables more verbose summary of added, removed and changed dependencies.
