package main

import (
	"os"
)


var globalDefaultStdout = os.Stdout
var globalDefaultStderr = os.Stderr
var globalDefaultMsgFmt = "deploykit: service %s update image to %s"

func main() {
	os.Exit(run())
}
