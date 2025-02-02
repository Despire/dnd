package main

import (
	"io"
	"maps"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
)

type MatchSet = map[Match]struct{}

type Match struct {
	program string
	dir     string
}

// FindApplicationBasedOnPattern will look at the PATH environment
// and inside /Applications directory for a application that matches
// the pattern.
func FindApplicationBasedOnPattern(out io.Writer, pattern string) []Match {
	matches := make(MatchSet)

	// TODO: find a way to search more broadly for apps.
	if runtime.GOOS == "darwin" {
		apps, err := patternMatch("/Applications", pattern)
		if err != nil {
			apps = []Match{}
		}
		for _, app := range apps {
			matches[app] = struct{}{}
		}
	}

	if val, ok := os.LookupEnv("GOPATH"); ok {
		bins, err := patternMatch(filepath.Join(val, "bin"), pattern)
		if err != nil {
			bins = []Match{}
		}
		for _, app := range bins {
			matches[app] = struct{}{}
		}
	}

	if val, ok := os.LookupEnv("PATH"); ok {
		separator := ":"
		if runtime.GOOS == "windows" {
			separator = ";"
		}
		for _, path := range strings.Split(val, separator) {
			bins, err := patternMatch(path, pattern)
			if err != nil {
				bins = []Match{}
			}
			for _, app := range bins {
				matches[app] = struct{}{}
			}
		}
	}

	return slices.Collect(maps.Keys(matches))
}

func match(candidate, target string) bool {
	// TODO: a simple substring match does the job
	// if one enter the exact application name.
	// This can be improved by using an edit
	// distance function.
	return strings.Contains(strings.ToLower(candidate), strings.ToLower(target))
}

func patternMatch(dir, pattern string) ([]Match, error) {
	var matches []Match

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, e := range entries {
		if match(e.Name(), pattern) {
			matches = append(matches, Match{
				program: e.Name(),
				dir:     dir,
			})
		}
	}

	return matches, nil
}
