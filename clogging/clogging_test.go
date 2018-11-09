package clogging_test

import (
	"cherrychain/clogging"
	"os"
	"testing"

	logging "github.com/op/go-logging"
)

const (
	CRITICAL = "CRITICAL"
	ERROR    = "ERROR"
	WARNING  = "WARNING"
	NOTICE   = "NOTICE"
	INFO     = "INFO"
	DEBUG    = "DEBUG"
)

func TestGetModuleLevelDefault(t *testing.T) {
	if clogging.DefaultLevel() != INFO {
		t.Fatal("Default level is not INFO")
	}
}

func TestSetModuleLevel(t *testing.T) {
	defer clogging.Reset()

	clogging.MustGetLogger("test")
	if level, err := clogging.SetModuleLevel("test", INFO); err == nil && level != INFO {
		t.Fatal("Default level is not INFO")
	}
	if level, err := clogging.SetModuleLevel("test", ERROR); err == nil && level != ERROR {
		t.Fatal("Default level is not ERROR")
	}
}

func TestGetModuleLevel(t *testing.T) {
	clogging.MustGetLogger("test")
	if level := clogging.GetModuleLevel("test"); level != INFO {
		t.Fatal("Default level is not ERROR")
	}
}

func TestSetLogLevel(t *testing.T) {
	clogging.MustGetLogger("test")
	clogging.SetLogLevel(ERROR)
	if level := clogging.GetModuleLevel("test"); level != ERROR {
		t.Fatal("Can not set level, level is not ERROR")
	}
}

func ExampleInitBackend() {
	logger := clogging.MustGetLogger("testModule")
	level, _ := logging.LogLevel(clogging.DefaultLevel())
	// initializes logging backend for testing and sets time to 1970-01-01 00:00:00.000 UTC
	logging.InitForTesting(level)

	formatSpec := "%{time:2006-01-02 15:04:05.000 MST} [%{module}] %{shortfunc} -> %{level:.4s} %{id:03x} %{message}"
	clogging.InitBackend(clogging.SetFormat(formatSpec), os.Stdout)

	logger.Info("test output")

	// Output:
	// 1970-01-01 00:00:00.000 UTC [testModule] ExampleInitBackend -> INFO 001 test output
}
