package shell

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/wxio/wx/internal/types"
)

type shell struct {
	Args []string `opts:"mode=arg"`
	rt   *types.Root
}

// New constructor for sh sub command
func New(rt *types.Root) interface{} {
	return &shell{rt: rt}
}

func (sh *shell) Run() {
	fmt.Printf("sh called %v\n", sh.Args)
	sh.rt.Pexec(sh.Args, func(ws *types.Workspace) {
		shArgs := []string{"-c", strings.Join(sh.Args, " ")}
		c := exec.Command("sh", shArgs...)
		c.Dir = ws.Path
		out, err := c.CombinedOutput()
		if err != nil {
			ws.Out = fmt.Sprintf("error getting sh %v %v\n---\n%s\n---\n", ws.Name(), err, string(out))
			return
		}
		ws.Out = string(out)
	})
	sh.rt.PrintOutput()
}
