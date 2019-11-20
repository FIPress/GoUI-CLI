package main

import (
	"fmt"
	"github.com/fipress/fiputil"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"text/template"
)

const (
	windowsManifestFile = "appxmanifest.xml"
	batFmt              = `@call %s
cl /EHsc /MT /favor:blend  /std:c++17 /await /c provider_windows.cpp  && link /dll provider_windows.obj /MACHINE:X64 /out:provider_windows.dll user32.lib`
	//windowsPackageFile    = "package.rj"
	//defaultWindowsKitPath = `C:\Program Files (x86)\Windows Kits\10\bin\`
	provider = "provider_windows"
	//packTool = "makeappx.exe"
	//unaligned           = "-unaligned.apk"
	//unsigned            = "-unsigned.apk"
	/*windowsConfTmpl = `
	windowsKitPath={{.WindowsKitPath}}
	targetArchitecture={{.Arch}}
	`*/
)

type windowsConfig struct {
	WindowsKitPath string
	VcVarsPath     string
}

type windowsPackager struct {
	*packagerBase
	windowsCfg     *windowsConfig
	windowsKitPath string
	dllFilename    string
	//manifestPath   string
	//extraFiles     []string
}

type windowsManifest struct {
	*packageConfig
	Id              string
	Publisher       string
	PublisherName   string
	Executable      string
	Desc            string
	Logo44          string
	Logo150         string
	Languages       []string
	BackgroundColor string
}

func newWindowsPackager(base *packagerBase) (packager, bool) {
	return &windowsPackager{packagerBase: base}, true
}

func (wp *windowsPackager) create() {
	_, err := os.Stat(wp.tempDir)
	if err == nil {
		os.Remove(wp.tempDir)
	}
	os.MkdirAll(wp.tempDir, 0766)

	steps := []func() bool{wp.getWindowsConfig,
		wp.writeManifestXML,
		wp.copyAssets,
		wp.buildProvider,
		wp.build,
		wp.pack,
	}
	for _, step := range steps {
		if !step() {
			return
		}
	}
}

func (wp *windowsPackager) getWindowsConfig() bool {
	wp.windowsCfg = new(windowsConfig)
	wp.getPlatformPackageCfg(wp.windowsCfg)

	if wp.windowsCfg.WindowsKitPath == "" || wp.windowsCfg.VcVarsPath == "" {
		fatal("Get windows packaging config failed. Please check the 'package.rj' file under the 'windows' directory of the project")
		return false
	}

	_, err := ioutil.ReadDir(wp.windowsCfg.WindowsKitPath)

	if err != nil {
		fatal("Get windows kit failed:", err.Error())
		return false
	}

	return true
}

func (wp *windowsPackager) buildProvider() bool {
	wp.dllFilename = path.Join(wp.platformDir, provider+".dll")
	cppFilename := path.Join(wp.platformDir, provider+".cpp")
	dllFile, errDll := os.Stat(wp.dllFilename)
	cppFile, errCpp := os.Stat(cppFilename)

	if errDll != nil {
		if errCpp != nil {
			fatal("Build windows provider failed, please make sure you have the 'provider_windows.dll' or 'provider_windows.cpp'")
			return false
		}
	} else {
		if errCpp != nil || cppFile.ModTime().Before(dllFile.ModTime()) {
			//no need to build if the cpp file not exists or not modified
			return true
		}
	}

	bat := path.Join(wp.platformDir, "build.bat")
	_, err := os.Stat(bat)
	if err != nil {
		//create the bat file
		if wp.windowsCfg.VcVarsPath == "" {
			fatal(`vcvarsPath not found in the config file. Please install Visual Studio VC++ tools, and config the path according to your tools version. It typically looks like 'C:\Program Files (x86)\Microsoft Visual Studio\2017\Enterprise\VC\Auxiliary\Build\vcvarsall.bat'.`)
			return false
		}

		f, err := os.Create(bat)
		if err != nil {
			fatal("Create build.bat failed")
			return false
		}
		_, err = fmt.Fprintf(f, batFmt, wp.windowsCfg.VcVarsPath)
		if err != nil {
			fatal("Write build.bat failed")
			return false
		}
	}

	// build
	cmd := NewCommand(bat)
	cmd.Dir = wp.platformDir
	err = cmd.Run()
	if err != nil {
		fatal("Build windows provider failed, error:", err)
		return false
	}
	return true
}

func (wp *windowsPackager) build() bool {
	executable := path.Join(wp.tempDir, wp.appName+".exe")
	b := builder{output: executable, isProd: wp.isProd}
	envStr := fmt.Sprintf(`CGO_LDFLAGS="-static %s"`, wp.dllFilename)
	debug("env:", envStr)
	b.addEnv(envStr)
	return b.build()
}

func (wp *windowsPackager) copyAssets() bool {
	err := fiputil.CopyDir(filepath.Join(wp.workingDir, defaultWebDir), filepath.Join(wp.tempDir, defaultWebDir), nil)
	return err == nil
}

func (wp *windowsPackager) pack() bool {
	//MakeAppx pack /v /h SHA256 /d "C:\My Files" /p MyPackage.msix
	cmd := NewCommand(wp.windowsKitPath + "\\x64\\MakeAppx.exe pack /v /d " + wp.tempDir + " /p " + wp.appName + ".msix")
	cmd.Dir = wp.outputDir
	err := cmd.Run()
	if err != nil {
		fatal("Pack windows package failed, error:", err)
		return false
	}
	return true
}

func (wp *windowsPackager) writeManifestXML() bool {
	manifestTemplPath := filepath.Join(wp.workingDir, wp.platform.String(), windowsManifestFile)
	manifestTempl, err := template.New(windowsManifestFile).ParseFiles(manifestTemplPath)

	if err != nil {
		errorf("Get %s failed, please make sure it exists and in the right place.", windowsManifestFile)
		return false
	}

	wp.manifestPath = filepath.Join(wp.tempDir, windowsManifestFile)
	file, err := os.Create(wp.manifestPath)
	if err != nil {
		errorf("Create manifest failed, %s", err.Error())
		return false
	}
	err = manifestTempl.Execute(file, windowsManifest{packageConfig: wp.packageConfig})
	if err != nil {
		errorf("Generate manifest failed, %s", err.Error())
		return false
	}

	return true
}
