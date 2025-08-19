package stacktrace

///
/// Stolen from palantir/stacktrace repo
/// -> https://github.com/palantir/stacktrace/blob/master/cleanpath/gopath.go
/// -> Apache 2.0 LICENSE
///

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// removeGoPath makes a path relative to one of the src directories in the $GOPATH
// environment variable. This function is used to clean up file paths in stack traces
// by removing the GOPATH prefix to make paths more readable and consistent.
//
// Parameters:
//   - path: The absolute file path to clean
//
// Returns the cleaned path relative to GOPATH, or the original path if it's not
// within GOPATH or if GOPATH is empty.
//
// The function:
//   - Splits GOPATH into individual directories
//   - Sorts directories by length (longest first) to find the best match
//   - Makes the path relative to the longest matching GOPATH/src directory
//   - Returns the original path if no match is found
//
// This function is used internally by the stacktrace package to provide
// cleaner, more readable file paths in debugging output.
//
// Example:
//
//	Input:  "/home/user/go/src/github.com/user/project/main.go"
//	GOPATH: "/home/user/go"
//	Output: "github.com/user/project/main.go"
func removeGoPath(path string) string {
	dirs := filepath.SplitList(os.Getenv("GOPATH"))
	// Sort in decreasing order by length so the longest matching prefix is removed
	sort.Stable(longestFirst(dirs))
	for _, dir := range dirs {
		srcdir := filepath.Join(dir, "src")
		rel, err := filepath.Rel(srcdir, path)
		// filepath.Rel can traverse parent directories, don't want those
		if err == nil && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return rel
		}
	}
	return path
}

// longestFirst is a custom sort interface for sorting strings by length in descending order.
// This type is used to sort GOPATH directories so that the longest matching prefix
// is processed first, ensuring the most specific GOPATH directory is used.
//
// The sort interface methods:
//   - Len: Returns the number of strings in the slice
//   - Less: Returns true if the string at index i is longer than the string at index j
//   - Swap: Swaps the strings at indices i and j
type longestFirst []string

func (strs longestFirst) Len() int           { return len(strs) }
func (strs longestFirst) Less(i, j int) bool { return len(strs[i]) > len(strs[j]) }
func (strs longestFirst) Swap(i, j int)      { strs[i], strs[j] = strs[j], strs[i] }
