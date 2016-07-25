package config

import (
	"errors"
	"io/ioutil"
	"os"
	"strings"
)

func FindConfigFilesInDir(dirPath string) ([]string, error) {
	return findConfigFilesInDir(dirPath)
}

func findConfigFilesInDir(dirPath string) ([]string, error) {

	contents, err := ioutil.ReadDir(dirPath)

	files := make([]string, 0)

	if err != nil {
		return nil, err
	}

	for _, info := range contents {

		fileName := info.Name()

		if info.Mode().IsDir() {

		} else if strings.HasSuffix(fileName, ".json") {

			files = append(files, dirPath+"/"+fileName)
		}

	}

	return files, nil
}

func ExpandToFiles(paths []string) ([]string, error) {
	files := make([]string, 0)

	for _, path := range paths {

		expanded, err := FileListFromPath(path)

		if err != nil {
			return nil, err
		}

		files = append(files, expanded...)

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
			files = append(files, path+"/"+fileName)
		}

	} else {
		files = append(files, file.Name())
	}

	return files, nil
}
