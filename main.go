package main

import (
	"fmt"
	"os"
	"log"

	"github.com/galdor/go-cmdline"
	"github.com/adrg/xdg"
)

func main() {
	// Setup logger
	log.SetFlags(0)
	log.SetPrefix(os.Args[0] + ": ")

	// Setup cmdline arguments
	cmd := cmdline.New()
	cmd.AddFlag("v", "verbosity","increase verbosity")
	cmd.AddOption("c", "config", "config", fmt.Sprintf("configuration file to use (defaults to %s/pistol.conf)", xdg.ConfigHome))
	cmd.AddArgument("file", "the file to preview")
	cmd.Parse(os.Args)

	// Handle configuration file path
	configPath := cmd.OptionValue("config")
	if configPath == "" {
		defaultConfigPath, err := xdg.SearchConfigFile("pistol.conf")
		if err != nil {
			log.Fatalf("could not find configuration file in the default location: %s/pistol.conf", xdg.ConfigHome)
			os.Exit(1)
		}
		configPath = defaultConfigPath
	}

	// handle file argument with configuration
	err := handle(configPath, cmd.ArgumentValue("file"), cmd.IsOptionSet("v"))
	if err != nil {
		log.Fatalf("%s\n", err)
		os.Exit(2)
	}
}
