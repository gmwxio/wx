


package main // import "github.com/wxio/wx"

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"github.com/wxio/wx/cmd"
)

func main() {
	c := make(chan struct{}, 1)
	go func() {
		c <- struct{}{}
		e := http.ListenAndServe(":8080", nil)
		if e != nil {
			fmt.Printf("Can't listen err: %v\n", e)
		}
	}()
	<-c
	cmd.Execute()
}
