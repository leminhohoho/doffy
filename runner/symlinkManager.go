package runner

import (
	"errors"
	"fmt"
	"os"
	"path"
)

type Symlink struct {
	originPath  string
	symlinkPath string
}

func (sl *Symlink) Link() error {
	return os.Symlink(sl.originPath, sl.symlinkPath)
}

func (sl *Symlink) Unlink() error {
	err := os.Remove(sl.symlinkPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	}

	return nil
}

func (sl *Symlink) IsConnected() bool {
	dest, err := os.Readlink(sl.symlinkPath)
	if err != nil {
		return false
	}

	if dest != sl.originPath {
		return false
	}

	return true
}

type Doffy struct {
	symlinks []*Symlink
	// TODO: Add parameters for the config in the future
	isDelete bool
}

// NOTE: Every path has to be converted to absolute before being processed
func NewDoffy(isDelete bool) *Doffy {
	return &Doffy{isDelete: isDelete}
}

func (d *Doffy) Create() error {
	// NOTE: Hard coded for testing, will switch this to read from config file later
	pathsToLink, err := GetValidPaths(
		os.Getenv("CURRENT_DIR"),
		[]string{path.Join(os.Getenv("CURRENT_DIR"), ".config")},
	)
	if err != nil {
		return err
	}

	for _, path := range pathsToLink {
		symlinkPath, err := PathResolver(path)

		if err != nil {
			if errors.As(err, &ErrSymlinkExist{}) || errors.As(err, &ErrFileExist{}) ||
				errors.As(err, &ErrDirExist{}) {
				fmt.Printf(err.Error())
				continue
			}
			return err
		}

		if err := os.Symlink(path, symlinkPath); err != nil {
			return err
		}
	}

	fmt.Println("Linked!")
	return nil
}

func (d *Doffy) Run() error {
	if d.isDelete {
		return d.Delete()
	}

	if err := d.Create(); err != nil {
		return err
	}

	return nil
}

func (d *Doffy) Delete() error {
	pathsToLink, err := GetValidPaths(
		os.Getenv("CURRENT_DIR"),
		[]string{path.Join(os.Getenv("CURRENT_DIR"), ".config")},
	)
	if err != nil {
		return err
	}

	for _, path := range pathsToLink {
		symlinkPath, err := CreateSymlinkPath(path)
		if err != nil {
			return err
		}

		err = os.Remove(symlinkPath)
		if err != nil {
			if !os.IsNotExist(err) {
				return err
			}
		}
	}

	fmt.Println("Deleted!")
	return nil
}
