package pistol

import (
	"bufio"
	"os"
	"io"
	"os/exec"
	"fmt"
	"strings"
	"regexp"

	log "github.com/sirupsen/logrus"
	"github.com/doronbehar/pistol/internal_writers"
	"github.com/rakyll/magicmime"
)

type Previewer struct {
	filePath string
	mimeType string
	// if the following are set, we use them, if not, we revert to using internal mechanisms
	command string
	args []string
}

func NewPreviewer(filePath, configPath string) (Previewer, error) {
	verbose := os.Getenv("PISTOL_DEBUG")
	if verbose != "" {
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(log.ErrorLevel)
	}
	// create an empty Previewer
	p := Previewer{}
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
	p.mimeType = mimetype
	log.Infof("detected mimetype is %s", p.mimeType)
	p.filePath = filePath
	// If configuration file doesn't exist, we don't try to read it
	if configPath == "" {
		log.Warnf("configuration file was not supplied")
		return p, nil
	}
	file, err := os.Open(configPath)
	if err != nil {
		return p, err
	}
	log.Infof("reading configuration from %s", configPath)
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		def := strings.Split(scanner.Text(), " ")
		match, err := regexp.MatchString(def[0], p.mimeType)
		if err != nil {
			return p, err
		}
		if match && len(def) > 1 {
			p.command = def[1]
			for _, arg := range def[2:] {
				if match, _ := regexp.MatchString("%s", arg); match {
					p.args = append(p.args, fmt.Sprintf(arg, filePath))
				} else {
					p.args = append(p.args, arg)
				}
			}
			log.Infof("previewer's command is %s %s\n", p.command, p.args)
			return p, nil
		}
		// Test if fpath keyword is used at the beginning, indicating it's a
		// file path match we should be looking for
		if def[0] == "fpath" && len(def) > 2 {
			log.Infof("found 'fpath' at the beginning, testing match against file path")
		} else {
			break
		}
		match, err = regexp.MatchString(def[1], filePath)
		if err != nil {
			return p, err
		}
		if match {
			log.Infof("matched file path match against filePath")
			p.command = def[2]
			for _, arg := range def[3:] {
				if match, _ := regexp.MatchString("%s", arg); match {
					p.args = append(p.args, fmt.Sprintf(arg, filePath))
				} else {
					p.args = append(p.args, arg)
				}
			}
			log.Infof("previewer's command is %s %s\n", p.command, p.args)
			return p, nil
		}
	}
	log.Infof("didn't find a match in configuration for detected mimetype: %s\n", p.mimeType)
	return p, nil
}

func (p *Previewer) Write(w io.Writer) (error) {
	// if a match was encountered when the configuration file was read
	if p.command != "" {
		cmd := exec.Command(p.command, p.args...)
		cmd.Stdout = w
		if err := cmd.Start(); err != nil {
			return err
		}
		cmd.Wait()
	} else {
		// try to match with internal writers
		internal_writer, err := pistol.MatchInternalWriter(p.mimeType, p.filePath)
		if err != nil {
			return err
		}
		if err := internal_writer(w); err != nil {
			return err
		}
	}
	return nil
}
