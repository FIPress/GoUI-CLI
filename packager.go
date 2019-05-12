package main

import (
	"os"
	"path/filepath"
	"strings"
)

// archType
type archType int

const (
	a386 archType = iota
	amd64
	arm
	arm64
)

var archStrings = []string{"386", "amd64", "arm", "arm64"}
var archAndroidToolchains = []androidToolchain{
	{
		abi:         "x86",
		toolPrefix:  "i686-linux-android",
		clangPrefix: "i686-linux-android16",
	},
	{
		abi:         "x86_64",
		toolPrefix:  "x86_64-linux-android",
		clangPrefix: "x86_64-linux-android21",
	},
	{
		abi:         "armeabi-v7a",
		toolPrefix:  "arm-linux-androideabi",
		clangPrefix: "armv7a-linux-androideabi16",
	},
	{
		abi:         "arm64-v8a",
		toolPrefix:  "aarch64-linux-android",
		clangPrefix: "aarch64-linux-android21",
	},
}

func archFromString(s string) archType {
	for k, v := range archStrings {
		if v == strings.ToLower(s) {
			return archType(k)
		}
	}
	return amd64
}

func (a archType) String() string {
	return archStrings[a]
}

func (a archType) androidToolchain() androidToolchain {
	return archAndroidToolchains[a]
}

// end of archType

type packager interface {
	create()
	getPlatform() platformType
}

type packagerBase struct {
	*context
	platform  platformType
	outputDir string
}

func (base *packagerBase) getPlatform() platformType {
	return base.platform
}

func getPackager(ctx *context, platform string) (pkger packager, ok bool) {
	base := &packagerBase{
		context:  ctx,
		platform: platformFromString(platform),
	}

	//todo: parse -o from args instead of "build"
	base.outputDir = filepath.Join(base.workingDir, "build", base.platform.String())

	switch base.platform {
	case android:
		pkger, ok = newAndroidPackager(base)
		break
	case iOS:
		pkger = newIOSPackager(base)
		break
	case macOS:
		pkger = newMacOSPackager(base)
		break
	case ubuntu:
		//ok = createUbuntu()
		break
	case windows:
		break
	}
	/*if ok {
		info("Packaging done")
	} else {
		info("Packaging failed")
	}*/
	return
}

func createApp() {

}

type builder struct {
	output        string
	platform      platformType
	arch          archType
	clangPath     string
	clangPlusPath string
	args          []string
	//env []string
}

func (b *builder) addArg(s string) {
	b.args = append(b.args, s)
}

/*
func (b builder) addEnv(s string)  {
	b.env = append(b.env,s)
}*/

func (b *builder) build() bool {
	//executable = filepath.Join(tempDir, packageConfig.Name)
	//genSettings(settings{Platform:platform})

	cmd := NewCommand("go", "build", "-v", "-o", b.output)
	cmd.Env = []string{
		"GOOS=" + b.platform.OS(),
		"GOARCH=" + b.arch.String(),
		"CC=" + b.clangPath,
		"CXX=" + b.clangPlusPath,
		"GO111MODULE=off",
		"CGO_ENABLED=1"}
	cmd.Args = append(cmd.Args, b.args...)

	if b.arch == arm {
		cmd.Env = append(cmd.Env, "GOARM=7")
	}

	err := cmd.Run(os.Stdout, os.Stderr, 0)
	//delSettings()
	if err != nil {
		errorf("build executable failed: %s", err.Error())
		return false
	}
	return true
}

/*
func createApk() (err error) {
	platform := "android"
	err = runBuildCmd(platform, platform, archArm)

	if err != nil {
		return
	}

	return
}*/

func createUbuntu() (err error) {
	/*platform := "linux"
	err = runBuildCmd(platform, platform, archAmd)

	if err != nil {
		return
	}*/

	return
}
