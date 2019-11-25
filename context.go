package main

import (
	"fmt"
	"github.com/fipress/fiputil"
	"github.com/fipress/go-rj"
	"os"
	"path/filepath"
	"strconv"
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

type packageConfig struct {
	Name        string
	VersionCode int
	VersionName string
}

type context struct {
	//task taskType
	*packageConfig
	workingDir string
	binDir     string
	isProd     bool
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

func createPackageConfig(filename string) (cfg *packageConfig) {
	cfg = &packageConfig{
		Name:        "GoUIApp Name",
		VersionCode: 1,
		VersionName: "0.0.1",
	}
	rj.MarshalToFile(cfg, filename)
	return
}

func (c *context) loadConfig() {
	cfg := new(packageConfig)
	filename := filepath.Join(c.workingDir, packageConfigFile)
	err := rj.UnmarshalFile(filename, cfg)
	if err != nil {
		cfg = createPackageConfig(filename)
	}

	//todo: if need to update version
	if c.isProd {
		cfg.VersionCode++

		codes := strings.Split(cfg.VersionName, ".")
		n := len(codes)
		str := codes[n-1]
		num, err := strconv.Atoi(str)
		if err != nil {
			logError("version name is not updatable")
			return
		}

		codes[n-1] = strconv.Itoa(num + 1)
		cfg.VersionName = fiputil.MkString(codes, "", ".", "")
		err = rj.MarshalToFile(cfg, filename)

		if err != nil {
			logError("Update version failed")
		}
	}
	c.packageConfig = cfg
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
