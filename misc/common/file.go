package common

import (
	"bufio"
	"errors"
	"fmt"
	"kgogame/util/logs"
	"os"
	"path/filepath"
	"strings"
)

// 判断文件夹是否存在
func DirectoryExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, err
		}
	} else {
		return true, nil
	}
}

// 判断文件是否存在
func FileExists(path string) (bool, error) {
	return DirectoryExists(path)
}

// 获取指定文件下所有文件，并制定需要的扩展文件
func GetDirectoryFile(path string, selectedExt []string) []string {
	list := doGetDirectoryFile(path, selectedExt)
	for i := 0; i < len(list); i++ {
		list[i] = strings.Replace(list[i], path, "", 1)
	}
	return list
}

func doGetDirectoryFile(path string, selectedExt []string) []string {
	l := make([]string, 0)
	files, err := os.ReadDir(path)
	if err != nil {
		return []string{}
	}
	for _, f := range files {
		if f.IsDir() {
			list := doGetDirectoryFile(filepath.Join(path, f.Name()), selectedExt)
			if len(list) > 0 {
				l = append(l, list...)
			}
		} else if len(selectedExt) == 0 || InStringArray(selectedExt, filepath.Ext(f.Name())) {
			l = append(l, filepath.Join(path, f.Name()))
		}
	}
	return l
}

type FileInfo struct {
	Name       string
	CreateTime int64
}

func GetDirectoryFileInfo(path string, selectedExt []string) []*FileInfo {
	list, err := doGetDirectoryFileInfo(path, selectedExt)
	if err != nil {
		logs.Info("GetDirectoryFileInfo error:%s", err)
		return []*FileInfo{}
	}
	for i := 0; i < len(list); i++ {
		list[i].Name = strings.Replace(list[i].Name, path, "", 1)
	}
	return list
}

func doGetDirectoryFileInfo(path string, selectedExt []string) ([]*FileInfo, error) {
	l := make([]*FileInfo, 0)
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, errors.New("failed visit path: " + path)
	}
	for _, f := range files {
		if f.IsDir() {
			list, _ := doGetDirectoryFileInfo(filepath.Join(path, f.Name()), selectedExt)
			for i, _ := range list {
				fileInfo := &FileInfo{
					Name:       list[i].Name,
					CreateTime: list[i].CreateTime,
				}
				l = append(l, fileInfo)
			}
		} else if len(selectedExt) == 0 || InStringArray(selectedExt, filepath.Ext(f.Name())) {
			info, fileErr := f.Info()
			if fileErr != nil {
				return nil, errors.New("failed get file info: " + f.Name())
			}
			fileInfo := &FileInfo{
				Name:       filepath.Join(path, f.Name()),
				CreateTime: info.ModTime().Unix(),
			}
			l = append(l, fileInfo)
		}
	}
	return l, nil
}

func CountLinesByFd(file *os.File) (int, error) {
	scanner := bufio.NewScanner(file)
	lineCount := 0
	for scanner.Scan() {
		lineCount++
	}
	if err := scanner.Err(); err != nil {
		return 0, err
	}
	_, err := file.Seek(0, 0)
	if err != nil {
		return 0, fmt.Errorf("error seeking to beginning of file err: %s", err)
	}
	return lineCount, nil
}
