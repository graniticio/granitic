// Copyright 2016-2023 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// FindJSONFilesInDir finds all files with a .json extension in the supplied directory path, recursively checking
// sub-directories. Note that each directory's contents are examined and added to the list of files lexicographically,
// so any files in a sub-directory 'b' would appear in the resulting list of files before 'c.json'
func FindJSONFilesInDir(dirPath string) ([]string, error) {

	contents, err := ioutil.ReadDir(dirPath)

	files := make([]string, 0)

	if err != nil {
		return nil, err
	}

	for _, info := range contents {

		fileName := info.Name()

		if info.Mode().IsDir() {

			dp := filepath.Join(dirPath, fileName)

			if sub, err := FindJSONFilesInDir(dp); err == nil {
				files = append(files, sub...)
			} else {
				return nil, err
			}

		} else if filepath.Ext(fileName) == ".json" {

			f := filepath.Join(dirPath, fileName)

			files = append(files, f)
		}

	}

	return files, nil
}

// FileListFromPath takes a string that could represent a path to a directory
// or a path to a file and returns a list of file paths. If the path is to directory,
// any files in that directory are included in the result. Any sub-directories are recursively entered.
func FileListFromPath(path string) ([]string, error) {

	files := make([]string, 0)

	file, err := os.Open(path)

	if err != nil {
		return files, fmt.Errorf("unable to open file/dir %s", path)
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

			p := filepath.Join(path, fileName)

			if info.IsDir() {

				if sf, err := FileListFromPath(p); err == nil {
					files = append(files, sf...)
				} else {
					return nil, err
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

// WriteJSONConfig writes the contents of a config.Accessor to a file
func WriteJSONConfig(a *Accessor, f string) error {
	if content, err := json.MarshalIndent(a.JSONData, "", " "); err != nil {
		return err
	} else if err = ioutil.WriteFile(f, content, 0644); err != nil {
		return err
	}

	return nil

}
