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
}

// NOTE: Every path has to be converted to absolute before being processed
func NewDoffy() *Doffy {
	return &Doffy{}
}

func (d *Doffy) init() error {
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
			if errors.Is(err, ErrSymlinkExist{}) || errors.Is(err, ErrFileExist{}) ||
				errors.Is(err, ErrDirExist{}) {
				fmt.Println(err.Error())
				continue
			}
			return err
		}

		d.symlinks = append(d.symlinks, &Symlink{path, symlinkPath})
	}

	return nil
}

func (d *Doffy) Run() error {
	if err := d.init(); err != nil {
		return err
	}

	for _, symlink := range d.symlinks {
		fmt.Println(*symlink)
		if err := symlink.Link(); err != nil {
			return err
		}
	}

	return nil
}
