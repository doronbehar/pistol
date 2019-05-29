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
	defaultConfigPath, _ := xdg.SearchConfigFile("pistol.conf")
	cmd := cmdline.New()
	cmd.AddOption("c", "config", "config file", "configuration file to use")
	cmd.SetOptionDefault("config", defaultConfigPath)
	cmd.AddTrailingArguments("file", "the file to preview")
	cmd.Parse(os.Args)

	// Handle configuration file path
	configPath := cmd.OptionValue("config")

	// handle file argument with configuration
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
