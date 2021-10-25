package merge

import (
	"fmt"
	"golang.org/x/tools/cover"
	"sort"
)

// MergeProfiles merges two coverage profiles.
// The profiles are expected to be similar - that is, from multiple invocations of a
// single binary, or multiple binaries using the same codebase.
// In particular, any source files with the same path must have had identical content
// when building the binaries.
// MergeProfiles expects its arguments to be sorted: Profiles in alphabetical order,
// and lines in files in the order those lines appear. These are standard constraints for
// Go coverage profiles. The resulting profile will also obey these constraints.
func MergeProfiles(a []*cover.Profile, b []*cover.Profile) ([]*cover.Profile, error) {
	var result []*cover.Profile
	files := make(map[string]*cover.Profile, len(a))
	for _, profile := range a {
		np := deepCopyProfile(*profile)
		result = append(result, &np)
		files[np.FileName] = &np
	}

	needsSort := false
	// Now merge b into the result
	for _, profile := range b {
		dest, ok := files[profile.FileName]
		if ok {
			if err := ensureProfilesMatch(profile, dest); err != nil {
				return nil, fmt.Errorf("error merging %s: %v", profile.FileName, err)
			}
			for i, block := range profile.Blocks {
				db := &dest.Blocks[i]
				db.Count += block.Count
			}
		} else {
			// If we get some file we haven't seen before, we just append it.
			// We need to sort this later to ensure the resulting profile is still correctly sorted.
			np := deepCopyProfile(*profile)
			files[np.FileName] = &np
			result = append(result, &np)
			needsSort = true
		}
	}
	if needsSort {
		sort.Slice(result, func(i, j int) bool { return result[i].FileName < result[j].FileName })
	}
	return result, nil
}

func Merge(file1, file2 string) error {
	profiles1, err := cover.ParseProfiles(file1)
	if err != nil {
		return err
	}
	profiles2, err := cover.ParseProfiles(file2)
	if err != nil {
		return err
	}
	res, err := MergeProfiles(profiles1, profiles2)
	if err != nil {
		return err
	}
	err = DumpProfile("res.cov", res)
	if err != nil {
		return err
	}
	return nil
}

func MergeTest(file1, file2 string,line2line map[int]int) (cov []*cover.Profile, err error) {
	profiles1, err := cover.ParseProfiles(file1)
	if err != nil {
		return
	}
	profiles2, err := cover.ParseProfiles(file2)
	if err != nil {
		return
	}
	res, err := MergeProfiles(profiles1, profiles2)
	if err != nil {
		return
	}
	return res,nil
}
