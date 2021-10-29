package util

import (
	"errors"
	"strings"
)

//GetGitPath 得到git地址的路径,eg:
//https://github.com/songxiyuan/future.git --> songxiyuan/future
//git@github.com:songxiyuan/future.git --> songxiyuan/future
func GetGitPath(git string) (string, error) {
	if !strings.HasSuffix(git, ".git") {
		return "", errors.New("git format error, no suffix '.git'")
	}
	git = strings.TrimSuffix(git, ".git")
	if strings.HasPrefix(git, "http") {
		strs := strings.Split(git, "//")
		if len(strs) < 2 {
			return "", errors.New("git format error")
		}
		strs = strings.Split(strs[1], "/")
		if len(strs) < 2 {
			return "", errors.New("git format error")
		}
		return strings.Join(strs[1:], "/"), nil
	}
	if strings.HasPrefix(git, "git") {
		strs := strings.Split(git, ":")
		if len(strs) < 2 {
			return "", errors.New("git format error")
		}
		return strs[1], nil
	}
	return "", errors.New("git format error")
}
