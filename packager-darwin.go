package main

import (
	"github.com/fipress/fiputil"
	"os"
	"path/filepath"
)

const (
	assetDir  = "assets"
	plistFile = "Info.plist"
)

type darwinPackager struct {
	*packagerBase
	tempDir string
}

func newDarwinPackager(base *packagerBase) *darwinPackager {
	return &darwinPackager{base, filepath.Join(base.outputDir, "temp")}
}

func (dp *darwinPackager) createDarwinApp(arch string, copyFunc func()) {
	err := dp.runBuildCmd(arch)

	if err != nil {
		errorf("build failed: %s.", err.Error())
		return
	}

	copyFunc()
	pkgPath := filepath.Join(dp.outputDir, dp.packageConfig.Name+".app")
	err = os.Rename(dp.tempDir, pkgPath)
	if err != nil {
		errorf("Rename failed: %s", err)
	}
	//logger.Info("Created package:", pkgPath)
	return
}

func (dp *darwinPackager) runBuildCmd(arch string) (err error) {
	dp.tempDir = filepath.Join(dp.outputDir, "temp")
	executable = filepath.Join(dp.tempDir, dp.packageConfig.Name)
	//genSettings(settings{Platform:platform})

	cmd := NewCommand("go", "build", "-v", "-o", executable /*,"-ldflags=\"-extld=$CC\""*/)
	cmd.Env = []string{"GOARM=7", "GOOS=" + dp.platform.OS(), "GOARCH=" + arch, "CGO_ENABLED=1"}

	err = cmd.Run()
	//delSettings()
	if err != nil {
		fatal("build executable failed:", err)
	}
	return
}

type iOSPackager struct {
	*darwinPackager
}

func newIOSPackager(base *packagerBase) *iOSPackager {
	return &iOSPackager{newDarwinPackager(base)}
}

func (ip *iOSPackager) create() {
	//todo: other archs
	arch := ip.platform.Arch()[0]
	if simulator {
		arch = macOS.Arch()[0]
	}
	ip.createDarwinApp(arch, func() {
		fiputil.CopyDir("web", filepath.Join(ip.tempDir, "web"), nil)
		fiputil.CopyFile(filepath.Join(assetDir, ip.platform.String(), plistFile), filepath.Join(ip.tempDir, plistFile))
	})
}

type macOSPackager struct {
	*darwinPackager
}

func newMacOSPackager(base *packagerBase) *macOSPackager {
	return &macOSPackager{newDarwinPackager(base)}
}

func (mp *macOSPackager) create() {
	//todo: other archs
	mp.createDarwinApp(mp.platform.Arch()[0], func() {
		fiputil.CopyFile(filepath.Join(assetDir, mp.platform.String(), plistFile), filepath.Join(mp.tempDir, plistFile))
		fiputil.CopyFile(executable, filepath.Join(mp.tempDir, "MacOS", mp.packageConfig.Name))
		fiputil.CopyDir("web", filepath.Join(mp.tempDir, "Resources", "web"), nil)
	})

}

func (mp *macOSPackager) createDmg() {

}
