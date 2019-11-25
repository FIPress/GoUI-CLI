package main

import (
	"bytes"
	"fmt"
	"github.com/fipress/fiputil"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"html/template"
	"io/ioutil"
	"regexp"
	"testing"
)

func TestBuildProvider(t *testing.T) {
	//s:= `/c chcp 65001 & "C:\Program Files (x86)\Microsoft Visual Studio\2017\Enterprise\VC\Auxiliary\Build\vcvarsall.bat" x64&& cl /EHsc /MT /favor:blend  /std:c++17 /await /c C:\mayunfeng\projects\go\src\github.com\fipress\GoUI-cli\test/windows/provider_windows.h&& link /dll provider_windows.obj /MACHINE:X64 /out:C:\mayunfeng\projects\go\src\github.com\fipress\GoUI-cli\test/windows/provider_windows.dll"user32.lib"`
	//s:= `@call "%C:\Program Files (x86)\Microsoft Visual Studio\2017\Enterprise\VC\Auxiliary\Build\vcvarsall.bat" x64 %* && cl /EHsc /MT /favor:blend  /std:c++17 /await /c "C:\mayunfeng\projects\go\src\github.com\fipress\GoUI-cli\test\windows\provider_windows.cpp" && link /dll provider_windows.obj /MACHINE:X64 /out:"C:\mayunfeng\projects\go\src\github.com\fipress\GoUI-cli\test\windows\provider_windows.dll user32.lib`
	//c1 := `C:\Program Files (x86)\Microsoft Visual Studio\2017\Enterprise\VC\Auxiliary\Build\vcvarsall.bat`
	cmd := NewCommand(`C:\mayunfeng\projects\go\src\github.com\fipress\GoUI-cli\test\windows\build.bat`)
	cmd.Dir = `C:\mayunfeng\projects\go\src\github.com\fipress\GoUI-cli\test\windows\`
	buf := new(bytes.Buffer)
	err := cmd.RunEx(buf, buf, 0)
	t.Log(err)
	reader := transform.NewReader(buf, simplifiedchinese.GBK.NewDecoder())
	d, e := ioutil.ReadAll(reader)
	t.Log(string(d), e)
	fiputil.CopyFile(`C:\mayunfeng\projects\go\src\github.com\fipress\GoUI-cli\test\windows\provider_windows.dll`, `C:\mayunfeng\projects\go\src\github.com\fipress\GoUI-cli\test\build\windows\temp\provider_windows.dll`)
}

func TestBuild(t *testing.T) {
	executable := `C:\mayunfeng\projects\go\src\github.com\fipress\GoUI-cli\test\build\windows\temp\t1.exe`
	b := builder{output: executable, dir: `C:\mayunfeng\projects\go\src\github.com\fipress\GoUI-cli\test\`}
	envStr := fmt.Sprintf(`CGO_LDFLAGS=-static %s`, `C:\mayunfeng\projects\go\src\github.com\fipress\GoUI-cli\test\build\windows\temp\provider_windows.dll`)
	debug("env:", envStr)
	b.addEnv(envStr)
	b.build()
	/*cmd := NewCommand("go","build","-o", executable)
	cmd.Dir = `C:\mayunfeng\projects\go\src\github.com\fipress\GoUI-cli\test\`
	buf := new(bytes.Buffer)
	err := cmd.RunEx(buf,buf,0)

	if err != nil {
		fatal("Build exe for windows failed, error:", err)
	}
	reader:=transform.NewReader(buf,simplifiedchinese.GBK.NewDecoder())
	d,e := ioutil.ReadAll(reader)
	t.Log(string(d),e)*/
}

func TestPack(t *testing.T) {
	cmd := NewCommand(`C:\Program Files (x86)\Windows Kits\10\bin\10.0.17763.0\x64\MakeAppx.exe`, "pack", "/v", "/d",
		`C:\mayunfeng\projects\go\src\github.com\fipress\demo\build\windows\temp`,
		"/p", `C:\mayunfeng\projects\go\src\github.com\fipress\demo\build\windows\testdemo.msix`)
	//cmd.Dir = wp.outputDir
	buf := new(bytes.Buffer)
	err := cmd.RunEx(buf, buf, 0)

	if err != nil {
		fatal("Build exe for windows failed, error:", err)
	}
	reader := transform.NewReader(buf, simplifiedchinese.GBK.NewDecoder())
	d, e := ioutil.ReadAll(reader)
	t.Log(string(d), e)
}

//SignTool sign /fd <Hash Algorithm> /a /f <Path to Certificate>.pfx /p <Your Password> <File path>.msix
func TestSign(t *testing.T) {
	cmd := NewCommand(`C:\Program Files (x86)\Windows Kits\10\bin\10.0.17763.0\x64\signtool.exe`,
		`sign`, `/fd`, `SHA256`, `/t`, `http://timestamp.verisign.com/scripts/timestamp.dll`,
		`/a`, `/f`, `C:\dcn\c1.pfx`, `/p`, `123456`, `testdemo.msix`)
	cmd.Dir = `C:\mayunfeng\projects\go\src\github.com\fipress\GoUI-cli\test\build\windows\`
	buf := new(bytes.Buffer)
	err := cmd.RunEx(buf, buf, 0)
	t.Log(err)
	reader := transform.NewReader(buf, simplifiedchinese.GBK.NewDecoder())
	d, e := ioutil.ReadAll(reader)
	t.Log(string(d), e)
}

func TestCert(t *testing.T) {
	cmd := NewCommand(`powershell.exe`, "New-SelfSignedCertificate", "-Type", "Custom",
		"-Subject", `"CN=FIPress, O=FIP, L=Beijing, C=CN"`,
		"-KeyUsage", "DigitalSignature",
		"-FriendlyName", `"TF"`, "-CertStoreLocation",
		`"Cert:\CurrentUser\My"`,
		"-TextExtension", ` @("2.5.29.37={text}1.3.6.1.5.5.7.3.3", "2.5.29.19={text}")`)
	//cmd.Dir = wp.outputDir
	buf := new(bytes.Buffer)
	err := cmd.RunEx(buf, nil, 0)
	t.Log(err)
	regex := regexp.MustCompile(`\b[0-9a-fA-F]{40}\b`)
	c := regex.Find(buf.Bytes())
	t.Log("thumb:", c)
	reader := transform.NewReader(buf, simplifiedchinese.GBK.NewDecoder())
	d, e := ioutil.ReadAll(reader)
	t.Log(string(d), e)
	if err != nil {
		fatal("gen key failed, error:", err)
	}
	t.Log("OK")
}

func TestExtract(t *testing.T) {
	regex := regexp.MustCompile(`\b[0-9a-fA-F]{40}\b`)
	c := regex.FindString(`PSParentPath:Microsoft.PowerShell.Security\Certificate::CurrentUser\My

