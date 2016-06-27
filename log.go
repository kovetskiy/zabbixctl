package main

import "fmt"
import "github.com/kovetskiy/lorg"
import "github.com/kovetskiy/spinner-go"
import "os"

func getLogger() *lorg.Log {
	logger := lorg.NewLog()
	logger.SetFormat(lorg.NewFormat("${level:[%s]:left:true} %s"))

	return logger
}

func fatalf(format string, values ...interface{}) {
	if spinner.IsActive() {
		spinner.Stop()
	}

	fmt.Fprintf(os.Stderr, format+"\n", values...)
	os.Exit(1)
}

func fatalln(value interface{}) {
	fatalf("%s", value)
}

func debugf(format string, values ...interface{}) {
	logger.Debugf(format, values...)
}

func tracef(format string, values ...interface{}) {
	logger.Tracef(format, values...)
}

func debugln(value interface{}) {
	debugf("%s", value)
}

func traceln(value interface{}) {
	tracef("%s", value)
}
