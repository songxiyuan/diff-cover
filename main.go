package main

import (
	"diff-cover/merge"
	"fmt"
	"github.com/sergi/go-diff/diffmatchpatch"
	"golang.org/x/tools/cover"
	"io/ioutil"
	"strings"
	"time"
)

type diffCover struct {
}

//var cov2cov map[int]int //equal行数记录，旧文件行数->新文件行数

//得到diff中不变的旧行数到新行数的map
func getLine2Line(diffs []diffmatchpatch.Diff) map[int]int {
	oldLine := 0
	newLine := 0
	line2line := make(map[int]int)
	for _, diff := range diffs {
		if diff.Type == diffmatchpatch.DiffEqual {
			for _, _ = range diff.Text {
				oldLine++
				newLine++
				line2line[oldLine] = newLine
			}
		} else if diff.Type == diffmatchpatch.DiffDelete {
			for _, _ = range diff.Text { //因为是rune,所以用len()数量不对
				oldLine++
			}
		} else if diff.Type == diffmatchpatch.DiffInsert {
			for _, _ = range diff.Text {
				newLine++
			}
		}
	}
	return line2line
}

func GetMap(str1, str2 string) map[int]int {
	dmp := diffmatchpatch.New()
	dmp.DiffTimeout = 10 * time.Second
	wSrc, wDst, wArray := dmp.DiffLinesToRunes(str1, str2)
	//lineHash 每行string->对应wArray中第几行
	diffs := dmp.DiffMainRunes(wSrc, wDst, false)
	line2line := getLine2Line(diffs)
	//拿到equal对应的行数内容
	//遍历旧文件，记录旧文件
	//为新文件中跟旧文件equal的行数记录上

	//合并操作
	diffs = dmp.DiffCharsToLines(diffs, wArray)
	return line2line
}

func main() {
	main1, err := ioutil.ReadFile("./test/main1.go")
	if err != nil {
		return
	}
	main1Str := string(main1)
	main2, err := ioutil.ReadFile("./test/main2.go")
	if err != nil {
		return
	}
	main2str := string(main2)
	line2line := GetMap(main1Str, main2str)

	profiles1, err := cover.ParseProfiles("./test/cov1.out")
	if err != nil {
		return
	}
	profiles2, err := cover.ParseProfiles("./test/cov2.out")
	if err != nil {
		return
	}

	profile := profiles1[0]
	newStartEndMap := make(map[int]int)
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
			fmt.Println("diff 和 覆盖率文件对应错误")
			return
		}
		newStartEndMap[newStart] = newEnd
	}

	profileNew := profiles2[0]
	for i := 0; i < len(profileNew.Blocks); i++ {
		endLine, ok := newStartEndMap[profileNew.Blocks[i].StartLine]
		if !ok {
			continue
		}
		if endLine != profileNew.Blocks[i].EndLine {
			//报错
			fmt.Println("diff 和 覆盖率文件对应错误")
			return
		}
		profileNew.Blocks[i].Count = 1
	}
	err = merge.DumpProfile("covRes.cov", profiles2)
	if err != nil {
		fmt.Println("错误", err)
		return
	}
	return
}

//
//func main() {
//	fmt.Println("init")
//	// Clones the given repository in memory, creating the remote, the local
//	// branches and fetching the objects, exactly as:
//	examples.Info("git clone https://github.com/go-git/go-billy")
//
//	r, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
//		URL: "https://github.com/go-git/go-billy",
//	})
//	examples.CheckIfError(err)
//
//	// Gets the HEAD history from HEAD, just like this command:
//	examples.Info("git log")
//
//	// ... retrieves the branch pointed by HEAD
//	ref, err := r.Head()
//	examples.CheckIfError(err)
//
//	// ... retrieves the commit history
//	cIter, err := r.Log(&git.LogOptions{From: ref.Hash()})
//	examples.CheckIfError(err)
//
//	// ... just iterates over the commits, printing it
//	err = cIter.ForEach(func(c *object.Commit) error {
//		fmt.Println(c)
//		return nil
//	})
//	examples.CheckIfError(err)
//
//}

// DiffCharsToLines rehydrates the text in a diff from a string of line hashes to real lines of text.
func DiffCharsToLines(diffs []diffmatchpatch.Diff, lineArray []string) []diffmatchpatch.Diff {
	hydrated := make([]diffmatchpatch.Diff, 0, len(diffs))
	for _, aDiff := range diffs {
		chars := aDiff.Text
		text := make([]string, len(chars))
		count := 1
		for i, r := range chars {
			text[i] = lineArray[r]
			fmt.Println(count)
			count++
		}

		aDiff.Text = strings.Join(text, "")
		hydrated = append(hydrated, aDiff)
	}
	return hydrated
}

// DiffLinesToRunes splits two texts into a list of runes. Each rune represents one line.
func DiffLinesToRunes(text1, text2 string) ([]rune, []rune, []string, map[string]int) {
	// '\x00' is a valid character, but various debuggers don't like it. So we'll insert a junk entry to avoid generating a null character.
	lineArray := []string{""}    // e.g. lineArray[4] == 'Hello\n'
	lineHash := map[string]int{} // e.g. lineHash['Hello\n'] == 4

	chars1 := diffLinesToRunesMunge(text1, &lineArray, lineHash)
	chars2 := diffLinesToRunesMunge(text2, &lineArray, lineHash)

	return chars1, chars2, lineArray, lineHash
}

// diffLinesToRunesMunge splits a text into an array of strings, and reduces the texts to a []rune where each Unicode character represents one line.
// We use strings instead of []runes as input mainly because you can't use []rune as a map key.
func diffLinesToRunesMunge(text string, lineArray *[]string, lineHash map[string]int) []rune {
	// Walk the text, pulling out a substring for each line. text.split('\n') would would temporarily double our memory footprint. Modifying text would create many large strings to garbage collect.
	lineStart := 0
	lineEnd := -1
	runes := []rune{}

	for lineEnd < len(text)-1 {
		lineEnd = indexOf(text, "\n", lineStart)

		if lineEnd == -1 {
			lineEnd = len(text) - 1
		}

		line := text[lineStart : lineEnd+1]
		lineStart = lineEnd + 1
		lineValue, ok := lineHash[line]

		if ok {
			runes = append(runes, rune(lineValue))
		} else {
			*lineArray = append(*lineArray, line)
			lineHash[line] = len(*lineArray) - 1
			runes = append(runes, rune(len(*lineArray)-1))
		}
	}

	return runes
}

// indexOf returns the first index of pattern in str, starting at str[i].
func indexOf(str string, pattern string, i int) int {
	if i > len(str)-1 {
		return -1
	}
	if i <= 0 {
		return strings.Index(str, pattern)
	}
	ind := strings.Index(str[i:], pattern)
	if ind == -1 {
		return -1
	}
	return ind + i
}
