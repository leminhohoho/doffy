package runner

import "fmt"

type ErrFileExist struct {
	path string
}

func (e ErrFileExist) Error() string {
	return fmt.Sprintf("File at %s already exists\n", e.path)
}

type ErrDirExist struct {
	path string
}

func (e ErrDirExist) Error() string {
	return fmt.Sprintf("Directory at %s already exists\n", e.path)
}

type ErrSymlinkExist struct {
	path string
}

func (e ErrSymlinkExist) Error() string {
	return fmt.Sprintf("Symlink at %s already exists\n", e.path)
}
