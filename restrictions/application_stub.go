//go:build !darwin

package restrictions

import "errors"

type RApplication struct {
	Pattern string
}

func NewApplication(item string) RApplication {
	return RApplication{
		Pattern: item,
	}
}

func SyncApplications() ([]RApplication, error) {
	return nil, errors.New("not implemented")
}

func (d *Diff) applicationCommit() error {
	return errors.New("not implemented")
}
