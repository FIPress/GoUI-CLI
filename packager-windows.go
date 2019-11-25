package main

import (
	"bytes"
	"fmt"
	"github.com/fipress/fiputil"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"text/template"
)

const (
	windowsManifestFile = "appxmanifest.xml"
	batFmt              = `@call "%s"
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
	CertFilename   string
}

type windowsPackager struct {
	*packagerBase
	*windowsConfig
	//windowsKitPath string

	dllFilename string
	Executable  string
	packageName string
	//manifestPath   string
	//extraFiles     []string
}

/*type windowsManifest struct {
	*packageConfig
	//Id              string
	//Publisher       string
	//PublisherName   string
	Executable      string
	//Desc            string
	//Logo44          string
	//Logo150         string
	//Languages       []string
	//BackgroundColor string
}*/

func newWindowsPackager(base *packagerBase) (packager, bool) {
	return &windowsPackager{packagerBase: base}, true
}

func (wp *windowsPackager) buildOnly() {
	steps := []func() bool{
		wp.getWindowsConfig,
		wp.clean,
		wp.copyAssets,
		wp.buildProvider,
		wp.build,
		wp.writeManifestXML,
	}

	wp.execute(steps)
}

func (wp *windowsPackager) packOnly() {
	steps := []func() bool{
		wp.getWindowsConfig,
		wp.writeManifestXML,
		wp.pack,
		wp.sign,
	}

	wp.execute(steps)
}

func (wp *windowsPackager) buildAndPack() {
	steps := []func() bool{
		wp.getWindowsConfig,
		wp.clean,
		wp.copyAssets,
		wp.buildProvider,
		wp.build,
		wp.writeManifestXML,
		wp.pack,
		wp.sign,
	}

	wp.execute(steps)
}

func (wp *windowsPackager) execute(steps []func() bool) {
	for _, step := range steps {
		if !step() {
			return
		}
	}
}

func (wp *windowsPackager) create() {
	_, err := os.Stat(wp.tempDir)
	if err == nil {
		os.Remove(wp.tempDir)
	}
	os.MkdirAll(wp.tempDir, 0766)

	steps := []func() bool{
		/*wp.clean,
		wp.copyAssets,
		wp.buildProvider,
		wp.build,*/
		wp.getWindowsConfig,
		wp.writeManifestXML,
		wp.pack,
		wp.sign,
	}
	for _, step := range steps {
		if !step() {
			return
		}
	}
}

func (wp *windowsPackager) clean() bool {
	_, err := os.Stat(wp.tempDir)
	if err == nil {
		err := os.RemoveAll(wp.outputDir)
		if err != nil {
			fatal("Clean output directory failed, error:", err)
			return false
		}
	}
	os.MkdirAll(wp.tempDir, 0766)

	return true
}

func (wp *windowsPackager) getWindowsConfig() bool {
	wp.windowsConfig = new(windowsConfig)
	wp.getPlatformConfig(wp.windowsConfig)

	if wp.WindowsKitPath == "" || wp.VcVarsPath == "" {
		fatal("Get windows packaging config failed. Please check the 'package.rj' file under the 'windows' directory of the project")
		return false
	}

	_, err := ioutil.ReadDir(wp.WindowsKitPath)

	if err != nil {
		fatal("Get windows kit failed:", err.Error())
		return false
	}

	wp.dllFilename = wp.platformDir + "\\" + provider + ".dll"
	wp.Executable = wp.appName + ".exe"
	wp.packageName = wp.appName + ".msix"

	return true
}

func (wp *windowsPackager) buildProvider() bool {
	cppFilename := wp.platformDir + "\\" + provider + ".cpp"
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
		if wp.VcVarsPath == "" {
			fatal(`vcvarsPath not found in the config file. Please install Visual Studio VC++ tools, and config the path according to your tools version. It typically looks like 'C:\Program Files (x86)\Microsoft Visual Studio\2017\Enterprise\VC\Auxiliary\Build\vcvarsall.bat'.`)
			return false
		}

		f, err := os.Create(bat)
		if err != nil {
			fatal("Create build.bat failed")
			return false
		}
		_, err = fmt.Fprintf(f, batFmt, wp.VcVarsPath)
		if err != nil {
			fatal("Write build.bat failed")
			return false
		}
		f.Close()
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
	//wp.dllFilename = `C:\mayunfeng\projects\go\src\github.com\fipress\demo\windows\provider_windows.dll`
	fiputil.CopyFile(wp.dllFilename, wp.tempDir+"\\"+provider+".dll")
	executable := path.Join(wp.tempDir, wp.Executable)
	b := builder{output: executable, isProd: wp.isProd}
	envStr := fmt.Sprintf(`CGO_LDFLAGS=-static %s`, wp.dllFilename)
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
	os.Remove(wp.packageName)

	cmd := NewCommand(wp.WindowsKitPath+"\\MakeAppx.exe",
		"pack", "/v", "/d", wp.tempDir, "/p", wp.packageName)
	cmd.Dir = wp.outputDir
	err := cmd.Run()

	if err != nil {
		fatal("Pack windows package failed, error:", err)
		return false
	}
	return true
}

func (wp *windowsPackager) sign() bool {
	cmd := NewCommand(wp.WindowsKitPath+`\signtool.exe`,
		`sign`, `/fd`, `SHA256`, `/t`, `http://timestamp.verisign.com/scripts/timestamp.dll`,
		`/a`, `/f`, wp.CertFilename, `/p`, `123456`, wp.packageName)
	cmd.Dir = wp.outputDir
	buf := new(bytes.Buffer)
	err := cmd.RunEx(buf, buf, 0)
	if err != nil {
		fatal("sing package failed, error:", err)
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
	defer file.Close()

	if err != nil {
		errorf("Create manifest failed, %s", err.Error())
		return false
	}

	//data := windowsManifest{packageConfig:wp.packageConfig}
	err = manifestTempl.Execute(file, wp)
	if err != nil {
		errorf("Generate manifest failed, %s", err.Error())
		return false
	}

	return true
}
