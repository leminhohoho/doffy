package runner

import (
	"errors"
	"fmt"
	ignore "github.com/sabhiram/go-gitignore"
	"os"
	"path"
)

// alternatePathExist check if the 2 paths exists and valid for linking.
// It return 2 boolean values.
// 1st one indicate whether the 2 paths exists and are the same.
// the 2nd one indicates whether they are directories or files.
func alternatePathExist(oldPath, newPath string) (bool, bool, bool, error) {
	fOld, err := os.Lstat(oldPath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return false, false, false, err
		}

		return false, false, false, nil
	}

	fNew, err := os.Lstat(newPath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return false, false, false, err
		}

		return false, false, false, nil
	}

	isSameType := fOld.IsDir() == fNew.IsDir()
	isSymlink := fNew.Mode()&os.ModeSymlink != 0

	var isSymlinkToOldPath bool

	if isSymlink {
		symlinkPath, err := os.Readlink(newPath)
		if err != nil {
			return false, false, false, err
		}

		isSymlinkToOldPath = symlinkPath == oldPath
	}

	return isSymlink && isSymlinkToOldPath, isSameType, fNew.IsDir(), nil
}

func Link(dotfileDir, targetDir string) error {
	cfg, err := NewConfig()
	if err != nil {
		return err
	}

	ig := ignore.CompileIgnoreLines(cfg.Files.Exclude...)

	entries, err := os.ReadDir(dotfileDir)
	if err != nil {
		return err
	}

	if len(entries) == 0 {
		fmt.Printf("Directory at %s contain no entry, skipping\n", dotfileDir)
	}

	for _, entry := range entries {
		pathOnTarget := path.Join(targetDir, entry.Name())
		pathOnDotfiles := path.Join(dotfileDir, entry.Name())

		linked, exists, isDir, err := alternatePathExist(pathOnDotfiles, pathOnTarget)
		if err != nil {
			return err
		}

		if linked {
			continue
		}

		if exists {
			if !isDir {
				fmt.Printf("file on %s is already exist\n", pathOnTarget)
				continue
			}

			if err := Link(pathOnDotfiles, pathOnTarget); err != nil {
				return err
			}
		} else {
			if ig.MatchesPath(pathOnDotfiles) {
				fmt.Printf("Path %s match exclude list, skipping\n", pathOnDotfiles)
				continue
			}

			if err := os.Symlink(pathOnDotfiles, pathOnTarget); err != nil {
				return err
			}

			fmt.Printf("Symlink created: %s -> %s\n", pathOnTarget, pathOnDotfiles)
		}
	}

	return nil
}
