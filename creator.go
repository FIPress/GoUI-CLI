package main

import (
	"fmt"
	"github.com/fipress/fiputil"
	"os"
	"path/filepath"
)

const (
	packageConfTmpl = `
# This config file is for GoUI-CLI to package your app
name={{.Name}}
versionCode={{.VersionCode}}
versionName={{.VersionName}}
`
)

func createProject(name string, ctx *context) {
	fullPath := filepath.Join(ctx.workingDir, name)
	fi, err := os.Stat(fullPath)
	if err == nil {
		if !fi.IsDir() {
			fmt.Println(fileExists, fullPath)
			return
		}

		if fi.Size() > 100 {
			fmt.Println(dirExists, fullPath)
			return
		}
	} else {
		err = os.MkdirAll(fullPath, 0755)
		if err != nil {
			fmt.Println("Create directory", fullPath, " error:", err)
			return
		}
	}

	src := filepath.Join(ctx.binDir, sampleDir)
	err = fiputil.CopyDir(src, fullPath, nil)
	/*func(fullPath string) bool {
		return !strings.HasSuffix(fullPath, "tmpl")
	}*/
	if err != nil {
		fmt.Println("Copy directory failed, error:", err)
		return
	}

	createPackageConfig(filepath.Join(fullPath, packageConfigFile))

	fmt.Println("A new project created.")
}
