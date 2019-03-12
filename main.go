package main

import (
	"fmt"
	"os"
)

func main() {
	dir, err := GetSafeRmDir()
	if err != nil {
		fmt.Printf("Error opening directory. %s\n", err.Error())
		return
	}

	entryFile, err := OpenEntryFile()
	if err != nil {
		fmt.Println("Couldn't find or open entry file. ", err)
		return
	}

	entries, err := ReadFromFile(entryFile)
	if err != nil {
		fmt.Println("Couldn't read from file")
		fmt.Println("Making new entry file")
		entries = New()
	}


	c := make(chan []FileEntry)

	go FindStagedToRemove(c, entries, dir)

	args, err := ParseArgs(os.Args[1:])
	if err != nil {
		fmt.Println("Error parsing arguments: ", err)
		return
	}

	if err := moveFile(args.file, dir); err != nil {
		fmt.Println("Error moving file: ", err)
		return
	}

	var kind FileKind
	if args.dir {
		kind = Directory
	} else {
		kind = File
	}

	staged := <- c

	if err := deleteStaged(staged, dir, nil); err != nil {
		fmt.Println("Had trouble removing file: ", err)
	}

	if args.file != "" {
		entries.AddEntry(args.file, kind, args.time)
		if err := moveFile(args.file, dir); err != nil {
			fmt.Println("Couldn't move file to safe_rm dir. ", err)
		}
	}

	if err := entries.WriteToFile(entryFile); err != nil {
		fmt.Println("Couldn't write to file. It is recommended to retrieve all files manually at ", dir)
	}
}

func moveFile(str, dir string) error {
	return os.Rename(str, dir+str)
}

func deleteStaged(staged []FileEntry, dir string, exclude *string) (e error) {

	maybeDelete := func(file FileEntry) {
		if exclude == nil || file.name != *exclude {
			if file.kind == Directory {
				e = os.RemoveAll(file.name)
			} else {
				e = os.Remove(file.name)
			}
		}
	}

	for _, file := range staged {
		maybeDelete(file)
	}

	return nil
}
