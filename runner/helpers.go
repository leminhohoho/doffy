package runner

import (
	"os"
	"path"
	"slices"
)

func GetValidPaths(rootPath string, passList []string) ([]string, error) {
	ignoreList := []string{path.Join(os.Getenv("CURRENT_DIR"), ".git")}

	var validPaths []string
	dirPaths := []string{rootPath}

	for len(dirPaths) > 0 {
		newDirPaths := []string{}

		for _, dirPath := range dirPaths {
			entries, err := os.ReadDir(dirPath)
			if err != nil {
				return nil, err
			}

			for _, entry := range entries {
				entryAbsPath := path.Join(dirPath, entry.Name())

				// Check ignore list
				if slices.Contains(ignoreList, entryAbsPath) {
					continue
				}

				// Check if in pass list
				if slices.Contains(passList, entryAbsPath) {
					newDirPaths = append(newDirPaths, entryAbsPath)
					continue
				}

				validPaths = append(validPaths, entryAbsPath)
			}
		}

		dirPaths = newDirPaths
	}

	return validPaths, nil
}
