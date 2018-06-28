package cmd

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	yaml "gopkg.in/yaml.v2"
)

var (
	wxconfig WXConfig
	oauth    string
	repos    Repos
	branch   string
	title    string
	project  *string

	dryrun      *bool
	groupOutput *bool
	projectFlag *string
	parallelism *int
)

func RootCmd() *cobra.Command {
	return rootCmd
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:              "wx",
	Short:            "command for dealing with multiple repo. For dev and ci environments that works",
	Long:             ``,
	TraverseChildren: true,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	parallelism = rootCmd.PersistentFlags().IntP("parallelism", "x", 0, "number of repo to run the command in on parallel")
	projectFlag = rootCmd.PersistentFlags().StringP("project", "p", "", "project name. as in .wx.<project>.yaml")
	dryrun = rootCmd.PersistentFlags().BoolP("dryrun", "d", false, "Don't make remote changes")
	groupOutput = rootCmd.PersistentFlags().BoolP("group-output", "g", false,
		`Group the output and show all the repos with the same result`)

	project = scanCmd.PersistentFlags().StringP("write", "w", "", "write a .wx.<project>.yaml file and .wx.yaml pointing to the project")
	wxconfig.Matches.GitOrigin = rootCmd.PersistentFlags().StringP("git.origin", "", "", "git.origin, overrides .wx.yaml")
	repos = Repos{}
	repos.DefaultGitOwner = rootCmd.PersistentFlags().StringP("git.default.owner", "o", "", "git.default.owner, overrides .wx.yaml")

	// gp := rootCmd.PersistentFlags().Lookup("group-output")
	// sgp := rootCmd.PersistentFlags().ShorthandLookup("g")
	// fmt.Fprintf(os.Stderr,"xxxxxx \n\t%+v \n\t%+v\n", *gp, *sgp)
	// fmt.Fprintf(os.Stderr,"\t%p\n", gp)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	flag.CommandLine.Parse([]string{})

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "cant get cwd .. %v\n", err)
		os.Exit(1)
	}
	last := ""
	for {
		if _, err := os.Open(".wx.yaml"); err != nil {
			wd, err := os.Getwd()
			if err != nil {
				fmt.Fprintf(os.Stderr, "cant get cwd .. %v\n", err)
				os.Exit(1)
			}
			if err = os.Chdir(".."); err != nil {
				fmt.Fprintf(os.Stderr, "cant open .. %v\n", err)
				os.Exit(1)
			}
			if last == wd {
				fmt.Fprintf(os.Stderr, "reached root without finding .wx.yaml using %v\n", cwd)
				if err = os.Chdir(cwd); err != nil {
					fmt.Fprintf(os.Stderr, "cant open %v %v\n", cwd, err)
					os.Exit(1)
				}
				break
			}
			last = wd
		} else {
			wd, err := os.Getwd()
			if err != nil {
				fmt.Fprintf(os.Stderr, "cant get cwd .. %v\n", err)
				os.Exit(1)
			}
			last = wd
			fmt.Fprintf(os.Stderr, "project dir %v\n", wd)
			break
		}
	}

	// todo bind glog flags
	configFile := ".wx.yaml"
	viper.SetConfigFile(configFile)
	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintf(os.Stderr, "Using config file: %s\n", viper.ConfigFileUsed())
	} else {
		fmt.Fprintf(os.Stderr, "Error reading config file: %s\n", viper.ConfigFileUsed())
		// log.Fatalf("Error reading config file '%s'\n", viper.ConfigFileUsed())
		// os.Exit(1)
	}

	cdata, err := ioutil.ReadFile(configFile)
	// fmt.Fprintln(os.Stderr,string(data))
	if err != nil {
		fmt.Fprintf(os.Stderr, "project file %s not found\n", configFile)
	} else {
		wxconfig = WXConfig{}
		err = yaml.Unmarshal(cdata, &wxconfig)
		if err != nil {
			log.Fatalf("can't read %s: %+v", configFile, err)
			os.Exit(1)
		}
	}

	viper.SetConfigFile(".tokens.yaml")
	if err := viper.MergeInConfig(); err == nil {
		fmt.Fprintf(os.Stderr, "Using config file: %s\n", viper.ConfigFileUsed())
	} else {
		log.Printf("Error reading config file '%s'\n", viper.ConfigFileUsed())
	}

	viper.AutomaticEnv() // read in environment variables that match

	oauth = viper.GetString("GITHUB_API_TOKEN")
	if oauth == "" {
		fmt.Fprintf(os.Stderr, "now git access token found\n")
	}

	filename := `.wx.yaml`
	if projectFlag != nil && *projectFlag != "" {
		filename = ".wx." + *projectFlag + ".yaml"
	} else {
		project := viper.GetString("project")
		if project != "" {
			filename = ".wx." + project + ".yaml"
		}
	}
	data, err := ioutil.ReadFile(filename)
	// fmt.Fprintln(os.Stderr,string(data))
	if err != nil {
		fmt.Fprintf(os.Stderr, "project file %s not found\n", filename)
	} else {
		err = yaml.Unmarshal(data, &repos)
		if err != nil {
			fmt.Fprintf(os.Stderr, "can't read %s: %v\n", filename, err)
		} else {
			fmt.Fprintf(os.Stderr, "Read project file: %s\n", filename)
		}
	}
}

func printOutput() {
	if *groupOutput {
		omap := make(map[string][]string)
		omapO := make([]string, 0)
		for _, r := range repos.Repos {
			if _, ex := omap[r.out]; !ex {
				omapO = append(omapO, r.out)
			}
			omap[r.out] = append(omap[r.out], r.Name())
		}
		for _, k := range omapO {
			fmt.Fprintf(os.Stderr, `
---- Repositories: -------------------------------------------------------------------
%[2]v
----    message:   -------------------------------------------------------------------
%[1]v--------------------------------------------------------------------------------------
`, k, omap[k])
		}
	} else {
		for _, r := range repos.Repos {
			fmt.Fprintf(os.Stderr, "--------- %s ---------:\n%s", r.Name(), r.out)
		}
	}
}
