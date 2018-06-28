package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	git "gopkg.in/src-d/go-git.v4"
	yaml "gopkg.in/yaml.v2"
)

// scanCmd represents the scan command
var scanCmd = &cobra.Command{
	Use:   "scan [options]",
	Short: "recursively scans directories look for match repo",
	Long:  ``,
	Run:   scanObj.Run,
}

type scanStr struct {
}

var scanObj = &scanStr{}

func (so *scanStr) Run(cmd *cobra.Command, args []string) {
	fmt.Fprintln(os.Stderr, "scan called")
	if *wxconfig.Matches.GitOrigin == "" {
		fmt.Println(`Git orgin match not specified in .wx.config
eg
matches:
  git.origin: git@github.com:wxio			
`)
		os.Exit(1)
	}
	srepos := make([]*Repo, 0)
	collectDir(".", &srepos)
	temprepos := Repos{
		DefaultGitOwner: repos.DefaultGitOwner,
		Repos:           srepos,
	}
	out, err := yaml.Marshal(&temprepos)
	if err != nil {
		log.Fatalf("marshall: %v", err)
		os.Exit(1)
	}
	fmt.Println(string(out))
	if *project != "" {
		proj := ".wx." + *project + ".yaml"
		err := ioutil.WriteFile(proj, out, os.ModePerm)
		wd, _ := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "can't write file %v %v %v\n", proj, wd, err)
		} else {
			fmt.Fprintf(os.Stderr, "wrote %v in %v\n", proj, wd)
		}
		wxconfig.Project = *project
		out, err := yaml.Marshal(&wxconfig)
		if err != nil {
			log.Fatalf("marshall: %v", err)
		}
		err = ioutil.WriteFile(".wx.yaml", out, os.ModePerm)
		if err != nil {
			fmt.Fprintf(os.Stderr, "can't write file %v %v %v\n", ".wx.yaml", wd, err)
		} else {
			fmt.Fprintf(os.Stderr, "wrote %v in %v\n", ".wx.yaml", wd)
		}
	}
}

func collectDir(dir string, dirs *[]*Repo) {
	if gr, err := git.PlainOpen(dir); err == nil {
		if cg, err := gr.Config(); err == nil {
			for _, url := range cg.Remotes["origin"].URLs {
				if strings.Contains(url, *wxconfig.Matches.GitOrigin) {
					re := &Repo{Path: dir}
					*dirs = append(*dirs, re)
					// if dir == "." {
					// 	re.Url = url
					// }
					if re.Address() != url {
						re.Url = url
					}
					break
				}
			}
		} else {
			fmt.Printf("error getting git config '%v'\n", dir)
		}
	}
	fs, err := ioutil.ReadDir(dir)
	if err != nil {
		fmt.Printf("error reading dirs '%v'\n", dir)
	}
	for _, d := range fs {
		if d.IsDir() && d.Name()[0] != '.' {
			if dir == "." {
				collectDir(d.Name(), dirs)
			} else {
				collectDir(dir+"/"+d.Name(), dirs)
			}
		}
	}
}

func init() {
	rootCmd.AddCommand(scanCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// scanCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// scanCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
