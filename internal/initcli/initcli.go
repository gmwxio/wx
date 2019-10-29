package initcli

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"text/template"

	"github.com/wxio/wx/internal/types"
)

type initOpts struct {
	rt        *types.Root
	Path      string
	Directory string `opts:"help=output directory"`
	Name      string `opts:"mode=arg"`
}

// New constructor for init
func New(rt *types.Root) interface{} {
	in := initOpts{
		rt:        rt,
		Directory: ".",
	}
	return &in
}

func (in *initOpts) Run() error {
	os.Chdir(in.rt.CWD)
	absOut, err := filepath.Abs(in.Directory)
	if err != nil {
		return fmt.Errorf(`%v`, absOut)
	}
	dir, err := os.OpenFile(in.Directory, os.O_APPEND, 0755)
	if err != nil {
		err = os.MkdirAll(in.Directory, 0755)
		if err != nil {
			return err
		}
	} else {
		names, err := dir.Readdirnames(1)
		if len(names) > 0 {
			return fmt.Errorf(`output directory not empty %v`, absOut)
		}
		if err != io.EOF {
			return err
		}
	}
	// if err := f.Close(); err != nil {
	// 	log.Fatal(err)
	// }

	modpath := in.Path
	if in.Path == "" {
		modpath = in.Name
	}
	data := struct {
		Name   string
		Path   string
		ModArg string
	}{
		Name:   in.Name,
		Path:   modpath,
		ModArg: "`opts:\"mode=arg\"`",
	}
	fmt.Printf("#init %+v\n", data)
	for _, fi := range files {
		tmpl, err := template.New(fi.Path).Parse(fi.Tmpl)
		if err != nil {
			fmt.Printf("tmpl parse error : %v\n", err)
			continue
		}
		fmt.Printf("#%v\n", fi.Path)
		pa := filepath.Join(in.Directory, path.Dir(fi.Path))
		_ = os.MkdirAll(pa, 0755)
		// if err != nil {
		// 	return err
		// }
		pa = filepath.Join(pa, path.Base(fi.Path))
		ofi, err := os.OpenFile(pa, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			fmt.Printf("new file error: %v", err)
			continue
		}
		err = tmpl.Execute(ofi, data)
		if err != nil {
			fmt.Printf("tmpl exec error : %v\n", err)
			continue
		}
	}
	return nil
}

type file struct {
	Path string
	Tmpl string
}

var files = []file{
	{
		Path: "go.mod",
		Tmpl: `module {{.Path}}

go 1.13

require github.com/jpillora/opts v1.0.0
`,
	},
	{
		Path: "main.go",
		Tmpl: `package main

import (
	"fmt"
	"os"

	"github.com/jpillora/opts"
)

type root struct {
	GlobalFlag string
}

type subCmd struct {
	Flag string
	Arg string {{.ModArg}}
}

func main() {
	r := root{}
	ro := opts.New(&r).Name("{{.Name}}").
		EmbedGlobalFlagSet().
		Complete().
		AddCommand(opts.New(&subCmd{}).Name("subcmd")).
		Parse()
	err := ro.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "run error %v\n", err)
		os.Exit(2)
	}
}

func (sc *subCmd) Run() {
	fmt.Printf("%v\n", sc)
}
`,
	},
}
