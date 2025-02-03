package restrictions

import (
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
)

var home = ""

func init() {
	h, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	home = h
}

var (
	// ErrPartialSync is returned when not all items could be checked
	// for a 100% synchronization.
	ErrPartialSync = errors.New("partially synchronized state")

	// ErrPartialCommit is returned when not all items could be commited
	// at the OS level, meaning not all items from the configuration
	// will take effect.
	ErrPartialCommit = errors.New("partially commited state")
)

const (
	// DndPrefix used throughout the domain restrictions to identify which are maintained by the program.
	DndDomainPrefix = "#dnd"
	// DndApplication prefix identify which resources were created by the program to restrict the appliations.
	DndApplicationPrefix = "com.dnd."
)

// Type of a BlockedItem.
//
//go:generate stringer -type=Type
type Type uint8

const (
	Invalid Type = iota
	// Represents a domain name that should be blocked.
	Domain
	// Represents application installed on the system that should
	// not be able to run.
	Application
	TypeEnd
)

var TypeFromString = map[string]Type{
	"Domain":      Domain,
	"Application": Application,
}

type Diff struct {
	Type    Type
	Matched []any
	Missing []any
	Delete  []any
}

func (t Type) Diff(l List) (Diff, error) {
	var diff Diff
	var err error

	switch t {
	case Domain:
		var actual []RDomain
		if actual, err = SyncDomains(); err != nil {
			if !errors.Is(err, ErrPartialSync) {
				return Diff{}, fmt.Errorf("failed to synchronize actual state of domains: %w", err)
			}
		}
		diff = diffDomain(actual, l)
	case Application:
		var actual []RApplication
		if actual, err = SyncApplications(); err != nil {
			if !errors.Is(err, ErrPartialSync) {
				return Diff{}, fmt.Errorf("failed to synchronize actula state of application restrictions: %w", err)
			}
		}
		diff = diffApplication(actual, l)
	}

	return diff, err // can be partial error
}

func (d *Diff) Print(w io.Writer) {
	builder := strings.Builder{}
	if d.Type == Domain {
		builder.WriteString("Domains:\n")
		builder.WriteString(fmt.Sprintf("~ matched [%v]\n", len(d.Matched)))
		for _, m := range d.Matched {
			for _, r := range m.(RDomain).Restrictions {
				builder.WriteString(fmt.Sprintf("\tIP:%s\tDomains:%v\n", r.IP, r.Domains))
			}
		}
		builder.WriteString(fmt.Sprintf("+ add [%v]\n", len(d.Missing)))
		for _, m := range d.Missing {
			for _, r := range m.(RDomain).Restrictions {
				builder.WriteString(fmt.Sprintf("\tIP:%s\tDomains:%v\n", r.IP, r.Domains))
			}
		}
		builder.WriteString(fmt.Sprintf("- delete [%v]\n", len(d.Delete)))
		for _, m := range d.Delete {
			for _, r := range m.(RDomain).Restrictions {
				builder.WriteString(fmt.Sprintf("\tIP:%s\tDomains:%v\n", r.IP, r.Domains))
			}
		}
	}
	if d.Type == Application {
		builder.WriteString("Applications\n")
		builder.WriteString(fmt.Sprintf("~ matched [%v]\n", len(d.Matched)))
		for _, m := range d.Matched {
			builder.WriteString(fmt.Sprintf("\tPattern:%v\n", m.(RApplication).Pattern))
		}
		builder.WriteString(fmt.Sprintf("+ add [%v]\n", len(d.Missing)))
		for _, m := range d.Missing {
			builder.WriteString(fmt.Sprintf("\tPattern:%v\n", m.(RApplication).Pattern))
		}
		builder.WriteString(fmt.Sprintf("- delete [%v]\n", len(d.Delete)))
		for _, m := range d.Delete {
			builder.WriteString(fmt.Sprintf("\tPattern:%v\n", m.(RApplication).Pattern))
		}
	}
	fmt.Fprintln(w, builder.String())
}

func (d *Diff) Commit() error {
	if d.Type == Domain {
		if err := d.domainCommit(); err != nil {
			return err
		}
	}

	if d.Type == Application {
		if err := d.applicationCommit(); err != nil {
			return err
		}
	}
	return nil
}

func diffDomain(actual []RDomain, wanted List) Diff {
	diff := Diff{
		Type: Domain,
	}

	var wr []RDomain
	for _, r := range wanted.Items() {
		wr = append(wr, NewDomain(r))
		contains := slices.ContainsFunc(actual, func(dr RDomain) bool { return dr.Equal(wr[len(wr)-1]) })
		if contains {
			diff.Matched = append(diff.Matched, wr[len(wr)-1])
		} else {
			diff.Missing = append(diff.Missing, wr[len(wr)-1])
		}
	}

	for _, r := range actual {
		if !slices.ContainsFunc(wr, func(dr RDomain) bool { return dr.Equal(r) }) {
			diff.Delete = append(diff.Delete, r)
		}
	}

	return diff
}

func diffApplication(actual []RApplication, wanted List) Diff {
	diff := Diff{
		Type: Application,
	}

	var wr []RApplication
	for _, r := range wanted.Items() {
		wr = append(wr, NewApplication(r))
		if slices.Contains(actual, wr[len(wr)-1]) {
			diff.Matched = append(diff.Matched, wr[len(wr)-1])
		} else {
			diff.Missing = append(diff.Missing, wr[len(wr)-1])
		}
	}

	for _, r := range actual {
		if !slices.Contains(wr, r) {
			diff.Delete = append(diff.Delete, r)
		}
	}

	return diff
}
