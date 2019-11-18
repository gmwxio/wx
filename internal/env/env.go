package env

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

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

type listTags struct {
	rt *types.Root
}

type addTagsToWS struct {
	rt        *types.Root
	Replace   bool      `help:"replace tags, default is append"`
	Workspace wspredict `opts:"mode=arg"`
}

// New constructor for initOptsWorkspace
func New(rt *types.Root) interface{} {
	return &initOpts{rt: rt}
}

// NewSetTags constructor for initOpts
func NewSetTags(rt *types.Root) interface{} {
	return &setTags{rt: rt}
}

// NewListTags constructor for initOpts
func NewListTags(rt *types.Root) interface{} {
	return &listTags{rt: rt}
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

func (in *listTags) Run() error {
	m := map[string]struct{}{}
	mt := map[string]bool{}
	for _, ws := range in.rt.RootCfg.Workspaces {
		for _, t := range ws.Tags {
			m[t] = struct{}{}
		}
	}
	tags := make([]string, 0, len(m))
	for _, t := range in.rt.Tags {
		mt[t] = true
	}
	for k := range m {
		if mt[k] {
			tags = append(tags, k+"*")
		} else {
			tags = append(tags, k)
		}
	}
	sort.Sort(sort.StringSlice(tags))
	fmt.Printf("tags : %v\n", tags)
	return nil
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
