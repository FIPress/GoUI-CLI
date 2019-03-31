package main

import "github.com/fipress/fiplog"

var logger fiplog.Logger

func initLogger() {
	fiplog.InitWithConfig(&fiplog.Config{
		Level:   fiplog.LevelDebug,
		Pattern: "[%level] %msg",
	})
	logger = fiplog.GetLogger()
}
