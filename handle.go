package main

import (
	"log"
	"bufio"
	"os"
	"os/exec"
	"fmt"
	"regexp"
	"strings"
	"errors"

	tm "github.com/buger/goterm"
	"github.com/gabriel-vasile/mimetype"
)

func run_command(command string, args []string) (error) {
	cmd := exec.Command(command, args...)
	r, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err:= cmd.Start(); err != nil {
		return err
	}
	s := bufio.NewScanner(r)
	s.Split(bufio.ScanLines)
	line_counter := 0
	tm_height := tm.Height()
	for s.Scan() {
		fmt.Println(s.Text())
		line_counter++
		if line_counter == tm_height {
			return nil
		}
	}
	return nil
}

func handle(configFile, filePath string, verbose bool) (error) {
	// get mimetype of given file, we don't care about the extension
	mime, _, err := mimetype.DetectFile(filePath)
	if err != nil {
		return err
	}
	if verbose {
		log.Printf("detected mimetype is %s", mime)
		log.Printf("reading configuration from %s", configFile)
	}
	file, err := os.Open(configFile)
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		def := strings.Split(scanner.Text(), " ")
		match, err := regexp.MatchString(def[0], mime)
		if err != nil {
			return err
		}
		if match {
			var command string
			var args []string
			command = def[1]
			for _, arg := range def[2:] {
				if match, _ := regexp.MatchString("%s", arg); match {
					args = append(args, fmt.Sprintf(arg, filePath))
				} else {
					args = append(args, arg)
				}
			}
			if verbose {
				log.Printf("running command %s %s\n", command, args)
			}
			return run_command(command, args)
		}
	}
	return errors.New(fmt.Sprintf("couldn't find a matching string for mime %s", mime))
}
