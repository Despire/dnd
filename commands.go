package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"strconv"
	"strings"

	"github.com/Despire/dnd/restrictions"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var Usage = `dnd (short for do not disturb) is a program to block access to selected
applications running on the operating system and/or websites on the internet.

For the restrictions either added/deleted to take effect the ':commit' subcommond must be executed.

subcommands:
	:help
	:commit                 Commits the configured restrictions. Requires sudo.
	:add    <type> <args>   Adds a new restriction.
	:del    <type> <args>   Removes an existing restriction.
	:print                  Prints the configured restrictions.
	:types                  Prints all available types.

The program will store its configuration under $HOME/.dnd_config.
`

func help(w io.Writer) { fmt.Fprintln(w, Usage) }

func add(w io.Writer, r io.Reader, args ...string) {
	if len(args) < 1 {
		fmt.Fprintf(w, "no <type> specified\n")
		return
	}

	typ := strings.TrimSpace(args[0])
	typ = strings.ToLower(typ)
	typ = cases.Title(language.AmericanEnglish).String(typ)

	matched := restrictions.TypeFromString[typ]
	if matched == restrictions.Invalid {
		fmt.Fprintf(w, "invalid type %v\n", args[0])
		return
	}

	if len(args) < 2 {
		fmt.Fprintf(w, "no <args> specified\n")
		return
	}

	c, err := ReadConfig()
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			fmt.Fprintf(w, "failed to read config %s: %v\n", MustConfigPath(), err)
			return
		}
		c = &Config{Restrictions: make(map[restrictions.Type]restrictions.List)}
	}

	processed := 0

	for _, item := range restrictions.List(args[1]).Items() {
		item := strings.TrimSpace(item)
		if item == "" {
			continue
		}

		if matched == restrictions.Application {
			apps := FindApplicationBasedOnPattern(w, item)
			apps = append(apps, Match{program: item, dir: "general purporse pattern match"})
			fmt.Fprintf(w, "found %v matches based on provided pattern %q\n", len(apps), item)
			fmt.Fprintf(w, "which one do you want to proceed with:\n")
			for i, v := range apps {
				fmt.Fprintf(w, "[%v]\t%s | %s\n", i+1, v.program, v.dir)
			}
			fmt.Fprintf(w, "choose number between [%v, %v]: ", 1, len(apps))
			b := make([]byte, 3) // dont expect more than 999 matches...
			read, _ := bufio.NewReader(r).Read(b)
			if b[read-1] == '\n' {
				read--
			}

			selected, err := strconv.Atoi(string(b[:read]))
			if err != nil {
				fmt.Fprintf(w, "failed to parse input: %v, skipping...\n", err)
				continue
			}
			if !((selected-1) >= 0 && ((selected - 1) < len(apps))) {
				fmt.Fprintf(w, "invalid input: %v, skipping...\n", selected)
				continue
			}
			item = apps[selected-1].program
			fmt.Fprintf(w, "option %v will be used to kill any processes that contains the given pattern %q\n", selected, item)
		}

		c.Restrictions[matched] = c.Restrictions[matched].Append(item)
		processed += 1
	}

	if processed == 0 {
		return
	}

	if err := WriteConfig(c); err != nil {
		fmt.Fprintf(w, "failed to update config: %v\n", err)
		return
	}

	fmt.Fprintf(w, "processed %v items\n", processed)
}

func del(w io.Writer, args ...string) {
	if len(args) < 1 {
		fmt.Fprintf(w, "no <type> specified\n")
		return
	}

	typ := strings.TrimSpace(args[0])
	typ = strings.ToLower(typ)
	typ = cases.Title(language.AmericanEnglish).String(typ)

	if got := restrictions.TypeFromString[typ]; got == restrictions.Invalid || got == restrictions.TypeEnd {
		fmt.Fprintf(w, "invalid type %v", args[0])
		return
	}

	if len(args) < 2 {
		fmt.Fprintf(w, "no <args> specified\n")
		return
	}

	c, err := ReadConfig()
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			fmt.Fprintf(w, "failed to read config %s: %v\n", MustConfigPath(), err)
			return
		}
		return
	}

	processed := 0
	current, ok := c.Restrictions[restrictions.TypeFromString[typ]]
	if !ok {
		return
	}

	for {
		n, deleted := current.Remove(args[1])
		if !deleted {
			break
		}
		processed++
		current = n
	}

	if current.Empty() {
		delete(c.Restrictions, restrictions.TypeFromString[typ])
	} else {
		c.Restrictions[restrictions.TypeFromString[typ]] = current
	}

	if err := WriteConfig(c); err != nil {
		fmt.Fprintf(w, "failed to update config: %v\n", err)
		return
	}

	fmt.Fprintf(w, "processed %v items\n", processed)
}

func print(w io.Writer) {
	c, err := ReadConfig()
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			fmt.Fprintf(w, "failed to read config %s: %v\n", MustConfigPath(), err)
			return
		}
		fmt.Fprintln(w, "{}")
		return
	}

	b, err := json.Marshal(c)
	if err != nil {
		fmt.Fprintln(w, "{}")
		return
	}

	fmt.Fprintf(w, "%s\n", string(b))
}

func types(w io.Writer) {
	builder := new(strings.Builder)
	builder.WriteString(fmt.Sprintf("-%s: A single domain name or a list of domains separated with ',' [www.google.com,www.youtube.com]\n", restrictions.Type(1).String()))
	builder.WriteString(fmt.Sprintf("-%s: A single application name or a list of applications names separated with ',' [spotify, chrome]\n", restrictions.Type(2).String()))
	fmt.Fprintf(w, "%s", builder.String())
}

func commit(out io.Writer, in io.Reader) {
	c, err := ReadConfig()
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			fmt.Fprintf(out, "failed to read config %s: %v\n", MustConfigPath(), err)
			return
		}
		c = &Config{}
	}

	for t := restrictions.Type(1); t < restrictions.Type(restrictions.TypeEnd); t++ {
		diff, err := t.Diff(c.Restrictions[t])
		if err != nil {
			if !errors.Is(err, restrictions.ErrPartialSync) {
				fmt.Fprintf(out, "failed to determine difference between actual and desired state: %v", err)
				continue
			}
			fmt.Fprintf(out, "partially synced actuall state from the OS, continuing\nfailed operations: %v\n", err)
		}
		diff.Print(out)

		if len(diff.Delete) == 0 && len(diff.Missing) == 0 {
			continue
		}

		fmt.Fprintf(out, "commit ? (yes/no): ")
		b := make([]byte, 3)
		bufio.NewReader(in).Read(b)
		if string(b) != "yes" {
			fmt.Fprintf(out, "aborting...\n")
			continue
		}
		if err := diff.Commit(); err != nil {
			fmt.Fprintf(out, "failed to commit: %v, aborting...\n", err)
			continue
		}
	}

	// shallow clone, doesn't matter since we're dealing with strings.
	c.LastCommited = &Config{
		LastCommited: nil,
		Version:      c.Version,
		Restrictions: maps.Clone(c.Restrictions),
	}

	if err := WriteConfig(c); err != nil {
		fmt.Fprintf(out, "failed to update config: %v\n", err)
		return
	}
}
