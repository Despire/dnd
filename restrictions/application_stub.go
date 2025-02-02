//go:build !darwin

package restrictions

type RApplication struct {
	Pattern string
}

func NewApplication(item string) RApplication {
	return RApplication{
		Pattern: item,
	}
}

func SyncApplications() ([]RApplication, error) {
	panic("not implemented")
}
