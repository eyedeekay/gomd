package main

import (
	"github.com/nochso/gomd/embeditor"
	. "github.com/nochso/gomd/embeditor"

	"gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	args := InputArgs{
		Port: kingpin.Flag("port", "Listening port used by webserver").Short('p').Default("1110").Int(),
		File: kingpin.Arg("file", "Markdown file").String(),
	}
	// Parse command line arguments
	kingpin.Version("0.0.1")
	kingpin.Parse()
	embeditor.Runner(args)
}
