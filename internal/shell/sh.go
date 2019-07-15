package shell

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/wxio/wx/internal/types"

	"github.com/jpillora/opts"
)

type shell struct {
	Args []string `opts:"mode=arg"`
	rt   *types.Root
}

func New(rt *types.Root) opts.Opts {
	sh := &shell{rt: rt}
	return opts.New(sh).Name("sh")
}

func (sh *shell) Run() {
	fmt.Printf("sh called %v\n", sh.Args)
	sh.rt.Pexec(sh.Args, func(ws *types.Workspace) {
		shArgs := []string{"-c", strings.Join(sh.Args, " ")}
		c := exec.Command("sh", shArgs...)
		c.Dir = ws.Path
		out, err := c.CombinedOutput()
		if err != nil {
			ws.Out = fmt.Sprintf("error getting sh %v %v\n", ws.Name(), err)
			return
		}
		ws.Out = string(out)
	})
	sh.rt.PrintOutput()
}
