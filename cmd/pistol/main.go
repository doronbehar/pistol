package main

import (
	"os"
	"log"

	"github.com/doronbehar/pistol"
	"github.com/alexflint/go-arg"
	"github.com/adrg/xdg"
)
type args struct {
	Config string `arg:"-c" help:"configuration file to use"`
	FilePath string `arg:"positional"`
	Extras []string `arg:"positional" help:"extra arguments passed to the command"`
}

var (
	Version string
)
func (args) Version() string {
	return Version
}

func main() {
	// Setup logger
	log.SetFlags(0)
	log.SetPrefix(os.Args[0] + ": ")
	var args args

	// Handle configuration file path
	xdgPaths := []string{"pistol/pistol.conf", "pistol.conf"}
	for _, xdgPath := range xdgPaths {
		defaultConfigPath, err := xdg.SearchConfigFile(xdgPath)
		// if a file was found
		if err == nil {
			args.Config = defaultConfigPath
			break
		}
	}
	// Setup cmdline arguments
	arg.MustParse(&args)

	// handle file argument with configuration
	if len(args.FilePath) == 0 {
		log.Fatalf("no arguments!")
		os.Exit(1)
	}
	previewer, err := pistol.NewPreviewer(args.FilePath, args.Config, args.Extras)
	if err != nil {
		log.Fatal(err)
		os.Exit(2)
	}
	if err := previewer.Write(os.Stdout); err != nil {
		log.Fatal(err)
		os.Exit(2)
	}
}
