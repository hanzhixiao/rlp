package utils

import (
	"fmt"
	"io"
	"os"
	"runtime"
)

func Fatalf(format string, args ...interface{}) {
	w := io.MultiWriter(os.Stdout, os.Stderr)
	if runtime.GOOS == "windows" {
		w = os.Stdout
	} else {
		outinfo, _ := os.Stdout.Stat()
		errinfo, _ := os.Stderr.Stat()
		if outinfo != nil && errinfo != nil && outinfo == errinfo {
			w = os.Stderr
		}
	}
	fmt.Fprintf(w, "Fatal: "+format+"\n", args...)
	os.Exit(1)
}