Thumbprint                                Subject                                                                      
----------                                -------                                                                      
46400A782BEA148E553FBF5B725C47D62EB17B78  CN=MyCompany, O=MyCompany, L=MyCity, S=MyState, C=MyCountry   `)
	t.Log("thumb:", c)
}

func TestExport(t *testing.T) {

	cmd := NewCommand(`powershell.exe`, `$pwd = ConvertTo-SecureString -String 123456 -Force -AsPlainText 
Export-PfxCertificate -cert "Cert:\CurrentUser\My\D3FDF04E7C87C1C07C14E1043A209F42C1179270" -FilePath C:\dcn\c1.pfx -Password $pwd`)
	//cmd.Dir = wp.outputDir
	buf := new(bytes.Buffer)
	err := cmd.RunEx(buf, buf, 0)
	t.Log(err)
	reader := transform.NewReader(buf, simplifiedchinese.GBK.NewDecoder())
	d, e := ioutil.ReadAll(reader)
	t.Log(string(d), e)
	if err != nil {
		fatal("gen key failed, error:", err)
	}
	t.Log("OK")

}

func TestTemp(t *testing.T) {
	type in struct {
		Name string
	}
	type s2 struct {
		*in
	}

	type s1 struct {
		*s2
	}
	type s struct {
		*s1
		Age int
	}
	tmpl, err := template.New(`a`).Parse(`name:{{.Name}}, age:{{.Age}}`)

	t.Log(err)

	b := new(bytes.Buffer)
	err = tmpl.Execute(b, s{&s1{&s2{&in{"A"}}}, 12})
	t.Log(err)
	t.Log(b.String())
}
