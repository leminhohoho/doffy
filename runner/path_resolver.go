package runner

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func PathResolver(pathToBeResolved string) error {
	absPath, err := filepath.Abs(pathToBeResolved)
	if err != nil {
		return err
	}

	_, pathRelativeToCurrentDir, found := strings.Cut(absPath, os.Getenv("CURRENT_DIR"))
	if !found {
		return fmt.Errorf(
			"Error constructing symlink path\n(May be you use link outside of the current directory ?)",
		)
	}

	pathToCompare := path.Join(os.Getenv("HOME_DIR"), pathRelativeToCurrentDir)

	fileInfo, err := os.Lstat(pathToCompare)

	// NOTE: Currently handling file, folder and symlink, will handle all in the future
	if err == nil {
		if fileInfo.IsDir() {
			return ErrDirExist{pathToCompare}
		} else if fileInfo.Mode().IsRegular() {
			return ErrFileExist{pathToCompare}
		} else if fileInfo.Mode() == fs.ModeSymlink {
			return ErrSymlinkExist{pathToCompare}
		}
	}

	if !os.IsNotExist(err) {
		return err
	}

	return nil
}
