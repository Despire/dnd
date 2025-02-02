package restrictions

import (
	"bytes"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
)

// RDomain is a domain restrictions
// that will block access to the specified
// domains, such as www.google.com etc.
type RDomain struct {
	header string
	footer string

	Restrictions []struct {
		IP      string
		Domains []string
		Raw     string
	}
}

func NewDomain(item string) RDomain {
	digest := sha512.Sum512([]byte(item))
	guard := fmt.Sprintf("%s%s\n", DndDomainPrefix, hex.EncodeToString(digest[:16]))

	r := RDomain{
		header: guard,
		footer: guard,
		Restrictions: []struct {
			IP      string
			Domains []string
			Raw     string
		}{
			{
				IP:      "127.0.0.1",
				Domains: []string{item},
				Raw:     fmt.Sprintf("127.0.0.1 %s\n", item),
			},
		},
	}

	return r
}

func (d RDomain) String() string {
	builder := strings.Builder{}

	builder.WriteString(d.header)
	for _, r := range d.Restrictions {
		builder.WriteString(r.Raw)
	}
	builder.WriteString(d.footer)

	return builder.String()
}

func (d RDomain) Equal(o RDomain) bool {
	if len(d.Restrictions) != len(o.Restrictions) {
		return false
	}
	for i, r := range d.Restrictions {
		if len(r.Domains) != len(o.Restrictions[i].Domains) {
			return false
		}
		if r.IP != o.Restrictions[i].IP {
			return false
		}
		for j, d := range r.Domains {
			if d != o.Restrictions[i].Domains[j] {
				return false
			}
		}
	}
	return d.header == o.header && d.footer == o.footer
}

func SyncDomains() ([]RDomain, error) {
	contents, err := os.ReadFile("/etc/hosts")
	if err != nil {
		return nil, err
	}

	var restrictions []RDomain

	lines := bytes.Split(contents, []byte{'\n'})
	for i := 0; i < len(lines); i++ {
		if !bytes.HasPrefix(lines[i], []byte(DndDomainPrefix)) {
			continue
		}
		start := i
		for {
			i++
			if i >= len(lines) {
				break
			}
			if bytes.HasPrefix(lines[i], []byte(DndDomainPrefix)) {
				break
			}
		}
		if i >= len(lines) {
			break
		}

		// NOTE: Include the trimmed new lines
		// Otherwise the match will not work.
		r := RDomain{
			header: string(lines[start]) + "\n",
			footer: string(lines[i]) + "\n",
		}

		for j := start + 1; j < i; j++ {
			d := strings.Fields(string(lines[j]))
			r.Restrictions = append(r.Restrictions, struct {
				IP      string
				Domains []string
				Raw     string
			}{
				IP:      d[0],
				Domains: d[1:],
				Raw:     string(lines[j]) + "\n",
			})
		}
		restrictions = append(restrictions, r)
	}
	return restrictions, nil
}
