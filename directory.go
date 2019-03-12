package main

import (
	"os"
)

const (
	safe_rm_dir  string = "/.safe_rm/"
	entries_file string = ".entries"
)

func GetSafeRmDir() (string, error) {
	dir := os.Getenv("HOME") + safe_rm_dir
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			if err := os.Mkdir(dir, 0644); err != nil {
				return "", err
			}
		} else {
			return "", err
		}
	}

	return dir, nil
}

// will open the entry file, and create it if it doesn't exist
func OpenEntryFile() (*os.File, error) {
	return os.OpenFile(entries_file, os.O_CREATE, 0644)
}

func Move(path string) error {
	return os.Rename(path, "./"+path)
}
