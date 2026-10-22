package simple_logger

import (
	"fmt"
	"log"
)

const (
	reset  = "\033[0m"
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
	blue   = "\033[34m"
)

func logColor(color, prefix, format string, v ...any) {
	log.Print(color + prefix + fmt.Sprintf(format, v...) + reset)
}

func Successf(format string, v ...any) {
	logColor(green, "[SUCCESS] ", format, v...)
}

func Warnf(format string, v ...any) {
	logColor(yellow, "[WARN] ", format, v...)
}

func Errorf(format string, v ...any) {
	logColor(red, "[ERROR] ", format, v...)
}

func Debugf(isDev bool, format string, v ...any) {
	if isDev {
		logColor(blue, "[DEBUG] ", format, v...)
	}
}

func Infof(format string, v ...any) {
	logColor(reset, "[INFO] ", format, v...)
}

func Fatalf(format string, v ...any) {
	log.Fatalf(red+"[FATAL] "+format+reset, v...)
}
