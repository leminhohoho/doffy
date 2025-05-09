package runner

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func CreateSymlinkPath(originPath string) (string, error) {
	absPath, err := filepath.Abs(originPath)
	if err != nil {
		return "", err
	}

	_, pathRelativeToCurrentDir, found := strings.Cut(absPath, os.Getenv("CURRENT_DIR"))
	if !found {
		return "", fmt.Errorf(
			"Error constructing symlink path\n(May be you use link outside of the current directory ?)",
		)
	}

	return path.Join(os.Getenv("HOME_DIR"), pathRelativeToCurrentDir), nil
}

func PathResolver(pathToBeResolved string) (string, error) {
	symlinkPath, err := CreateSymlinkPath(pathToBeResolved)
	if err != nil {
		return "", err
	}

	fileInfo, err := os.Lstat(symlinkPath)

	// NOTE: Currently handling file, folder and symlink, will handle all in the future
	if err == nil {
		if fileInfo.IsDir() {
			return "", ErrDirExist{symlinkPath}
		} else if fileInfo.Mode().IsRegular() {
			return "", ErrFileExist{symlinkPath}
		} else if fileInfo.Mode()&os.ModeSymlink != 0 {
			return "", ErrSymlinkExist{symlinkPath}
		}
	}

	if !os.IsNotExist(err) {
		return "", err
	}

	return symlinkPath, nil
}
