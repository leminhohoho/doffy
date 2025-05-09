package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/leminhohoho/doffy/runner"
)

var (
	currentDir string
	homeDir    string
)

func init() {
	args := os.Args[1:]

	if len(args) == 0 {
		log.Fatal("No path specified")
	}

	currentDir, err := filepath.Abs(args[0])
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
	doffy := runner.NewDoffy()

	if err := doffy.Run(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Linked!")
}
