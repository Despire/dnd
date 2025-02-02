//go:build darwin

package restrictions

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type RApplication struct {
	Pattern string
}

func NewApplication(item string) RApplication {
	return RApplication{
		Pattern: item,
	}
}

func SyncApplications() ([]RApplication, error) {
	type Plist struct {
		Dict struct {
			Label   string
			Pattern string
		}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	parentDir := filepath.Join(home, "Library", "LaunchAgents")
	entries, err := os.ReadDir(parentDir)
	if err != nil {
		return nil, err
	}

	var restrictions []RApplication
	var errAll error

dir:
	for _, e := range entries {
		target := filepath.Join(parentDir, e.Name())
		b, err := os.ReadFile(target)
		if err != nil {
			errAll = errors.Join(errAll, fmt.Errorf("failed to read file %s: %w", target, err))
			continue
		}

		var r Plist
		currentKey := ""
		d := xml.NewDecoder(bytes.NewReader(b))
		for {
			tok, err := d.Token()
			if err != nil {
				if err == io.EOF {
					break
				}
				errAll = errors.Join(errAll, fmt.Errorf("failed decoding file %s: %w", target, err))
				continue dir
			}

			switch current := tok.(type) {
			case xml.StartElement:
				switch current.Name.Local {
				case "key":
					if err := d.DecodeElement(&currentKey, &current); err != nil {
						errAll = errors.Join(errAll, fmt.Errorf("failed to decode key token %#v, file: %s: %w", current, target, err))
						continue dir
					}
				case "string":
					var keyValue string
					if err := d.DecodeElement(&keyValue, &current); err != nil {
						errAll = errors.Join(errAll, fmt.Errorf("failed to decode key token %#v, file: %s: %w", current, target, err))
						continue dir
					}
					switch currentKey {
					case "TargetedPattern":
						r.Dict.Pattern = keyValue
					case "Label":
						r.Dict.Label = keyValue
					}
				}
			}
		}

		if strings.HasPrefix(r.Dict.Label, DndApplicationPrefix) && r.Dict.Pattern != "" {
			restrictions = append(restrictions, RApplication{
				Pattern: r.Dict.Pattern,
			})
		}
	}

	if errAll != nil && len(restrictions) > 0 {
		errAll = fmt.Errorf("%w: %w", ErrPartialSync, errAll)
	}

	return restrictions, errAll
}
