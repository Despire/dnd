package main

import (
	"cmp"
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
	score   int
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

	result := slices.Collect(maps.Keys(matches))
	slices.SortStableFunc(result, func(left, right Match) int { return cmp.Compare(left.program, right.program) })
	slices.SortStableFunc(result, func(left, right Match) int { return cmp.Compare(left.score, right.score) })

	end := 15
	if len(result) < 15 {
		end = len(result)
	}
	return result[:end]
}

func match(candidate, target string) int {
	candidate = strings.ToLower(candidate)
	target = strings.ToLower(target)
	matches := 0

	grid := make([][]int, len(candidate)+1)
	for i := range grid {
		grid[i] = make([]int, len(target)+1)
	}

	for i := 1; i < len(grid); i++ {
		grid[i][0] = grid[i-1][0] + 1
	}
	for i := 1; i < len(grid[0]); i++ {
		grid[0][i] = grid[0][i-1] + 1
	}

	for i := 1; i < len(grid); i++ {
		for j := 1; j < len(grid[i]); j++ {
			if candidate[i-1] != target[j-1] {
				grid[i][j] = 1 + min(
					grid[i-1][j],
					grid[i][j-1],
					grid[i-1][j-1],
				)
			} else {
				if i == j {
					matches++ // extra points for guessing the right character at the right position.
				}
				grid[i][j] = grid[i-1][j-1]
			}
		}
	}

	// extra points for each character matched
	for _, c := range target {
		if i := strings.IndexRune(candidate, c); i >= 0 {
			matches++
		}
	}

	return grid[len(candidate)][len(target)] - 10*matches
}

func patternMatch(dir, pattern string) ([]Match, error) {
	var matches []Match

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, e := range entries {
		score := match(e.Name(), pattern)
		matches = append(matches, Match{
			program: e.Name(),
			dir:     dir,
			score:   score,
		})
	}

	return matches, nil
}
