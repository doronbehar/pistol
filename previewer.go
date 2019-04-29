package main

import (
	"log"
	"bufio"
	"os"
	"io"
	"os/exec"
	"fmt"
	"regexp"
	"strings"
	// "errors"

	"github.com/rakyll/magicmime"
)

type Previewer struct {
	mimetype string
	verbose bool
	// if the following are set, we use them, if not, we revert to using internal mechanisms
	command string
	args []string
}

func NewPreviewer(filePath, configPath string, verbose bool) (Previewer, error) {
	// create an empty Previewer
	p := Previewer{}
	p.verbose = verbose
	// opens the magic library
	if err := magicmime.Open(magicmime.MAGIC_MIME_TYPE | magicmime.MAGIC_SYMLINK); err != nil {
		return p, err
	}
	// get mimetype of given file, we don't care about the extension
	mimetype, err := magicmime.TypeByFile(filePath)
	defer magicmime.Close()
	if err != nil {
		return p, err
	}
	p.mimetype = mimetype
	if verbose {
		log.Printf("detected mimetype is %s", p.mimetype)
		log.Printf("reading configuration from %s", configPath)
	}
	file, err := os.Open(configPath)
	if err != nil {
		return p, err
	}
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		def := strings.Split(scanner.Text(), " ")
		match, err := regexp.MatchString(def[0], p.mimetype)
		if err != nil {
			return p, err
		}
		if match {
			p.command = def[1]
			for _, arg := range def[2:] {
				if match, _ := regexp.MatchString("%s", arg); match {
					p.args = append(p.args, fmt.Sprintf(arg, filePath))
				} else {
					p.args = append(p.args, arg)
				}
			}
			if verbose {
				log.Printf("previewer's command is %s %s\n", p.command, p.args)
			}
			break
		}
	}
	if verbose && p.command == "" {
		log.Printf("didn't find a match for detected mimetype %s\n", p.mimetype)
	}
	return p, nil
}

func (p *Previewer) Write(w io.Writer) {
	// if a match was encountered when the configuration file was read
	if p.command != "" {
		cmd := exec.Command(p.command, p.args...)
		cmd.Stdout = w
		if err := cmd.Start(); err != nil {
			panic(err)
		}
		cmd.Wait()
		return
	}
	// if the above didn't happen, we try to use an internal viewer
}
