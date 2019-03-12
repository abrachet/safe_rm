package main

import (
	"fmt"
	"log"
	"os"
	gotime "time"
)

type Time = uint64

const defaultTime Time = 24

type FileKind = int

const (
	File      FileKind = 0
	Directory FileKind = 1
)

type FileEntry struct {
	time Time
	kind FileKind
	name string
}

type Entries struct {
	defaultTime Time
	files       []FileEntry
}

func New() *Entries {
	ret := new(Entries)

	ret.defaultTime = defaultTime

	return ret
}

func (ent *Entries) AddEntry(name string, kind FileKind, time Time) {
	if time == 0 {
		time = ent.defaultTime
	}

	if kind == Directory {
		if name[len(name)-1] != '/' {
			name += "/"
		}
	}

	ent.files = append(ent.files, FileEntry{time, kind, name})
}

func (ent *Entries) RemoveEntry(name string) error {

	for i, entry := range ent.files {
		if entry.name == name {
			ent.files = append(ent.files[:i], ent.files[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("entry not found")
}

/*
	Example of entry file:

	Entry {
		defaultTime: 24,
		numFiles: 1,
		files: [
			fileEntry { time: 100000, name: "filename"},
		]
	}

*/
func parseEntry(file *os.File) (time Time, name string, err error) {

	_, err = fmt.Fscanf(file,
		"        fileEntry { time: %d, name: %q},\n", &time, &name)

	return
}

func ReadFromFile(file *os.File) (*Entries, error) {

	invalid_file := fmt.Errorf("Invalid entry file")

	stat, err := file.Stat()
	if err != nil {
		return New(), err
	}

	if stat.Size() == 0 {
		return New(), nil
	}

	ret := new(Entries)

	if _, err := fmt.Fscanf(file, "Entry {\n"); err != nil {
		return nil, invalid_file
	}

	if _, err := fmt.Fscanf(file, "    defaultTime: %d,\n", &ret.defaultTime); err != nil {
		return ret, invalid_file
	}

	var num_entries int

	if _, err := fmt.Fscanf(file, "    numFiles: %d,\n", &num_entries); err != nil {
		return ret, invalid_file
	}

	if _, err := fmt.Fscanf(file, "    files: [\n"); err != nil {
		return ret, invalid_file
	}

	for i := 0; i < num_entries; i++ {
		time, name, err := parseEntry(file)
		if err != nil {
			return ret, invalid_file
		}

		var kind FileKind

		if name[len(name)-1] == '/' {
			kind = Directory
		} else {
			kind = File
		}

		ret.files = append(ret.files, FileEntry{time, kind, name})
	}

	return ret, nil
}

func (ent *Entries) WriteToFile(file *os.File) error {

	if _, err := fmt.Fprint(file, "Entry {\n"); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(file, "    defaultTime: %d,\n", ent.defaultTime); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(file, "    numFiles: %d,\n", len(ent.files)); err != nil {
		return err
	}

	if _, err := fmt.Fprint(file, "    files: [\n"); err != nil {
		return err
	}

	for _, entry := range ent.files {
		if _, err := fmt.Fprintf(file, "        fileEntry { time: %d, name: %q},\n", entry.time, entry.name); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprint(file, "    ]\n}"); err != nil {
		return err
	}

	return nil
}

func (ent *Entries) PrintEntries() {
	fmt.Printf("Default time: %d\n", ent.defaultTime)

	for _, i := range ent.files {
		fmt.Printf("file: %s time: %d\n", i.name, i.time)
	}
}

func (file *FileEntry) shouldRemove(now gotime.Time) (bool, error) {
	info, err := os.Stat(file.name)
	if err != nil {
		log.Fatalln("couldn't get file info with stat(2)", err)
		return false, err
	}

	modTime := info.ModTime()

	modTime.Add(gotime.Second * gotime.Duration(file.time))

	if modTime.Before(now) {
		return false, nil
	}

	return true, nil

}

func FindStagedToRemove(c chan []FileEntry, entries *Entries, dir string) {
	ret := make([]FileEntry, 0)

	now := gotime.Now()

	for _, entry := range entries.files {
		if shouldDel, err := entry.shouldRemove(now); err != nil {
			log.Println(err)
		} else if shouldDel {
			log.Printf("Marking %s to be unlinked\n", entry.name)
			ret = append(ret, entry)
		}
	}

	c <- ret
}
