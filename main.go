package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/leminhohoho/doffy/runner"
)

func init() {
	currentDir, err := filepath.Abs(".")
	if err != nil {
		log.Fatal(err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	os.Setenv("CURRENT_DIR", currentDir)
	os.Setenv("HOME_DIR", homeDir)
}

func main() {
	args := os.Args[1:]

	for _, arg := range args {
		if err := runner.PathResolver(arg); err != nil {
			log.Fatal(err)
		}
	}
}
