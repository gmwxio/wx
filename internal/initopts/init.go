package initopts

import (
	"fmt"

	"github.com/jpillora/opts"
)

type initOpts struct {
}

func New() opts.Opts {
	in := initOpts{	}
	return opts.New(&in).Name("init")
}

func (in *initOpts) Run() error {
	fmt.Printf("#init %+v\n", in)
	return nil
}

