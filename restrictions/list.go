package restrictions

import (
	"slices"
	"strings"
)

// List is a comma separated list of values
// that describe addresses that must restricted
type List string

func (l List) Empty() bool { return l == "" }

// Returns the items of the list.
func (l List) Items() []string {
	var out []string
	for _, s := range strings.Split(string(l), ",") {
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

// Appends the new element to the list.
func (l List) Append(v string) List {
	return List(strings.Trim(strings.Join([]string{string(l), v}, ","), ","))
}

// Remove deletes the first occurence of v in the list.
func (l List) Remove(v string) (List, bool) {
	deleted := false
	items := strings.Split(string(l), ",")
	i := slices.Index(items, v)
	if i >= 0 {
		deleted = true
		items = append(items[:i], items[i+1:]...)
	}
	return List(strings.Join(items, ",")), deleted
}
