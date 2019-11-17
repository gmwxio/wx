package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/wxio/wx/internal/dump"
	"github.com/wxio/wx/internal/genmd"
	"github.com/wxio/wx/internal/git"
	"github.com/wxio/wx/internal/initcli"
	"github.com/wxio/wx/internal/shell"
	"github.com/wxio/wx/internal/types"

	"github.com/jpillora/opts"
)

func main() {
	cwd, path, err := getConfig(".wx.yaml")
	if err != nil && os.Getenv("COMP_LINE") == "" {
		fmt.Fprintf(os.Stderr, "Error %v\n", err)
		os.Exit(1)
	}
	rcfg := &types.Root{}
	rflg := &types.Root{}
	r := &types.Root{
		WorkspaceRoot: path,
		CWD:           cwd,
		RootCfg:       rcfg,
	}
	cfg := filepath.Join(path, ".wx.yaml")
	op := opts.New(rflg).
		Name("wx").
		EmbedGlobalFlagSet().
		Complete().
		// Version(types.Version).
		AddCommand(opts.New(&types.VersionCmd{}).Name("version")).
		AddCommand(opts.New(initcli.New(r)).Name("init")).
		AddCommand(opts.New(dump.New(r)).Name("dump")).
		AddCommand(opts.New(git.New(r)).Name("git")).
		AddCommand(opts.New(shell.New(r)).Name("sh")).
		AddCommand(opts.New(genmd.New(r)).Name("gen-markdown")).
		FieldConfigPath(cfg, rcfg).
		Parse()
	r.Configure(rcfg, rflg)
	op.RunFatal()
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
