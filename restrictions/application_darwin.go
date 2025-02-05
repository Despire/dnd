//go:build darwin

package restrictions

import (
	"bytes"
	"crypto/sha512"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/Despire/dnd/atomicfile"
)

type RApplication struct {
	Pattern string

	metadata struct {
		label string
		file  string
	}
}

func NewApplication(item string) RApplication {
	app := RApplication{
		Pattern: item,
	}
	digest := sha512.Sum512([]byte(item))
	// 2^24 items needed for a collision, fair enough.
	app.metadata.label = fmt.Sprintf("%spkill%s", DndApplicationPrefix, hex.EncodeToString(digest[:6]))
	app.metadata.file = filepath.Join(home, "Library", "LaunchAgents", app.metadata.label+".plist")
	return app
}

func SyncApplications() ([]RApplication, error) {
	type Plist struct {
		Dict struct {
			Label   string
			Pattern string
		}
	}

	parentDir := filepath.Join(home, "Library", "LaunchAgents")
	entries, err := os.ReadDir(parentDir)
	if err != nil {
		return nil, err
	}

	var restrictions []RApplication
	var errSynchronized error

dir:
	for _, e := range entries {
		target := filepath.Join(parentDir, e.Name())
		b, err := os.ReadFile(target)
		if err != nil {
			errSynchronized = errors.Join(errSynchronized, fmt.Errorf("failed to read file %s: %w", target, err))
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
				errSynchronized = errors.Join(errSynchronized, fmt.Errorf("failed decoding file %s: %w", target, err))
				continue dir
			}

			switch current := tok.(type) {
			case xml.StartElement:
				switch current.Name.Local {
				case "key":
					if err := d.DecodeElement(&currentKey, &current); err != nil {
						errSynchronized = errors.Join(errSynchronized, fmt.Errorf("failed to decode key token %#v, file: %s: %w", current, target, err))
						continue dir
					}
				case "string":
					var keyValue string
					if err := d.DecodeElement(&keyValue, &current); err != nil {
						errSynchronized = errors.Join(errSynchronized, fmt.Errorf("failed to decode key token %#v, file: %s: %w", current, target, err))
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
			app := NewApplication(r.Dict.Pattern)
			app.metadata.file = target
			app.metadata.label = r.Dict.Label
			restrictions = append(restrictions, app)
		}
	}

	if errSynchronized != nil && len(restrictions) > 0 {
		errSynchronized = fmt.Errorf("%w: %w", ErrPartialSync, errSynchronized)
	}

	return restrictions, errSynchronized
}

func (d *Diff) applicationCommit() error {
	// if called with sudo, sudo_uid is the
	// id of the caller who called sudo.
	u := os.Getenv("SUDO_UID")
	if u == "" {
		user, _ := user.Current()
		u = user.Uid
	}

	// runes the pkill command on the given pattern every 30 secs.
	contentsTemplate := `
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>TargetedPattern</key>
    <string>%s</string>
    <key>Label</key>
    <string>%s</string>
    <key>ProgramArguments</key>
    <array>
		    <string>pkill</string>
		    <string>-SIGKILL</string>
		    <string>-i</string>
		    <string>-f</string>
		    <string>%s</string>
    </array>
    <key>StandardOutPath</key>
    <string>%s</string>
    <key>StandardErrorPath</key>
    <string>%s</string>
    <key>StartInterval</key>
    <integer>30</integer>
</dict>
</plist>
`

	var errCommited error
	var commited int

	for _, d := range d.Delete {
		app := d.(RApplication)

		err := os.Remove(app.metadata.file)
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				errCommited = errors.Join(errCommited, fmt.Errorf("failed to delete synced file %s: %w", app.metadata.file, err))
				continue
			}
		}

		commited++

		// now, on new logins the service will not run, but to make it exit immmediately
		//we need to remove it from launchctl.
		output := bytes.Buffer{}
		cmd := exec.Command("launchctl", "bootout", fmt.Sprintf("gui/%s/%s", u, app.metadata.label))
		cmd.Stdout = &output
		cmd.Stderr = &output
		if err := cmd.Run(); err != nil {
			errCommited = errors.Join(errCommited, fmt.Errorf("pkill removed for %s, but failed to disable it, a retry may be worth: %w: %s", app.metadata.label, err, output.String()))
			continue
		}
	}

	for _, d := range d.Missing {
		app := d.(RApplication)

		logs := filepath.Join(filepath.Dir(ConfigPath()), fmt.Sprintf("%spkill.log", DndApplicationPrefix))
		contents := fmt.Sprintf(contentsTemplate, app.Pattern, app.metadata.label, app.Pattern, logs, logs)

		if err := atomicfile.Write(app.metadata.file, []byte(contents), 0644); err != nil {
			errCommited = errors.Join(errCommited, fmt.Errorf("failed to write file %s: %w", app.metadata.file, err))
			continue
		}

		commited++

		// now, on new logins the service will run, but to make it run immmediately
		// we need to load it using launchctl.
		output := bytes.Buffer{}
		cmd := exec.Command("launchctl", "bootstrap", fmt.Sprintf("gui/%s", u), app.metadata.file)
		cmd.Stdout = &output
		cmd.Stderr = &output
		if err := cmd.Run(); err != nil {
			errCommited = errors.Join(errCommited, fmt.Errorf("pkill configured for %s, but failed to immediately launch it, a retry may be worth:%w: %s", app.metadata.file, err, output.String()))
			continue
		}
	}

	if errCommited != nil && commited > 0 {
		errCommited = fmt.Errorf("%w:%w", ErrPartialCommit, errCommited)
	}

	return errCommited
}
