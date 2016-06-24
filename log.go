package main

import "fmt"
import "github.com/kovetskiy/lorg"
import "os"

func getLogger() *lorg.Log {
	logger := lorg.NewLog()
	logger.SetFormat(lorg.NewFormat("${level:[%s]:left:true} %s"))

	return logger
}

func fatalf(format string, values ...interface{}) {
	fmt.Fprintf(os.Stderr, format, values...)
	os.Exit(1)
}

func fatalln(value interface{}) {
	fmt.Fprintln(os.Stderr, value)
	os.Exit(1)
}

func debugf(format string, values ...interface{}) {
	logger.Debugf(format, values...)
}

func tracef(format string, values ...interface{}) {
	logger.Tracef(format, values...)
}

func debugln(value interface{}) {
	logger.Debug(value)
}

func traceln(value interface{}) {
	logger.Trace(value)
}
