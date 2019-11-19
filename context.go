package main

import (
	"fmt"
	"github.com/fipress/fml"
	"os"
	"path/filepath"
	"strings"
)

// taskType
type taskType int

const (
	create taskType = iota
	build
	help
)

var taskStrings = [...]string{"create", "build", "help"}

func taskFromString(s string) taskType {
	for k, v := range taskStrings {
		if v == s {
			return taskType(k)
		}
	}
	return help
}

// end of taskType

// platformType
type platformType int

const (
	android platformType = iota
	iOS
	macOS
	ubuntu
	windows
	unknown
)

var platformStrings = []string{"android", "ios", "macos", "ubuntu", "windows"}
var platformOSs = []string{"android", "darwin", "darwin", "linux", "windows"}
var platformArchs = [][]string{{""}, {"amd64", "amd"}, {"arm64", "arm"}, {""}, {"x64"}}

func platformFromString(s string) platformType {
	for k, v := range platformStrings {
		if v == strings.ToLower(s) {
			return platformType(k)
		}
	}
	return unknown
}

func (pt platformType) String() string {
	return platformStrings[pt]
}

func (pt platformType) OS() string {
	return platformOSs[pt]
}
func (pt platformType) Arch() []string {
	return platformArchs[pt]
}

// end of platformType

type PackageConfig struct {
	Name        string
	VersionCode string
	VersionName string
}

type context struct {
	//task taskType
	packageConfig PackageConfig
	workingDir    string
	binDir        string
	isProd        bool
}

func newContext() (c *context, ok bool) {
	c = new(context)

	c.binDir, ok = getBinDir()
	if !ok {
		return
	}

	var err error
	c.workingDir, err = os.Getwd()
	if err != nil {
		errorf("Get working directory failed: %s", err.Error())
		return
	}
	return c, true
}

func (c *context) loadConfig() {
	cfg, err := fml.Load(filepath.Join(c.workingDir, packageConfigFile))
	if err != nil {
		fmt.Println("get config failed")
		c.packageConfig = PackageConfig{Name: "GoUIApp", VersionCode: "1", VersionName: "1.0"}
	} else {
		c.packageConfig.Name = cfg.GetStringOrDefault("name", "GoUIApp")
		c.packageConfig.VersionCode = cfg.GetStringOrDefault("versionCode", "1")
		c.packageConfig.VersionName = cfg.GetStringOrDefault("versionName", "1.0")
	}

}

func getBinDir() (binDir string, ok bool) {
	dir, err := os.Executable()
	if err != nil {
		errorf("Get executable directory of GoUI-CLI failed: %s", err.Error())
		return
	}

	dir, err = filepath.EvalSymlinks(dir)

	if err != nil {
		errorf("Get executable directory of GoUI-CLI failed: %s", err.Error())
		return
	}

	binDir = filepath.Dir(dir)
	ok = true

	return
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
