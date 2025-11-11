package runner

import "github.com/fatih/color"

type Result struct {
	OldPath string
	NewPath string
}

type Results []Result

func (rs Results) Log() {
	for _, result := range rs {
		color.New(color.FgHiWhite).Print("Symlink created: ")
		color.New(color.FgHiCyan, color.Bold).Printf("%s", result.NewPath)
		color.New(color.FgHiWhite).Print(" -> ")
		color.New(color.FgHiMagenta, color.Bold).Printf("%s\n", result.OldPath)
	}
}

func (rs Results) Summary() {
	if len(rs) > 1 {
		color.New(color.FgHiGreen).Printf("%d new symlinks created\n", len(rs))
	} else {
		color.New(color.FgHiGreen).Printf("%d new symlink created\n", len(rs))
	}
}
