package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jpillora/opts"

	"github.com/wxio/wx/internal/git"
	"github.com/wxio/wx/internal/initopts"
	"github.com/wxio/wx/internal/shell"
	"github.com/wxio/wx/internal/types"
)

func main() {
	cwd, path, err := getConfig(".wx.json")
	if err != nil && os.Getenv("COMP_LINE") == "" {
		fmt.Fprintf(os.Stderr, "Error %v\n", err)
		os.Exit(1)
	}
	r := &types.Root{
		WorkspaceRoot: path,
		CWD:           cwd,
	}
	cfg := filepath.Join(path, ".wx.json")
	opts.New(r).
		Name("wx").
		EmbedGlobalFlagSet().
		Complete().
		Version(types.Version).
		AddCommand(initopts.New()).
		AddCommand(git.New(r)).
		AddCommand(shell.New(r)).
		ConfigPath(cfg).
		Parse().
		RunFatal()
}

func getConfig(cfg string) (string, string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", "", fmt.Errorf("cant get cwd .. %v", err)
	}
	last := ""
	for {
		wd, err := os.Getwd()
		if err != nil {
			return cwd, "", fmt.Errorf("cant get cwd .. %v", err)
		}
		if _, err := os.Open(cfg); err != nil {
			if err = os.Chdir(".."); err != nil {
				return cwd, "", fmt.Errorf("cant open .. %v", err)
			}
			if last == wd {
				return cwd, "", fmt.Errorf("reached root without finding '%s' from %v", cfg, cwd)
			}
			last = wd
		} else {
			return cwd, wd, nil
		}
	}
}
