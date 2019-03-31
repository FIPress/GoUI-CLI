package main

import (
	"fmt"
	"github.com/fipress/fiputil"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

const (
	confFile   = "package.conf"
	tmplSuffix = ".tmpl"
)

type PackageConfig struct {
	Name string
}

func create(name string) {
	fi, err := os.Stat(name)
	if err == nil {
		if !fi.IsDir() {
			fmt.Println(fileExists, name)
			return
		}

		if fi.Size() > 100 {
			fmt.Println(dirExists, name)
			return
		}
	} else {
		err = os.MkdirAll(name, 0755)
		if err != nil {
			fmt.Println("Create directory", name, " error:", err)
			return
		}
	}

	exPath, err := getExecutableDir()
	if err != nil {
		fmt.Println("Get executable directory failed, error:", err)
		return
	}

	src := filepath.Join(exPath, sampleDir)
	err = fiputil.CopyDir(src, name, func(name string) bool {
		return !strings.HasSuffix(name, "tmpl")
	})
	if err != nil {
		fmt.Println("Copy directory failed, error:", err)
		return
	}

	t, err := template.ParseFiles(filepath.Join(src, confFile+tmplSuffix))
	if err != nil {
		fmt.Println("Template config file not found")
		return
	}

	out, err := os.OpenFile(filepath.Join(name, confFile), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		fmt.Println("Create config file failed")
		return
	}
	t.Execute(out, PackageConfig{Name: name})

	fmt.Println("A new project created.")
}

func getExecutableDir() (exPath string, err error) {
	ex, err := os.Executable()
	if err != nil {
		return
	}
	exPath = filepath.Dir(ex)
	fmt.Println("exe path", exPath)
	exPath, err = filepath.EvalSymlinks(exPath)
	fmt.Println("after eval symbol links:", exPath)
	return
}
