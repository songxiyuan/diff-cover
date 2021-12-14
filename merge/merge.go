package merge

import (
	"errors"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/songxiyuan/diff-cover/util"
	"golang.org/x/tools/cover"
	"sort"
	"strings"
)

type CommitCover struct {
	Repository    string
	Branch        string
	CommitId      string
	CoverFilePath string
}

// CoverPos record the position of cover file
type CoverPos struct {
	Line int
	Col  int
}

var (
	ErrRepositoryNotSame = errors.New("repository not same")
)

// SameCommitCoverMerge 相同commit的测试覆盖报告的合并,覆盖到第二个文件中
func SameCommitCoverMerge(filePath1, filePath2 string) error {
	//合并
	p1, err := cover.ParseProfiles(filePath1)
	if err != nil {
		return err
	}
	p2, err := cover.ParseProfiles(filePath2)
	if err != nil {
		return err
	}
	p12, err := SameCommitProfileMerge(p2, p1)
	if err != nil {
		return err
	}
	err = DumpProfile(filePath2, p12)
	if err != nil {
		return err
	}
	return nil
}

// SameCommitProfileMerge 相同commit的测试覆盖报告的合并,注意只会将未覆盖到的行数合并到新的
func SameCommitProfileMerge(newP []*cover.Profile, oldP []*cover.Profile) (res []*cover.Profile, err error) {
	var result []*cover.Profile
	files := make(map[string]*cover.Profile, len(newP))
	for _, profile := range newP {
		np := deepCopyProfile(*profile)
		result = append(result, &np)
		files[np.FileName] = &np
	}

	needsSort := false
	// Now merge oldP into the result
	for _, profile := range oldP {
		dest, ok := files[profile.FileName]
		if ok {
			if err := ensureProfilesMatch(profile, dest); err != nil {
				return nil, fmt.Errorf("error merging %s: %w", profile.FileName, err)
			}
			for i, block := range profile.Blocks {
				db := &dest.Blocks[i]
				if db.Count == 0 { //只会将未覆盖到的行数合并到新的
					db.Count += block.Count
				}
			}
		} else {
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

// DiffCoverMerge 不同commit的测试覆盖报告合并
func DiffCoverMerge(ccFrom, ccTo CommitCover, gitDir string) (profile []*cover.Profile, err error) {
	if ccFrom.Repository != ccTo.Repository {
		err = ErrRepositoryNotSame
		return
	}
	//todo 相同commit id合并问题
	if ccFrom.CommitId == ccTo.CommitId {
		err = errors.New("commit id equal")
		return
	}

	//path, err := util.GetGitPath(ccFrom.Repository)
	//if err != nil {
	//	util.Logger.Println(err)
	//	return
	//}

	r, err := git.PlainOpen(gitDir)
	if err != nil {
		util.Logger.Println(err)
		r, err = git.PlainClone(gitDir, false, &git.CloneOptions{
			URL: ccFrom.Repository,
		})
		if err != nil {
			util.Logger.Println(err)
			return
		}
	}

	commitFrom, err := object.GetCommit(r.Storer, plumbing.NewHash(ccFrom.CommitId))
	if err != nil {
		return
	}
	treeFrom, err := commitFrom.Tree()
	if err != nil {
		util.Logger.Println(err)
		return
	}
	//get module name
	moduleNameFrom := GetModuleFromTree(treeFrom)
	util.Logger.Println("module=", moduleNameFrom)

	commitTo, err := object.GetCommit(r.Storer, plumbing.NewHash(ccTo.CommitId))
	if err != nil {
		return
	}
	treeTo, err := commitTo.Tree()
	if err != nil {
		util.Logger.Println(err)
		return
	}

	profilesFrom, err := cover.ParseProfiles(ccFrom.CoverFilePath)
	if err != nil {
		util.Logger.Println(err)
		return
	}
	profilesTo, err := cover.ParseProfiles(ccTo.CoverFilePath)
	if err != nil {
		util.Logger.Println(err)
		return
	}
	err = DiffProfileMerge(moduleNameFrom, profilesFrom, profilesTo, treeFrom, treeTo)
	if err != nil {
		util.Logger.Println(err)
		return
	}
	return profilesTo, nil
}

func DiffProfileMerge(moduleNameFrom string, profilesFrom, profilesTo []*cover.Profile, treeFrom, treeTo *object.Tree) (err error) {
	changes, err := object.DiffTree(treeFrom, treeTo)
	var changeFileMap = make(map[string]struct{})
	for _, change := range changes {
		if change.From.Name != "" {
			changeFileMap[change.From.Name] = struct{}{}
		}
	}
	for _, profile := range profilesFrom {
		fileName := strings.TrimPrefix(profile.FileName, moduleNameFrom)
		_, ok := changeFileMap[fileName]
		if !ok {
			//不在差异文件里，将所有覆盖情况转移到新的覆盖情况中
			for i := 0; i < len(profilesTo); i++ {
				if profilesTo[i].FileName != profile.FileName {
					continue
				}
				for b := 0; b < len(profile.Blocks); b++ {
					//todo check order
					if profilesTo[i].Blocks[b].StartLine != profile.Blocks[b].StartLine {
						return errors.New("start line not equal")
					}
					profilesTo[i].Blocks[b].Count += profile.Blocks[b].Count
				}
				break
			}
			continue
		}
		//在差异文件中，diff两个commit文件，拿到旧覆盖行数对应新覆盖的新行数，然后遍历添加到coverage2
		file1Str, err := getFileString(treeFrom, fileName)
		if err != nil {
			fmt.Println(err)
			return err
		}
		file2Str, err := getFileString(treeTo, fileName)
		if err != nil {
			fmt.Println(err)
			return err
		}

		line2line := GetLineMap(file1Str, file2Str)
		//generate map of "ToCoverFile record line" to "ToCoverFile clock"
		startLineBlockMap := make(map[CoverPos]cover.ProfileBlock)
		for _, block := range profile.Blocks {
			if block.Count == 0 {
				continue
			}
			newStart, ok := line2line[block.StartLine]
			if !ok {
				continue
			}
			newEnd, ok := line2line[block.EndLine]
			if !ok {
				continue
			}
			block.StartLine = newStart
			block.EndLine = newEnd
			startLineBlockMap[CoverPos{newStart, block.StartCol}] = block
		}
		// range new file, replace block
		for i := 0; i < len(profilesTo); i++ {
			if profilesTo[i].FileName != profile.FileName {
				continue
			}
			for j := 0; j < len(profilesTo[i].Blocks); j++ {
				coverPos := CoverPos{
					Line: profilesTo[i].Blocks[j].StartLine,
					Col:  profilesTo[i].Blocks[j].StartCol,
				}
				block, ok := startLineBlockMap[coverPos]
				if !ok {
					continue
				}
				if block.EndLine != profilesTo[i].Blocks[j].EndLine {
					continue
					//return errors.New("get end line error")
				}
				profilesTo[i].Blocks[j].Count += block.Count
			}
			break
		}
	}
	return
}
func getFileString(tree *object.Tree, fileName string) (string, error) {
	file, err := tree.File(fileName)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	res, err := file.Contents()
	if err != nil {
		return "", err
	}
	return res, nil
}
