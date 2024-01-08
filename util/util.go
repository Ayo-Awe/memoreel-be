package util

import (
	"errors"
	"runtime"
	"strings"
)

// Returns the project root directory
func GetProjectRoot() (string, error) {
	// get absolute file path for util.go e.g /home/aweayo/memoreel-be/util/util.go
	_, absolutFilePath, _, ok := runtime.Caller(0)
	if !ok {
		return "", errors.New("failed to get project root dir")
	}

	projectRoot := strings.TrimSuffix(absolutFilePath, "/util/util.go")
	return projectRoot, nil
}
