package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/leminhohoho/doffy/runner"
)

func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) == 0 {
		log.Fatal("Dotfiles path must not be empty\n")
	}

	dotfilesPath, err := filepath.Abs(args[0])

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	var targetDir string

	if len(args) < 2 {
		targetDir, err = filepath.Abs(homeDir)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		targetDir, err = filepath.Abs(args[1])
	}

	cfg, err := runner.NewConfig(dotfilesPath)
	if err != nil {
		log.Fatal(err)
	}

	results := runner.Results{}

	if err := runner.Link(dotfilesPath, targetDir, cfg, &results); err != nil {
		log.Fatal(err)
	}

	results.Log()
	results.Summary()
}
