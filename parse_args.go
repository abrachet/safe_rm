package main

import (
	"fmt"
	"testing"
)

type Arguments struct {
	verbose bool
	dir     bool
	now     bool

	file string
	time Time
}

type expected struct {
	expect func(string) bool
	parse  func([]string, int, *Arguments) (int, error)
}

func expect(ex expected, args []string, index int, arguments *Arguments) (int, error) {
	if ex.expect(args[index]) {
		return ex.parse(args, index, arguments)
	}

	return index, fmt.Errorf("didn't match")
}

func printHelp() {
	fmt.Println("safe_rm [-vrRnN] filename [%d%c]")
	fmt.Println("    move a file to a directory for a specified time before unlinking")
	fmt.Println("Copyright Alex Brachet 2019")
}

func printVersion() {
	fmt.Println("safe_rm version 0.1")
	fmt.Println("Alex Brachet 2019 (abrachet@purdue.edu")
	fmt.Println("github.com/abrachet/safe_rm for bugs and feature requests")
}

func parseOptions(args []string, index int, arguments *Arguments) (int, error) {
	if args[index][1] == '-' {
		switch args[index] {
		case "help":
			printHelp()
		case "version":
			printVersion()
		case "verbose":
			arguments.verbose = true
		case "directory":
			fallthrough
		case "recursive":
			arguments.dir = true
		case "now":
			fallthrough
		case "rm":
			arguments.now = true
		default:
			return 0, fmt.Errorf("long option '%s' not matched", args[index][2:])
		}
	}

	for i := 1; i < len(args[index]); i++ {
		switch args[index][i] {
		case 'v':
			arguments.verbose = true
		case 'r':
			fallthrough
		case 'R':
			arguments.verbose = true
		case 'n':
			fallthrough
		case 'N':
			arguments.verbose = true
		default:
			return 0, fmt.Errorf("short option '%c' not matched", args[index][i])
		}
	}

	return index + 1, nil
}

var (
	option = expected{
		func(s string) bool {
			return s[0] == '-'
		},
		parseOptions,
	}

	time = expected{
		func(_ string) bool {
			return true
		},
		func(args []string, index int, arguments *Arguments) (int, error) {
			secs, err := parseTime(args[index])
			if err != nil {
				return index, err
			}

			arguments.time = secs

			return index + 1, nil
		},
	}

	file = expected{
		func(_ string) bool {
			return true
		},
		func(args []string, index int, arguments *Arguments) (int, error) {
			arguments.file = args[index]

			// don't look for time if we are at the end of the list
			if index == len(args)-1 {
				return index, nil
			}

			return expect(time, args, index+1, arguments)
		},
	}
)

func ParseArgs(args []string) (arguments *Arguments, e error) {
	arguments = new(Arguments)
	e = nil

	for index := 0; index < len(args); {
		i, err := expect(option, args, index, arguments)
		if err != nil {
			e = err
			return
		}

		index += i
		if index >= len(args) {
			break
		}

		i, err = expect(file, args, index, arguments)
		if err != nil {
			e = err
			return
		}

		index += i
	}

	return
}

const (
	secs    = 1
	minutes = 60 * secs
	hours   = 60 * minutes
	days    = 24 * hours
)

func parseTime(time string) (Time, error) {
	var num int64
	var char byte

	_, err := fmt.Sscanf(time, "%d%c", &num, &char)
	if err != nil {
		_, err = fmt.Sscanf(time, "%d", &num)
		if err != nil {
			return 0, err
		}
		char = 'd'
	}

	if num < 0 {
		return 5, fmt.Errorf("negative number")
	}

	switch char {
	case 's':
		num *= secs
	case 'm':
		num *= minutes
	case 'h':
		num *= hours
	case 'd':
		num *= days
	}

	return Time(num), nil
}

func TestParseTime(t *testing.T) {
	got, _ := parseTime("1s")
	if got != 1 {
		t.Errorf("Expected 1 got %d", got)
	}

	got, _ = parseTime("2d")
	if got != 2*days {
		t.Errorf("Expected %d got %d", 2*days, got)
	}

	got, err := parseTime("2")
	if got != 2*days {
		t.Errorf("Cannot handle args without char")
	}

	got, err = parseTime("-1d")
	if err == nil {
		t.Errorf("Should error on negative numbers but recieved %d", got)
	}

	if _, err = parseTime("--flag"); err == nil {
		t.Errorf("--flag is an invalid time")
	}
}
