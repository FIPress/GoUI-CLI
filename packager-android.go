package main

import (
	"github.com/fipress/fiputil"
	"github.com/fipress/fml"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
)

const (
	manifestFile    = "AndroidManifest.xml"
	unaligned       = "-unaligned.apk"
	unsigned        = "-unsigned.apk"
	androidConfTmpl = `
sdkPath={{.SdkPath}}
ndkPath={{.NdkPath}}
compileSdkVersion={{.SdkVersion}}
minSdkVersion={{.MinSdkVersion}}
`
)

type androidConfig struct {
	SdkPath           string
	SdkVersion        string
	NdkPath           string
	CompileSdkVersion string
}

// androidToolchain
type androidToolchain struct {
	abi         string
	toolPrefix  string
	clangPrefix string
}

func (at androidToolchain) getToolPath(ndkPath, name string) string {
	prefix := at.toolPrefix
	if strings.HasPrefix(name, "clang") {
		prefix = at.clangPrefix
	}
	return filepath.Join(ndkPath, "toolchains", "llvm", "prebuilt", ndkArch(), "bin", prefix+"-"+name)
}

func ndkArch() string {
	if runtime.GOOS == "windows" && runtime.GOARCH == "386" {
		return "windows"
	} else {
		var arch string
		switch runtime.GOARCH {
		case "386":
			arch = "x86"
		case "amd64":
			arch = "x86_64"
		default:
			panic("unsupported GOARCH: " + runtime.GOARCH)
		}
		return runtime.GOOS + "-" + arch
	}
}

// end of androidToolchain

type androidPackager struct {
	*packagerBase
	androidCfg     *androidConfig
	androidJar     string
	buildToolsPath string
	apkName        string
	srcPath        string
	tempDir        string
	manifestPath   string
	extraFiles     []string
}

func newAndroidPackager(base *packagerBase) (packager, bool) {
	return &androidPackager{packagerBase: base,
		srcPath: filepath.Join(base.workingDir, base.platform.String()),
		tempDir: filepath.Join(base.outputDir, "temp")}, true
}

func (ap *androidPackager) create() {
	_, err := os.Stat(ap.tempDir)
	if err == nil {
		os.Remove(ap.tempDir)
	}
	os.MkdirAll(ap.tempDir, 0766)

	ap.apkName = strings.ToLower(ap.packageConfig.Name)

	steps := []func() bool{ap.getAndroidConfig,
		ap.writeManifestXML,
		ap.copyAssets,
		ap.linkApk,
		ap.buildJava,
		ap.buildLib,
		ap.addFiles,
		ap.zipalign,
		ap.sign,
		ap.emulate,
	}
	for _, step := range steps {
		if !step() {
			return
		}
	}
}

