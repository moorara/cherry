package service

import (
	"io/ioutil"
	"os"
)

func createTempFile(content string) (string, func(), error) {
	file, err := ioutil.TempFile("", "cherry-test-")
	if err != nil {
		return "", nil, err
	}

	if len(content) > 0 {
		// WriteFile will close the file as well
		err = ioutil.WriteFile(file.Name(), []byte(content), 0644)
		if err != nil {
			return "", nil, err
		}
	}

	filepath := file.Name()
	deleteFunc := func() {
		os.Remove(filepath)
	}

	return filepath, deleteFunc, nil
}
