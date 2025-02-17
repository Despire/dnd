//go:build !windows

package restrictions

import (
	"bytes"
	"fmt"
	"os"

	"github.com/Despire/dnd/atomicfile"
)

const hostsFile = "/etc/hosts"

func (d *Diff) domainCommit() error {
	b, err := os.ReadFile(hostsFile)
	if err != nil {
		return fmt.Errorf("failed to open '/etc/hosts': %w", err)
	}

	for _, d := range d.Delete {
		target := d.(RDomain).String()
		i := bytes.Index(b, []byte(target))
		if i < 0 {
			continue // possible the file was changed...
		}
		b = append(b[:i], b[i+len(target):]...)
	}

	for _, d := range d.Missing {
		b = append(b, d.(RDomain).String()...)
	}

	if err := atomicfile.Write("/etc/hosts", b, 0644); err != nil {
		return fmt.Errorf("failed to atomically write to '/etc/hosts/: %w", err)
	}
	return nil
}
