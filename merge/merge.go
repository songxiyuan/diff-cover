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
	CommitId      string
	CoverFilePath string
}

var (
	ErrRepositoryNotSame = errors.New("repository not same")
)

func DiffCoverMerge(cc1, cc2 CommitCover, tempDir string) (err error) {
	if cc1.Repository != cc2.Repository {
		return ErrRepositoryNotSame
	}
	if cc1.CommitId == cc2.CommitId {
		return
	}

	path, err := util.GetGitPath(cc1.Repository)
	if err != nil {
		return
	}

	r, err := git.PlainOpen(tempDir + path)
	if err != nil {
		fmt.Println(err)
		r, err = git.PlainClone(tempDir+path, false, &git.CloneOptions{
			URL: cc1.Repository,
		})
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	//待确定fetch的使用
	err = r.Fetch(&git.FetchOptions{
		RemoteName: "origin",
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	sto := r.Storer

	var tree1 *object.Tree
	c1, err := object.GetCommit(sto, plumbing.NewHash(cc1.CommitId))
	if err != nil {
		return
	}
	tree1, err = c1.Tree()
	if err != nil {
		return
	}

	var tree2 *object.Tree
	c2, err := object.GetCommit(sto, plumbing.NewHash(cc2.CommitId))
	if err != nil {
		return
	}
	tree2, err = c2.Tree()
	if err != nil {
		return
	}
	changes, err := object.DiffTree(tree1, tree2)
	fmt.Println(changes)

	var changeFileMap = make(map[string]struct{})

	for _, change := range changes {
		if change.From.Name != "" {
			changeFileMap[change.From.Name] = struct{}{}
		}
	}

	profiles1, err := cover.ParseProfiles(cc1.CoverFilePath)
	if err != nil {
		return
	}
	profiles2, err := cover.ParseProfiles(cc2.CoverFilePath)
	if err != nil {
		return
	}

	for _, profile := range profiles1 {
		//去掉模块名字的前缀
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
					//todo 检查顺序是否一直
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
		file1, err := tree1.File(fileName)
		if err != nil {
			fmt.Println(err)
			return err
		}
		file2, err := tree2.File(fileName)
		if err != nil {
			fmt.Println(err)
			return err
		}
		file1Str, err := file1.Contents()
		if err != nil {
			fmt.Println(err)
			return err
		}
		file2Str, err := file2.Contents()
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
				//报错
				fmt.Println("DiffCoverMerge 和 覆盖率文件对应错误")
				return errors.New("")
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
					//报错
					fmt.Println("DiffCoverMerge 和 覆盖率文件对应错误")
					return errors.New("")
				}
				profiles2[i].Blocks[j].Count += block.Count
			}
			break
		}
	}
	err = DumpProfile("res.cov", profiles2)
	if err != nil {
		fmt.Println("错误", err)
		return
	}

	return nil

}
