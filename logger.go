package main

import "fmt"

//import "github.com/fipress/fiplog"

//var logger fiplog.Logger

/*func initLogger() {
	fiplog.InitWithConfig(&fiplog.Config{
		Level:   fiplog.LevelDebug,
		Pattern: "[%level] %msg",
	})
	logger = fiplog.GetLogger()
}*/
var verbose bool

func log(level string, a ...interface{}) {
	fmt.Print(level)
	fmt.Println(a...)
}

func debug(a ...interface{}) {
	if verbose {
		fmt.Println(a...)
	}
}

func info(a ...interface{}) {
	fmt.Println(a...)
}

func logError(a ...interface{}) {
	fmt.Print("[ERROR] ")
	fmt.Println(a...)
}

func fatal(a ...interface{}) {
	fmt.Print("[FATAL ERROR] ")
	fmt.Println(a...)
	fmt.Println("Aborting...")
}

func errorf(format string, a ...interface{}) {
	//fmt.Errorf(format,a...)
	fmt.Printf(format, a...)
}
