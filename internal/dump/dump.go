package dump

import (
	"fmt"

	"gopkg.in/yaml.v2"

	"github.com/wxio/wx/internal/types"
)

type initOpts struct {
	rt *types.Root
}

// New constructor for initOpts
func New(rt *types.Root) interface{} {
	return &initOpts{rt: rt}
}

func (in *initOpts) Run() error {
	out, err := yaml.Marshal(in.rt)
	fmt.Printf("#dump \n%+v\n", string(out))
	return err
}
