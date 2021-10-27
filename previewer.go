package pistol

import (
	"bufio"
	"os"
	"io"
	"os/exec"
	// "fmt"
	"strconv"
	"strings"
	"regexp"
	"path/filepath"
	"errors"

	she "github.com/alessio/shellescape"
	log "github.com/sirupsen/logrus"
	"github.com/doronbehar/pistol/internal_writers"
	"github.com/doronbehar/magicmime"
)

// A type NewPreviewer returns
type Previewer struct {
	// The path to the magic.mgc database, usually empty
	MagicDb string
	// The file to be previewed
	FilePath string
	// Extra arguments passed to pistol
	Extras []string
	// The mime type detected
	MimeType string
	// The command that will be used to print the file. If empty, internal
	// writers are used. Concluded by the configuration file passed to
	// NewPreviewer.
	Command string
	// The arguments to the command
	Args []string
}

// Return a new Previewer that can .Write to an io.Writer.
//
//   * `filePath` is the file you'd like to preview.
//   * `configPath` should point to a configuration file in the format as explained
//   here: https://github.com/doronbehar/pistol#configuration
//
// You can set the environmental variable `PISTOL_DEBUG` to any non empty value
// and it will instruct NewPreviewer to spit additional log messages when
// parsing configPath and detecting the mime type of `filePath`. If configPath
// is an empty string, a warning message will be printed to stderr,
// unconditionally.
//
// `pistol`, the command line tool, searches for a default configuration file
// in ~/.config/pistol/pistol.conf. The API doesn't include this functionality.
//
// Mime type detection is provided by libmagic (through github.com/doronbehar/magicmime)
//
// Many mime types are handled internally by Pistol, see table here:
// https://github.com/doronbehar/pistol#introduction
func NewPreviewer(magic_db_path, filePath, configPath string, extras []string) (Previewer, error) {
	verbose := os.Getenv("PISTOL_DEBUG")
	if verbose != "" {
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(log.ErrorLevel)
	}
	// create an empty Previewer
	p := Previewer{}
	// opens the magic library
	if err := magicmime.OpenWithPath(magic_db_path, magicmime.MAGIC_MIME_TYPE | magicmime.MAGIC_SYMLINK); err != nil {
		return p, err
	}
	p.MagicDb = magic_db_path
	// get mimetype of given file, we don't care about the extension
	mimetype, err := magicmime.TypeByFile(filePath)
	defer magicmime.Close()
	if err != nil {
		return p, err
	}
	p.MimeType = mimetype
	log.Infof("detected mimetype is %s", p.MimeType)
	p.FilePath = filePath
	p.Extras = extras
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
		match, err := regexp.MatchString(def[0], p.MimeType)
		if err != nil {
			return p, err
		}
		if match && len(def) > 1 {
			p.Command = def[1]
			p.Args = def[2:]
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
			p.Command = def[2]
			p.Args = def[3:]
			return p, nil
		}
	}
	log.Infof("didn't find a match in configuration for detected mimetype: %s\n", p.MimeType)
	return p, nil
}

// Write to an io.Writer the Previewer's text - concluded mostly by
// NewPreviewer
func (p *Previewer) Write(w io.Writer) (error) {
	// if a match was encountered when the configuration file was read
	if p.Command != "" {
		if match, _ := regexp.MatchString("%pistol-filename%", strings.Join(p.Args, " ")); !match {
			return errors.New("no %pistol-filename% found in definition command")
		}
		var replStr string
		if p.Command == "sh:" {
			replStr = she.Quote(p.FilePath)
		} else {
			replStr = p.FilePath
		}
		var cmd *exec.Cmd
		var argsOut []string

		re := regexp.MustCompile(`%pistol-extra([0-9]+)%`)

		for _, arg := range p.Args {
			argAux := arg
			if(re.MatchString(arg)) {
				// We iterate all indices of matches in every command line
				// argument written in the config, because the match can occur
				// in multiple arguments, see #56.
				allIndexes := re.FindAllStringSubmatchIndex(arg, -1)
				for _, loc := range allIndexes {
					// We try to convert the string found in the argument to a
					// number.
					auxInt, err := strconv.Atoi(arg[loc[2]:loc[3]])
					current := arg[loc[0]:loc[1]]
					if (err == nil && len(p.Extras) > auxInt) {
						// substitute the %pistol-extra[#]% argument in the
						// final CLI string.
						argAux = strings.ReplaceAll(argAux, current, p.Extras[auxInt])
					} else {
						argAux = strings.ReplaceAll(argAux, current, "")
					}
				}
			} else {
				argAux = strings.ReplaceAll(argAux, "%pistol-filename%", replStr)
			}
			if(len(argAux) > 0) {
				argsOut = append(argsOut, argAux)
			}
		}

		if p.Command == "sh:" {
			log.Infof("previewer's command is (shell interpreted): %#v\n", argsOut)
			cmd = exec.Command("sh", "-c", strings.Join(argsOut, " "))
		} else {
			log.Infof("previewer's command is (no shell) %#v with args: %#v\n", p.Command, argsOut)
			cmd = exec.Command(p.Command, argsOut...)
		}
		cmd.Stdout = w
		cmd.Stderr = os.Stderr
		if err := cmd.Start(); err != nil {
			log.Fatalf("We've had issues running your command: %v, %s", p.Command, p.Args)
			return err
		}
		return cmd.Wait()
	} else {
		// try to match with internal writers
		internal_writer, err := pistol.MatchInternalWriter(p.MagicDb, p.MimeType, p.FilePath)
		if err != nil {
			return err
		}
		if err := internal_writer(w); err != nil {
			return err
		}
	}
	return nil
}
