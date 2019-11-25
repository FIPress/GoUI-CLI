package main

import (
	"os"
)

const (
	sampleDir         = "./sample"
	packageConfigFile = "package.rj"
	//commandNeeded = "Please enter a task you "
	nameNeeded = "Please specify a directory/project name."
	fileExists = "File exists and is not a directory:"
	dirExists  = "Dir exists and not empty:"
)

var (

	//tempDir       string
	executable string
	simulator  = true
)

func main() {
	args := os.Args[1:]

	l := len(args)

	if l == 0 {
		showHelp()
		return
	}

	//cxt.initTask(args[0])
	task := taskFromString(args[0])
	if task == help {
		showHelp()
		return
	} else {
		if l < 2 {
			showHelp()
			return
		}

		ctx, ok := newContext()
		if !ok {
			showHelp()
			return
		}

		if task == create {
			createProject(args[1], ctx)
		} else {
			if l > 2 {
				if args[2] == "-prod" {
					ctx.isProd = true
				}
			}

			ctx.loadConfig()
			pkg, ok := getPackager(ctx, args[1])

			if pkg.getPlatform() == unknown {
				showHelp()
				return
			}

			if !ok {
				return
			}

			pkg.buildAndPack()
		}
	}

}

func showHelp() {

}
