package main

import "testing"

func TestAndroidPackager_writeManifestXML(t *testing.T) {
	ap := &androidPackager{packagerBase: &packagerBase{platform: android, context: &context{workingDir: "/work/GOUI/Test/Hello424"}}}
	ok := ap.writeManifestXML()
	t.Log(ok)
}
