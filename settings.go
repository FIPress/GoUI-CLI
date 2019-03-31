package main

import (
	"html/template"
	"os"
)

const tmplS = `package main

goui.Platform = "{{.Platform}}"
`

func getPlatform() {

}

var tmpl, _ = template.New("").Parse(tmplS)

type settings struct {
	Platform string
	//Release int
}

const envFile = "goui-env.go"

func genSettings(st settings) {
	f, err := os.OpenFile(envFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		logger.Error("Create config file failed")
		return
	}
	tmpl.Execute(f, st)
}

func delSettings() {
	err := os.Remove(envFile)
	if err != nil {
		logger.Debug("remove settings file failed:", err)

	}
}
