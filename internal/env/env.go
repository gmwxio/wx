package env

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

type setTags struct {
	rt   *types.Root
	Tags []string `opts:"mode=arg"`
}

type addTagsToWS struct {
	rt        *types.Root
	Replace   bool      `help:"replace tags, default is append"`
	Workspace wspredict `opts:"mode=arg"`
}

type wspredict []string

func (wspredict) Complete(arg string) []string {
	cfg := filepath.Join(types.WorkspaceRoot, ".wx.yaml")
	// ioutil.R
	f, err := os.Open(cfg)
	if err != nil {
		return []string{}
	}
	rcfg := types.Root{}
	err = yaml.NewDecoder(f).Decode(&rcfg)
	if err != nil {
		return []string{}
	}
	ret := make([]string, len(rcfg.Workspaces), len(rcfg.Workspaces))
	for i, w := range rcfg.Workspaces {
		ret[i] = w.Name()
	}
	return ret
}

// New constructor for initOptsWorkspace
func New(rt *types.Root) interface{} {
	return &initOpts{rt: rt}
}

// NewSetTags constructor for initOpts
func NewSetTags(rt *types.Root) interface{} {
	return &setTags{rt: rt}
}

// NewAddTagsToWS constructor for initOpts
func NewAddTagsToWS(rt *types.Root) interface{} {
	return &addTagsToWS{
		rt: rt,
	}
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

func (in *setTags) Run() error {
	in.rt.RootCfg.Tags = in.Tags
	out, err := yaml.Marshal(in.rt.RootCfg)
	err = ioutil.WriteFile(filepath.Join(in.rt.WorkspaceRoot, ".wx.yaml"), out, os.ModePerm)
	return err
}

func (in *addTagsToWS) Run() error {
	m := map[string]bool{}
	for _, v := range in.Workspace {
		m[v] = true
	}
	for i := range in.rt.RootCfg.Workspaces {
		ws := in.rt.RootCfg.Workspaces[i]
		if m[ws.Name()] {
			if in.Replace {
				ws.Tags = in.rt.Tags
			} else {
				ws.Tags = append(ws.Tags, in.rt.Tags...)
			}
		}
		in.rt.RootCfg.Workspaces[i] = ws
	}
	out, err := yaml.Marshal(in.rt.RootCfg)
	err = ioutil.WriteFile(filepath.Join(in.rt.WorkspaceRoot, ".wx.yaml"), out, os.ModePerm)
	return err
}
