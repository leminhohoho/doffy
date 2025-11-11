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
func alternatePathExist(pathA, pathB string) (bool, bool, error) {
	fA, err := os.Stat(pathA)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return false, false, err
		}

		return false, false, nil
	}

	fB, err := os.Stat(pathB)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return false, false, err
		}

		return false, false, nil
	}

	return fA.IsDir() == fB.IsDir(), fA.IsDir(), nil
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

		exists, isDir, err := alternatePathExist(pathOnDotfiles, pathOnTarget)
		if err != nil {
			return err
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
