//+build !EMBED_MAGIC_DB

package main

func GetDbPath(magicmime_version int) (string, error) {
	// And by that use the default location for the magic.mgc database
	return "", nil
}
