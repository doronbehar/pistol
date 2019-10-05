package main

import (
	"os"
	"log"

	"github.com/doronbehar/pistol"
	"github.com/galdor/go-cmdline"
	"github.com/adrg/xdg"
)

func main() {
	// Setup logger
	log.SetFlags(0)
	log.SetPrefix(os.Args[0] + ": ")

	// Setup cmdline arguments
	cmd := cmdline.New()
	cmd.AddOption("c", "config", "config file", "configuration file to use (defaults to ~/.config/pistol/pistol.conf)")
	cmd.AddTrailingArguments("file", "the file to preview")
	cmd.Parse(os.Args)

	// Handle configuration file path
	xdgPaths := []string{"pistol/pistol.conf", "pistol.conf"}
	for _, xdgPath := range xdgPaths {
		defaultConfigPath, err := xdg.SearchConfigFile(xdgPath)
		// if a file was found
		if err == nil {
			cmd.SetOptionDefault("config", defaultConfigPath)
			break
		}
	}
	configPath := cmd.OptionValue("config")

	// handle file argument with configuration
	if len(cmd.TrailingArgumentsValues("file")) == 0 {
		log.Fatalf("no arguments!")
		os.Exit(1)
	}
	previewer, err := pistol.NewPreviewer(cmd.TrailingArgumentsValues("file")[0], configPath)
	if err != nil {
		log.Fatal(err)
		os.Exit(2)
	}
	if err := previewer.Write(os.Stdout); err != nil {
		log.Fatal(err)
		os.Exit(2)
	}
}
