package merge

import (
	"diff-cover/util"
	"errors"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"golang.org/x/tools/cover"
	"strings"
)

type CommitCover struct {
	Repository    string
	Branch        string
	CommitId      string
	CoverFilePath string
}

var (
	ErrRepositoryNotSame = errors.New("repository not same")
)

func DiffCoverMerge(cc1, cc2 CommitCover, tempDir string) (profile []*cover.Profile, err error) {
	if cc1.Repository != cc2.Repository {
		err = ErrRepositoryNotSame
		return
	}
	//todo 相同commit id合并问题
	if cc1.CommitId == cc2.CommitId {
		err = errors.New("commit id equal")
		return
	}

	path, err := util.GetGitPath(cc1.Repository)
	if err != nil {
		util.Logger.Println(err)
		return
	}

	r, err := git.PlainOpen(tempDir + path)
	if err != nil {

		util.Logger.Println(err)
		r, err = git.PlainClone(tempDir+path, false, &git.CloneOptions{
			URL: cc1.Repository,
		})
		if err != nil {
			util.Logger.Println(err)
			return
		}
	}

	c1, err := object.GetCommit(r.Storer, plumbing.NewHash(cc1.CommitId))
	if err != nil {
		return
	}
	tree1, err := c1.Tree()
	if err != nil {
		util.Logger.Println(err)
		return
	}
	c2, err := object.GetCommit(r.Storer, plumbing.NewHash(cc2.CommitId))
	if err != nil {
		return
	}
	tree2, err := c2.Tree()
	if err != nil {
		util.Logger.Println(err)
		return
	}

	profiles1, err := cover.ParseProfiles(cc1.CoverFilePath)
	if err != nil {
		util.Logger.Println(err)
		return
	}
	profiles2, err := cover.ParseProfiles(cc2.CoverFilePath)
	if err != nil {
		util.Logger.Println(err)
		return
	}
	err = ProfileMerge(profiles1, profiles2, tree1, tree2)
	if err != nil {
		util.Logger.Println(err)
		return
	}
	return profiles2, nil
}

func ProfileMerge(profiles1, profiles2 []*cover.Profile, tree1, tree2 *object.Tree) (err error) {
	changes, err := object.DiffTree(tree1, tree2)
	var changeFileMap = make(map[string]struct{})
	for _, change := range changes {
		if change.From.Name != "" {
			changeFileMap[change.From.Name] = struct{}{}
		}
	}
	for _, profile := range profiles1 {
		fileName := strings.Join(strings.Split(profile.FileName, "/")[1:], "/")
		_, ok := changeFileMap[fileName]
		if !ok {
			//不在差异文件里，将所有覆盖情况转移到新的覆盖情况中
			for i := 0; i < len(profiles2); i++ {
				if profiles2[i].FileName != profile.FileName {
					continue
				}
				if len(profiles2[i].Blocks) != 1 {
					return errors.New("len(p.Blocks)!=1")
				}
				for b := 0; b < len(profile.Blocks); b++ {
					//todo 检查顺序是否一致
					if profiles2[i].Blocks[b].StartLine != profile.Blocks[b].StartLine {
						return errors.New("顺序不一致")
					}
					profiles2[i].Blocks[b].Count += profile.Blocks[b].Count
				}
				break
			}
			continue
		}
		//在差异文件中，diff两个commit文件，拿到旧覆盖行数对应新覆盖的新行数，然后遍历添加到coverage2
		file1Str, err := getFileString(tree1, fileName)
		if err != nil {
			fmt.Println(err)
			return err
		}
		file2Str, err := getFileString(tree2, fileName)
		if err != nil {
			fmt.Println(err)
			return err
		}

		line2line := GetLineMap(file1Str, file2Str)
		//生成 map[新覆盖文件开始行]新文件该行对应的block
		startLineBlockMap := make(map[int]cover.ProfileBlock)
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
				return errors.New("DiffCoverMerge 和 覆盖率文件对应错误")
			}
			block.StartLine = newStart
			block.EndLine = newEnd
			startLineBlockMap[newStart] = block
		}
		//遍历新覆盖文件,替换新的block
		for i := 0; i < len(profiles2); i++ {
			if profiles2[i].FileName != profile.FileName {
				continue
			}
			for j := 0; j < len(profiles2[i].Blocks); j++ {
				block, ok := startLineBlockMap[profiles2[i].Blocks[j].StartLine]
				if !ok {
					continue
				}
				if block.EndLine != profiles2[i].Blocks[j].EndLine {
					return errors.New("DiffCoverMerge 和 覆盖率文件对应错误")
				}
				profiles2[i].Blocks[j].Count += block.Count
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
