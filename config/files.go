package config

import (
	"errors"
	"io/ioutil"
	"os"
	"strings"
)

// FindConfigFilesInDir finds all files with a .json extension in the supplied directory path, recursively checking
// sub-directories. Note that each directory's contents are examined and added to the list of files lexicographically,
// so any files in a sub-directory 'b' would appear in the resulting list of files before 'c.json'
func FindConfigFilesInDir(dirPath string) ([]string, error) {

	contents, err := ioutil.ReadDir(dirPath)

	files := make([]string, 0)

	if err != nil {
		return nil, err
	}

	for _, info := range contents {

		fileName := info.Name()

		if info.Mode().IsDir() {

			if sub, err := FindConfigFilesInDir(dirPath + "/" + fileName); err != nil {
				return nil, err
			} else {
				files = append(files, sub...)
			}

		} else if strings.HasSuffix(fileName, ".json") {

			files = append(files, dirPath+"/"+fileName)
		}

	}

	return files, nil
}

func FileListFromPath(path string) ([]string, error) {

	files := make([]string, 0)

	file, err := os.Open(path)

	if err != nil {
		err := errors.New("Unable to open file/dir " + path)
		return files, err
	}

	defer file.Close()

	fileInfo, err := file.Stat()

	if err != nil {
		err := errors.New("Unable to obtain file info for file/dir " + path)
		return files, err
	}

	if fileInfo.IsDir() {
		contents, err := ioutil.ReadDir(path)

		if err != nil {
			err := errors.New("Unable to read contents of directory " + path)
			return files, err
		}

		for _, info := range contents {
			fileName := info.Name()

			p := path + "/" + fileName

			if info.IsDir() {

				if sf, err := FileListFromPath(p); err != nil {
					return nil, err
				} else {
					files = append(files, sf...)
				}
			} else {
				files = append(files, p)
			}
		}

	} else {
		files = append(files, file.Name())
	}

	return files, nil
}
