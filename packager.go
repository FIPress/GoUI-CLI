package main

import (
	"fmt"
	"github.com/fipress/fiputil"
	"github.com/fipress/fml"
	"os"
	"path/filepath"
)

const (
	darwin    = "darwin"
	archArm   = "arm"
	archAmd   = "amd64"
	assetDir  = "assets"
	plistFile = "Info.plist"
)

var (
	packageConfig PackageConfig
	buildDir      string
	tempDir       string
	executable    string
)

func build(args []string) {
	parsePackageConfig()
	var err error
	switch args[0] {
	case "ios":
		err = createIOSApp()
		break
	case "macos":
		err = createMacOSApp()
		break
	case "android":
		err = createApk()
		break
	case "ubuntu":
		err = createUbuntu()
		break
	case "windows":
		break
	}
	if err != nil {
		logger.Info("Package done")
	}
}

func parsePackageConfig() {
	cfg, err := fml.Load(confFile)
	if err != nil {
		fmt.Println("get config failed")
		return
	}
	packageConfig.Name = cfg.GetString("name", "GoUIApp")
}

func runBuildCmd(platform, platformOS, arch string) (err error) {
	buildDir = filepath.Join("build", platform)
	tempDir = filepath.Join(buildDir, "temp")
	executable = filepath.Join(tempDir, packageConfig.Name)
	//genSettings(settings{Platform:platform})

	cmd := NewCommand("go", "build", "-v", "-o", executable /*,"-ldflags=\"-extld=$CC\""*/)
	cmd.Env = []string{"GOARM=7", "GOOS=" + platformOS, "GOARCH=" + arch, "CGO_ENABLED=1"}

	err = cmd.Run(os.Stdout, os.Stderr, 0)
	//delSettings()
	if err != nil {
		logger.Error("build executable failed:", err)
	}
	return
}

func createApp(platform, arch string, copyFunc func()) (err error) {
	err = runBuildCmd(platform, darwin, arch)

	if err != nil {
		return
	}

	copyFunc()
	pkgPath := filepath.Join(buildDir, packageConfig.Name+".app")
	err = os.Rename(tempDir, pkgPath)
	if err != nil {
		logger.Error("Rename failed:", err)
	}
	logger.Info("Created package:", pkgPath)
	return
}

func createIOSApp() (err error) {
	//todo: other archs
	platform := "iOS"
	return createApp(platform, archArm, func() {
		fiputil.CopyDir("web", filepath.Join(tempDir, "web"), nil)
		fiputil.CopyFile(filepath.Join(assetDir, platform, plistFile), filepath.Join(tempDir, plistFile))
	})
}

func createMacOSApp() (err error) {
	//todo: other archs
	platform := "macOS"
	return createApp(platform, archAmd, func() {
		fiputil.CopyFile(filepath.Join(assetDir, platform, plistFile), filepath.Join(tempDir, plistFile))
		fiputil.CopyFile(executable, filepath.Join(tempDir, "MacOS", packageConfig.Name))
		fiputil.CopyDir("web", filepath.Join(tempDir, "Resources", "web"), nil)
	})

}

func createDmg() {

}

func createApk() (err error) {
	platform := "android"
	err = runBuildCmd(platform, platform, archArm)

	if err != nil {
		return
	}

	return
}

func createUbuntu() (err error) {
	platform := "linux"
	err = runBuildCmd(platform, platform, archAmd)

	if err != nil {
		return
	}

	return
}
