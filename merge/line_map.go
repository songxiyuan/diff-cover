package merge

import (
	"bytes"
	"github.com/sergi/go-diff/diffmatchpatch"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/transform"
	"io/ioutil"
	"time"
)

// GetLine2Line 得到diff中不变的旧行数到新行数的map
func GetLine2Line(diffs []diffmatchpatch.Diff) map[int]int {
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

/*
	拿到equal对应的行数内容
	遍历旧文件，记录旧文件
	为新文件中跟旧文件equal的行数记录上
*/

func GetLineMap(str1, str2 string) map[int]int {
	str1, _ = toUtf8Encoding(str1)
	str2, _ = toUtf8Encoding(str2)
	dmp := diffmatchpatch.New()
	dmp.DiffTimeout = 10 * time.Second
	wSrc, wDst, wArray := dmp.DiffLinesToRunes(str1, str2)
	//lineHash 每行string->对应wArray中第几行
	diffs := dmp.DiffMainRunes(wSrc, wDst, false)
	line2line := GetLine2Line(diffs)

	//todo 文本合并操作
	diffs = dmp.DiffCharsToLines(diffs, wArray)
	return line2line
}

func toUtf8Encoding(str string) (res string, err error) {
	encode, _, _ := charset.DetermineEncoding([]byte(str), "")
	reader := transform.NewReader(bytes.NewReader([]byte(str)), encode.NewDecoder())
	resByte, err := ioutil.ReadAll(reader)
	if err != nil {
		return str, err
	}
	return string(resByte), err
}
