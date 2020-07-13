package pistol

import (
	"bufio"
	"os"
	"io"
	"os/exec"
	// "fmt"
	"strings"
	"regexp"
	"path/filepath"
	"errors"

	she "github.com/alessio/shellescape"
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
		def := strings.Fields(scanner.Text())
		if len(def) == 0 {
			// Empty lines are skipped
			continue
		}
		match, err := regexp.MatchString(def[0], p.mimeType)
		if err != nil {
			return p, err
		}
		if match && len(def) > 1 {
			p.command = def[1]
			p.args = def[2:]
			return p, nil
		}
		match, err = regexp.MatchString("^#", def[0])
		if err != nil {
			return p, err
		}
		if match {
			// This is a comment, line skipped
			continue
		}
		// Test if fpath keyword is used at the beginning, indicating it's a
		// file path match we should be looking for
		if def[0] == "fpath" {
			log.Infof("found 'fpath' at the beginning, testing match against file path")
			if len(def) < 3 {
				log.Warnf("found 'fpath' keyword but it's line contains less then 3 words:\n%s", def)
				log.Warnf("skipping")
				continue
			}
		} else {
			// skip this line
			continue
		}
		absFpath, err := filepath.Abs(filePath)
		if err != nil {
			return p, err
		}
		match, err = regexp.MatchString(def[1], absFpath)
		if err != nil {
			return p, err
		}
		if match {
			log.Infof("matched file path against absFpath: %s", absFpath)
			p.command = def[2]
			p.args = def[3:]
			return p, nil
		}
	}
	log.Infof("didn't find a match in configuration for detected mimetype: %s\n", p.mimeType)
	return p, nil
}

func (p *Previewer) Write(w io.Writer) (error) {
	// if a match was encountered when the configuration file was read
	if p.command != "" {
		if match, _ := regexp.MatchString("%pistol-filename%", strings.Join(p.args, " ")); !match {
			return errors.New("no %pistol-filename% found in definition command")
		}
		var replStr string
		if p.command == "sh:" {
			replStr = she.Quote(p.filePath)
		} else {
			replStr = p.filePath
		}
		var cmd *exec.Cmd
		var argsOut []string
		for _, arg := range p.args {
			argsOut = append(argsOut, strings.ReplaceAll(arg, "%pistol-filename%", replStr))
		}
		if p.command == "sh:" {
			log.Infof("previewer's command is (shell interpreted): %#v\n", argsOut)
			cmd = exec.Command("sh", "-c", strings.Join(argsOut, " "))
		} else {
			log.Infof("previewer's command is (no shell) %#v with args: %#v\n", p.command, argsOut)
			cmd = exec.Command(p.command, argsOut...)
		}
		cmd.Stdout = w
		cmd.Stderr = os.Stderr
		if err := cmd.Start(); err != nil {
			log.Fatalf("We've had issues running your command: %v, %s", p.command, p.args)
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