func (ap *androidPackager) getAndroidConfig() bool {
	androidDir := filepath.Join(ap.workingDir, ap.platform.String())
	_, err := os.Stat(androidDir)
	if err != nil {
		fiputil.CopyDir(filepath.Join(ap.binDir, ap.platform.String()), androidDir, nil)
	}
	cfgFile := filepath.Join(ap.workingDir, ap.platform.String(), "androidPackage.conf")
	cfg := new(androidConfig)
	err = fml.UnmarshalFile(cfgFile, cfg)
	ap.androidCfg = cfg
	if err != nil {
		cfg.SdkPath = os.Getenv("ANDROID_HOME")

		if cfg.SdkPath == "" {
			errorf("Please install Android SDK, and set ANDROID_HOME to the root of it.")
			return false
		}

		cfg.NdkPath = filepath.Join(cfg.SdkPath, "ndk-bundle")

		buildToolsDir := filepath.Join(cfg.SdkPath, "build-tools")
		children, err := ioutil.ReadDir(buildToolsDir)
		if err != nil {
			errorf("Get android build tools failed, %s.", err.Error())
			return false
		}

		for i := len(children) - 1; i >= 0; i-- {
			name := children[i].Name()
			if fiputil.IsHidden(name) || strings.Contains(name, "rc") {
				continue
			}
			cfg.CompileSdkVersion = name
			break
		}

		if cfg.CompileSdkVersion == "" {
			errorf("Get android build tools failed.")
			return false
		}

		t := template.Must(template.New("config").Parse(androidConfTmpl))
		dest, err := os.OpenFile(cfgFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
		if err != nil {
			errorf("Create android config file failed")
		} else {
			t.Execute(dest, cfg)
		}
	}

	cfg.SdkVersion = strings.Split(cfg.CompileSdkVersion, ".")[0]

	_, err = os.Stat(cfg.NdkPath)
	if err != nil {
		errorf("Please install Android NDK with 'sdkmanager', or manually install it in $ANDROID_HOME/ndk-bundle.")
		return false
	}

	ap.buildToolsPath = filepath.Join(cfg.SdkPath, "build-tools", cfg.CompileSdkVersion)

	platformPath := filepath.Join(cfg.SdkPath, "platforms")
	children, err := ioutil.ReadDir(platformPath)
	if len(children) == 0 {
		errorf("Please add Android platform with 'sdkmanager', or manually install one.")
		return false
	}

	platform := ""
	for _, child := range children {
		if fiputil.IsHidden(child.Name()) {
			continue
		}
		platform = child.Name()
		break
	}

	//use the lowest version for compatibility
	ap.androidJar = filepath.Join(platformPath, platform, "android.jar")
	//ap.adb = filepath.Join(cfg.SdkPath,"platform-tools","adb")

	_, err = os.Stat(ap.buildToolsPath)
	if err != nil {
		errorf("Please install android build tools with 'sdkmanager'.")
		return false
	}

	return true
}

func (ap *androidPackager) getBuildTool(tool string) (path string) {
	path = filepath.Join(ap.buildToolsPath, tool)
	return getTool(path)
}

func getTool(tool string) (path string) {
	path = tool
	if runtime.GOOS == "windows" {
		path += ".exe"
	}
	return
}

func (ap *androidPackager) buildLib() bool {
	supported := []archType{a386, amd64, arm, arm64}
	for _, arch := range supported {
		b, ok := ap.newAndroidBuilder(arch)
		if !ok {
			return false
		}

		b.addArg("-buildmode=c-shared")
		ok = b.build()
		if !ok {
			return false
		}
	}

	return true
}

func (ap *androidPackager) newAndroidBuilder(arch archType) (b *builder, ok bool) {
	toolchain := arch.androidToolchain()
	clangPath := toolchain.getToolPath(ap.androidCfg.NdkPath, "clang")

	_, err := os.Stat(clangPath)
	if err != nil {
		errorf("Get clang for %s failed, please check if your android NDK is installed properly.", arch.String())
		return
	}

	libPath := "lib/" + toolchain.abi + "/libgoui.so"
	//output := filepath.Join(ap.tempDir, libPath)
	output := filepath.Join(ap.tempDir, "lib", toolchain.abi, "libgoui.so")

	b = &builder{
		arch:          arch,
		output:        output,
		clangPath:     clangPath,
		clangPlusPath: toolchain.getToolPath(ap.androidCfg.NdkPath, "clang++")}
	ok = true
	ap.extraFiles = append(ap.extraFiles, libPath)
	return
}

func (ap *androidPackager) buildJava() bool {
	dexName := "classes.dex"

	clsPath := filepath.Join(ap.tempDir, "classes")
	_, err := os.Stat(clsPath)
	if err != nil {
		os.MkdirAll(clsPath, 0755)
	}
	srcPath := filepath.Join(ap.workingDir, ap.platform.String(), "java")
	cmd := NewCommand("javac",
		"-classpath", ap.androidJar,
		"-sourcepath", srcPath,
		"-d", clsPath,
		filepath.Join(srcPath, "org", "fipress", "goui", "android", "GoUIActivity.java"),
	)

	err = cmd.Run(os.Stdout, os.Stderr, 0)
	if err != nil {
		errorf("Build java code failed: %s", err.Error())
		return false
	}

	/*sdk,err :=  strconv.Atoi(ap.androidCfg.SdkVersion)
	if err == nil && sdk >= 28 {
		cmd = NewCommand(ap.getBuildTool("d8"),
			filepath.Join(clsPath,"org","fipress","goui","android","*.class"),
			"--lib",ap.androidJar,
			"--output",ap.tempDir)
		//todo: --release
	} else {*/
	cmd = NewCommand(ap.getBuildTool("dx"),
		"--dex",
		"--output", filepath.Join(ap.tempDir, dexName),
		clsPath)
	//}

	err = cmd.Run(os.Stdout, os.Stderr, 0)
	if err != nil {
		errorf("Generate dex failed: %s", err.Error())
		return false
	}

	ap.extraFiles = append(ap.extraFiles, dexName)

	return true
}

type manifestConfig struct {
	PackageConfig
	SdkVersion string
	Debug      bool
}

func (ap *androidPackager) writeManifestXML() bool {
	manifestTemplPath := filepath.Join(ap.workingDir, ap.platform.String(), manifestFile)
	manifestTempl, err := template.New(manifestFile).ParseFiles(manifestTemplPath)

	if err != nil {
		errorf("Get %s failed, please make sure it exists and in the right place.", manifestFile)
		return false
	}

	ap.manifestPath = filepath.Join(ap.tempDir, manifestFile)
	file, err := os.Create(ap.manifestPath)
	if err != nil {
		errorf("Create manifest failed, %s", err.Error())
		return false
	}
	err = manifestTempl.Execute(file, manifestConfig{ap.packageConfig, ap.androidCfg.SdkVersion, true})
	if err != nil {
		errorf("Generate manifest failed, %s", err.Error())
		return false
	}

	return true
}

func (ap *androidPackager) copyAssets() bool {
	err := fiputil.CopyDir(filepath.Join(ap.workingDir, "web"), filepath.Join(ap.tempDir, "assets", "web"), nil)
	return err == nil
}

func (ap *androidPackager) linkApk() bool {
	aapt2 := ap.getBuildTool("aapt2")
	cmd := NewCommand(aapt2, "link",
		"-o", filepath.Join(ap.tempDir, ap.apkName+unaligned),
		"-I", ap.androidJar,
		"-A", filepath.Join(ap.tempDir, "assets"),
		"--manifest", ap.manifestPath,
	)

	err := cmd.Run(os.Stdout, os.Stderr, 0)
	if err != nil {
		errorf("Compile manifest failed: %s", err.Error())
		return false
	}
	return true
}

func (ap *androidPackager) addFiles() bool {
	aapt := ap.getBuildTool("aapt")
	cmd := NewCommand(aapt, "add",
		filepath.Join(ap.tempDir, ap.apkName+unaligned),
	)
	cmd.Args = append(cmd.Args, ap.extraFiles...)
	cmd.Dir = ap.tempDir
	err := cmd.Run(os.Stdout, os.Stderr, 0)
	if err != nil {
		errorf("aapt add files failed: %s", err.Error())
		return false
	}

	return true
}

func (ap *androidPackager) zipalign() bool {
	zipalign := ap.getBuildTool("zipalign")
	cmd := NewCommand(zipalign, "-f",
		"4",
		filepath.Join(ap.tempDir, ap.apkName+unaligned),
		filepath.Join(ap.outputDir, ap.apkName+unsigned),
	)
	err := cmd.Run(os.Stdout, os.Stderr, 0)
	if err != nil {
		errorf("zipalign apk failed: %s", err.Error())
		return false
	}

	return true
}

func (ap *androidPackager) sign() bool {
	apksigner := ap.getBuildTool("apksigner")
	//todo: key
	cmd := NewCommand(apksigner, "sign",
		"--ks", filepath.Join(ap.workingDir, ap.platform.String(), "key", "goui-debug.jks"),
		"--ks-pass", "pass:123456",
		"--out", filepath.Join(ap.outputDir, ap.apkName+".apk"),
		filepath.Join(ap.outputDir, ap.apkName+unsigned),
	)
	err := cmd.Run(os.Stdout, os.Stderr, 0)
	if err != nil {
		errorf("sign apk failed: %s", err.Error())
		return false
	}

	return true
}

const emulator = "emulator"
const appID = "org.fipress.goui.android"

func (ap *androidPackager) emulate() bool {
	/*
		emulator -list-avds
		emulator @avd_name [ {-option [value]} … ]
		emulatorPath := filepath.Join(ap.androidCfg.SdkPath,emulator,emulator)
		emulatorPath = getTool(emulatorPath)
		cmd := NewCommand(emulatorPath,"list",
			"--ks",filepath.Join(ap.workingDir,ap.platform.String(),"key","goui-debug.jks"),
			"--ks-pass","pass:123456",
			filepath.Join(ap.outputDir,ap.apkName+".apk"),
		)
		err := cmd.Run(os.Stdout, os.Stderr, 0)
		if err != nil {
			errorf("sign apk failed: %s", err.Error())
			return false
		}*/

	adbPath := filepath.Join(ap.androidCfg.SdkPath, "platform-tools", "adb")
	adb := getTool(adbPath)

	cmd := NewCommand(adb, "uninstall", appID)
	err := cmd.Run(os.Stdout, os.Stderr, 0)

	cmd = NewCommand(adb, "install", filepath.Join(ap.outputDir, ap.apkName+".apk"))
	err = cmd.Run(os.Stdout, os.Stderr, 0)
	if err != nil {
		errorf("install apk failed: %s", err.Error())
		return false
	}

	cmd = NewCommand(adb, "shell", "am", "start",
		"-a", "android.intent.action.MAIN",
		"-n", appID+"/.GoUIActivity")

	err = cmd.Run(os.Stdout, os.Stderr, 0)
	if err != nil {
		errorf("start app failed: %s", err.Error())
		return false
	}

	return true
}
