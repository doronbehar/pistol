//+build EMBED_MAGIC_DB

package main

import (
	_ "embed"
	"os"
	"fmt"
	"errors"
	"path/filepath"

	"github.com/adrg/xdg"
)

//go:embed magic.mgc
var magic_db_data []byte

func GetDbPath(magicmime_version int) (string, error) {
	// Create a copy of the magic database file with magic_db_data
	magic_db_dir := fmt.Sprintf("%s/pistol", xdg.DataHome)
	err := os.MkdirAll(magic_db_dir, 0755)
	if err != nil {
		return "", errors.New(fmt.Sprintf(
			"We've had issues creating a directory for the libmagic database at %s, error is: %s",
			magic_db_dir, err,
		))
	}
	magic_db_path := fmt.Sprintf("%s/%d.mgc", magic_db_dir, magicmime_version)
	// Don't write the database if there's already a file there.
	if _, err := os.Stat(magic_db_path); errors.Is(err, os.ErrNotExist) {
		old_dbs, err := filepath.Glob(filepath.Join(magic_db_dir, "*.mgc"))
		if err != nil {
			return "", err
		}
		for _, old_db := range old_dbs {
			err = os.Remove(old_db)
			if err != nil {
				return "", err
			}
		}
		err = os.WriteFile(magic_db_path, magic_db_data, 0644)
		if err != nil {
			return "", errors.New(fmt.Sprintf(
				"Could not create a copy of libmagic database at %s, error is: %s",
				magic_db_path, err,
			))
		}
	}
	return magic_db_path, nil
}
