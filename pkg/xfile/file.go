package xfile

import (
	"os"
	"runtime"
	"strings"
)

/**
 *	文件工具
 * Copyright (C) @2020 hugo network Co. Ltd
 * @description
 * @updateRemark
 * @author               hugo
 * @updateUser
 * @createDate           2020/8/20 10:48 上午
 * @updateDate           2020/8/20 10:48 上午
 * @version              1.0
**/
func substr(s string, pos, length int) string {
	runes := []rune(s)
	l := pos + length
	if l > len(runes) {
		l = len(runes)
	}
	return string(runes[pos:l])
}

func getParentDirectory(dirctory string) string {
	if runtime.GOOS == "windows" {
		dirctory = strings.Replace(dirctory, "\\", "/", -1)
	}
	return substr(dirctory, 0, strings.LastIndex(dirctory, "/"))
}

// IsDirectory ...
func IsDirectory(path string) (bool, error) {
	f, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	switch mode := f.Mode(); {
	case mode.IsDir():
		return true, nil
	case mode.IsRegular():
		return false, nil
	}
	return false, nil
}

// CheckAndGetParentDir ... 检查并获取文件夹
func CheckAndGetParentDir(path string) string {
	// check path is the directory
	isDir, err := IsDirectory(path)
	if err != nil || isDir {
		return path
	}
	return getParentDirectory(path)
}

// MkdirIfNecessary ... 是否有必要创建文件夹
func MkdirIfNecessary(createDir string) error {
	var path string
	var err error
	//前边的判断是否是系统的分隔符
	if os.IsPathSeparator('\\') {
		path = "\\"
	} else {
		path = "/"
	}

	s := strings.Split(createDir, path)
	startIndex := 0
	dir := ""
	if s[0] == "" {
		startIndex = 1
	} else {
		dir, _ = os.Getwd() //当前的目录
	}
	for i := startIndex; i < len(s); i++ {
		d := dir + path + strings.Join(s[startIndex:i+1], path)
		if _, e := os.Stat(d); os.IsNotExist(e) {
			//在当前目录下生成md目录
			err = os.Mkdir(d, os.ModePerm)
			if err != nil {
				break
			}
		}
	}
	return err
}

//检查文件是否存在
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
