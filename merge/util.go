package merge

import (
	"errors"
	"fmt"
	"golang.org/x/tools/cover"
	"io"
	"io/ioutil"
	"os"
)

// DumpProfile dumps the profile to the given file destination.
// If the destination is "-", it instead writes to stdout.
func DumpProfile(destination string, profile []*cover.Profile) error {
	var output io.Writer
	if destination == "-" {
		output = os.Stdout
	} else {
		f, err := os.Create(destination)
		if err != nil {
			return fmt.Errorf("failed to open %s: %v", destination, err)
		}
		defer f.Close()
		output = f
	}
	err := dumpProfile(profile, output)
	if err != nil {
		return fmt.Errorf("failed to dump profile: %v", err)
	}
	return nil
}

// LoadProfile loads a profile from the given filename.
// If the filename is "-", it instead reads from stdin.
func LoadProfile(origin string) ([]*cover.Profile, error) {
	filename := origin
	if origin == "-" {
		// Annoyingly, ParseProfiles only accepts a filename, so we have to write the bytes to disk
		// so it can read them back.
		// We could probably also just give it /dev/stdin, but that'll break on Windows.
		tf, err := ioutil.TempFile("", "")
		if err != nil {
			return nil, fmt.Errorf("failed to create temp file: %v", err)
		}
		defer tf.Close()
		defer os.Remove(tf.Name())
		if _, err := io.Copy(tf, os.Stdin); err != nil {
			return nil, fmt.Errorf("failed to copy stdin to temp file: %v", err)
		}
		filename = tf.Name()
	}
	return cover.ParseProfiles(filename)
}

// DumpProfile dumps the profiles given to writer in go coverage format.
func dumpProfile(profiles []*cover.Profile, writer io.Writer) error {
	if len(profiles) == 0 {
		return errors.New("can't write an empty profile")
	}
	if _, err := io.WriteString(writer, "mode: "+profiles[0].Mode+"\n"); err != nil {
		return err
	}
	for _, profile := range profiles {
		for _, block := range profile.Blocks {
			if _, err := fmt.Fprintf(writer, "%s:%d.%d,%d.%d %d %d\n", profile.FileName, block.StartLine, block.StartCol, block.EndLine, block.EndCol, block.NumStmt, block.Count); err != nil {
				return err
			}
		}
	}
	return nil
}

func deepCopyProfile(profile cover.Profile) cover.Profile {
	p := profile
	p.Blocks = make([]cover.ProfileBlock, len(profile.Blocks))
	copy(p.Blocks, profile.Blocks)
	return p
}

// blocksEqual returns true if the blocks refer to the same code, otherwise false.
// It does not care about Count.
func blocksEqual(a cover.ProfileBlock, b cover.ProfileBlock) bool {
	return a.StartCol == b.StartCol && a.StartLine == b.StartLine &&
		a.EndCol == b.EndCol && a.EndLine == b.EndLine && a.NumStmt == b.NumStmt
}

func ensureProfilesMatch(a *cover.Profile, b *cover.Profile) error {
	if a.FileName != b.FileName {
		return fmt.Errorf("coverage filename mismatch (%s vs %s)", a.FileName, b.FileName)
	}
	if len(a.Blocks) != len(b.Blocks) {
		return fmt.Errorf("file block count for %s mismatches (%d vs %d)", a.FileName, len(a.Blocks), len(b.Blocks))
	}
	if a.Mode != b.Mode {
		return fmt.Errorf("mode for %s mismatches (%s vs %s)", a.FileName, a.Mode, b.Mode)
	}
	for i, ba := range a.Blocks {
		bb := b.Blocks[i]
		if !blocksEqual(ba, bb) {
			return fmt.Errorf("coverage block mismatch: block #%d for %s (%+v mismatches %+v)", i, a.FileName, ba, bb)
		}
	}
	return nil
}
