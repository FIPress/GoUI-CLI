package main

import (
	"fmt"
	"github.com/fipress/fiputil"
	"github.com/fipress/go-rj"
	"os"
	"path/filepath"
	"strings"
)

const defaultWebDir = "web"

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
	platform     platformType
	platformDir  string
	appName      string
	outputDir    string
	tempDir      string
	srcPath      string
	manifestPath string
}

func (base *packagerBase) getPlatform() platformType {
	return base.platform
}

func (base *packagerBase) getPackageCfg(cfg interface{}) {
	dir := filepath.Join(base.workingDir, base.platform.String())
	_, err := os.Stat(dir)
	if err != nil {
		fiputil.CopyDir(filepath.Join(base.binDir, base.platform.String()), dir, nil)
	}

	cfgFile := filepath.Join(dir, "package.rj")

	err = rj.UnmarshalFile(cfgFile, cfg)
	if err != nil {
		logError("Unmarshalling the package.rj file for", base.platform, " failed:", err)
	}
}

func getPackager(ctx *context, platform string) (pkg packager, ok bool) {
	base := &packagerBase{
		context:  ctx,
		platform: platformFromString(platform),
	}

	//todo: parse -o from args instead of "build"
	base.outputDir = filepath.Join(base.workingDir, "build", base.platform.String())
	base.platformDir = filepath.Join(base.workingDir, base.platform.String())
	base.srcPath = filepath.Join(base.workingDir, base.platform.String())
	base.tempDir = filepath.Join(base.outputDir, "temp")
	base.appName = strings.ToLower(base.packageConfig.Name)

	switch base.platform {
	case android:
		pkg, ok = newAndroidPackager(base)
		break
	case iOS:
		pkg = newIOSPackager(base)
		break
	case macOS:
		pkg = newMacOSPackager(base)
		break
	case ubuntu:
		//ok = createUbuntu()
		break
	case windows:
		pkg, ok = newWindowsPackager(base)
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
	output string
	dir    string
	os     string
	arch   archType
	isProd bool
	//ccPath   string
	//cxxPath  string
	args []string
	env  []string
}

func (b *builder) addArg(s string) {
	b.args = append(b.args, s)
}

func (b *builder) addEnv(s string) {
	b.env = append(b.env, s)
}

func (b *builder) build() bool {
	//executable = filepath.Join(tempDir, packageConfig.Name)
	//genSettings(settings{Platform:platform})

	cmd := NewCommand("go", "build", "-v", "-o", b.output)
	cmd.Env = append(b.env, "CGO_ENABLED=1")
	cmd.Dir = b.dir
	//"GO111MODULE=off"？
	cmd.Args = append(cmd.Args, b.args...)
	if b.isProd {
		cmd.Args = append(cmd.Args, "-tags", "prod")
	} else {
		cmd.Args = append(cmd.Args, "-tags", "dev")
	}
	fmt.Println("env:", cmd.Env)
	err := cmd.Run()
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
