package main

import (
	"fmt"
	"strings"

	"github.com/kovetskiy/lorg"
)

func getLogger(verbosity int) lorg.Logger {
	logger := lorg.NewLog()
	logger.SetFormat(lorg.NewFormat("${level:[%s]:left:true} %s"))
	if verbosity == 1 {
		logger.SetLevel(lorg.LevelDebug)
	} else if verbosity == 2 {
		logger.SetLevel(lorg.LevelTrace)
	}

	return logger
}

func fatalf(format string, values ...interface{}) {
	logger.Fatalf(wrapNewLines(format, values...))
}

func fatalln(value interface{}) {
	logger.Fatal(wrapNewLines("%s", value))
}

func debugf(format string, values ...interface{}) {
	logger.Debugf(wrapNewLines(format, values...))
}

func debugln(value interface{}) {
	logger.Debug(wrapNewLines("%s", value))
}

func tracef(format string, values ...interface{}) {
	logger.Trace(wrapNewLines(format, values...))
}

func wrapNewLines(format string, values ...interface{}) string {
	contents := fmt.Sprintf(format, values...)
	contents = strings.TrimSuffix(contents, "\n")
	contents = strings.Replace(
		contents,
		"\n",
		"\n"+strings.Repeat(" ", 8),
		-1,
	)

	return contents
}
