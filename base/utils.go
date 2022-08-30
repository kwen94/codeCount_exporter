package base

import (
	"errors"
	"os"
)

func InStringSlice(haystack []string, needle string) int {
	for index, e := range haystack {
		if e == needle {
			return index
		}
	}
	return -2
}

func Makedir(dirPath string) (err error) {
	stat, err := os.Stat(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			// 路径不存在,创建
			err = os.Mkdir(dirPath, 0755)
			if err != nil {
				return err
			}
			return nil
		} else {
			return err
		}
	}

	if stat.IsDir() {
		return nil
	} else {
		return errors.New("给定的目录路径已被文件占用")
	}
}
