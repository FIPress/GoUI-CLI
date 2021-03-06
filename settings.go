package main

import (
	"html/template"
	"os"
)

const tmplS = `package main

goui.Platform = "{{.Platform}}"
`

var tmpl, _ = template.New("").Parse(tmplS)

type settings struct {
	Platform string
	//Release int
}

const envFile = "goui-env.go"

func genSettings(st settings) {
	f, err := os.OpenFile(envFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		fatal("Create config file failed")
		return
	}
	tmpl.Execute(f, st)
}

func delSettings() {
	err := os.Remove(envFile)
	if err != nil {
		info("remove settings file failed:", err)

	}
}
