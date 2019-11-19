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
	Tags tagspredict `opts:"mode=arg"`
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
	tags := listTagsFunc(in.rt.RootCfg.Workspaces)
	fmt.Printf("active : %v\n", in.rt.Tags)
	fmt.Printf("all    : %v\n", tags)
	return nil
}

func listTagsFunc(workspaces []types.Workspace) []string {
	m := map[string]struct{}{}
	for _, ws := range  workspaces {
		for _, t := range ws.Tags {
			m[t] = struct{}{}
		}
	}
	tags := make([]string, 0, len(m))
	for k := range m {
			tags = append(tags, k)
	}
	sort.Sort(sort.StringSlice(tags))
	return tags
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
	rcfg, err := loadCfg()
	if err != nil {
		return []string{}
	}
	ret := make([]string, len(rcfg.Workspaces), len(rcfg.Workspaces))
	for i, w := range rcfg.Workspaces {
		ret[i] = w.Name()
	}
	return ret
}

type tagspredict []string

func (tagspredict) Complete(arg string) []string {
	rcfg, err := loadCfg()
	if err != nil {
		return []string{}
	}
	tags := listTagsFunc(rcfg.Workspaces)
	return tags
}

func loadCfg() (*types.Root, error) {
	cfg := filepath.Join(types.WorkspaceRoot, ".wx.yaml")
	// ioutil.R
	f, err := os.Open(cfg)
	if err != nil {
		return nil, err
	}
	rcfg := types.Root{}
	err = yaml.NewDecoder(f).Decode(&rcfg)
	if err != nil {
		return nil, err
	}
	return &rcfg, nil
}