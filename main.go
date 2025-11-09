package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
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

	fmt.Printf("Dotfile dir: %s\n", dotfilesPath)
	fmt.Printf("Target dir: %s\n", targetDir)
}
