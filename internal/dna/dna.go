package dna

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/wxio/wx/internal/types"
)

type godna struct {
	rt *types.Root
}

// New constructor for initOptsWorkspace
func New(rt *types.Root) interface{} {
	return &godna{rt: rt}
}

func (in *godna) Run() error {
	for _, ws := range in.rt.RootCfg.Workspaces {
		if ws.Dna != nil {
			path := filepath.Join(in.rt.WorkspaceRoot, ws.Path)
			abs, err := filepath.Abs(filepath.Join(path, ws.Dna.Output))
			if err != nil {
				return err
			}
			cmd := exec.Command(
				"docker",
				"run",
				"--rm",
				"-v", path+":/dna",
				"-v", abs+":/dna-dst",
				"-w", "/dna/",
				"wxio/godna:v1.14.0",
				"godna", "-d", "--logtostderr", "generate", "/dna-dst",
			)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Dir = filepath.Join(in.rt.WorkspaceRoot, ws.Path)
			err = cmd.Run()
			if err != nil {
				return err
			}
		}
	}
	return nil
}
