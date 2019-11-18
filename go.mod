module github.com/wxio/wx

go 1.12

// replace github.com/jpillora/opts => /home/garym/go/src/github.com/millergarym/opts
replace github.com/jpillora/opts => github.com/millergarym/opts v1.1.10

require (
	github.com/Masterminds/vcs v1.13.1
	github.com/google/go-github v17.0.0+incompatible
	github.com/google/go-querystring v1.0.0 // indirect
	github.com/jpillora/md-tmpl v1.2.2
	github.com/jpillora/opts v1.1.0
	gopkg.in/src-d/go-git.v4 v4.12.0
	gopkg.in/yaml.v2 v2.2.2
)
