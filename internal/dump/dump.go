package dump

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"

	"github.com/wxio/wx/internal/types"
)

type initOpts struct {
	rt    *types.Root
	Write bool `help:"Write the config back to the file."`
}

// New constructor for initOpts
func New(rt *types.Root) interface{} {
	return &initOpts{rt: rt}
}

func (in *initOpts) Run() error {
	if in.Write {
		out, err := yaml.Marshal(in.rt.RootCfg)
		err = ioutil.WriteFile(filepath.Join(in.rt.WorkspaceRoot, ".wx.yaml"), out, os.ModePerm)
		return err
	}
	out, err := yaml.Marshal(in.rt)
	fmt.Printf("#dump cfg : %v/.wx.yaml\n%+v\n", in.rt.WorkspaceRoot, string(out))
	return err
}
