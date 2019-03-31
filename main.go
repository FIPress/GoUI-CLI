package main

import (
	"fmt"
	"os"
)

const (
	sampleDir = "./sample"
	//commandNeeded = "Please enter a task you "
	nameNeeded = "Please specify a directory/project name."
	fileExists = "File exists and is not a directory:"
	dirExists  = "Dir exists and not empty:"
)

func main() {
	initLogger()
	args := os.Args[1:]
	l := len(args)
	if l == 0 {
		help()
		return
	}

	switch args[0] {
	case "create":
		if l == 1 {
			fmt.Println(nameNeeded)
			return
		}
		create(args[1])
		break
	case "build":
		if l == 1 {
			fmt.Println(nameNeeded)
			return
		}
		build(args[1:])
		break
	case "run":
		run(args[1:])
	default:
		help()
	}

}

func help() {

}
