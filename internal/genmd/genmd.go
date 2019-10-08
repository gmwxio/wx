package genmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/jpillora/md-tmpl/mdtmpl"
	"github.com/wxio/wx/internal/types"
)

type genmd struct {
	rt *types.Root
	//
	Filename   string `opts:"mode=arg"`
	WorkingDir string
	Preview    bool
}

// New constructor for subcommand
func New(rt *types.Root) interface{} {
	return &genmd{
		WorkingDir: ".",
		rt:         rt,
	}
}

func (gen *genmd) Run() error {
	fp := filepath.Join(gen.rt.CWD, gen.WorkingDir, gen.Filename)
	if b, err := ioutil.ReadFile(fp); err != nil {
		return err
	} else {
		if gen.Preview {
			for i, c := range mdtmpl.Commands(b) {
				fmt.Printf("%18s#%d %s\n", gen.Filename, i+1, c)
			}
			return nil
		}
		b = mdtmpl.ExecuteIn(b, filepath.Join(gen.rt.CWD, gen.WorkingDir))
		if err := ioutil.WriteFile(fp, b, 0655); err != nil {
			return err
		}
		log.Printf("executed templates and rewrote '%s'", gen.Filename)
		return nil
	}
}
