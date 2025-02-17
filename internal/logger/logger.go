package logger

import (
	log "unknwon.dev/clog/v2"
)

var noopLogFunction = func(format string, v ...interface{}) { /* noop */ }
var noopStop = func() { /* noop */ }

var (
	Trace = noopLogFunction
	Info  = noopLogFunction
	Warn  = noopLogFunction
	Error = noopLogFunction
	Fatal = noopLogFunction

	Stop = noopStop
)

func Setup(debug bool) {
	// Set the log level
	logLevel := log.LevelInfo
	if debug {
		logLevel = log.LevelTrace
	}

	// Create the logger
	err := log.NewConsole(log.ConsoleConfig{Level: logLevel})
	if err != nil {
		panic("unable to create new logger: " + err.Error())
	}

	// Export the log methods
	// We don't need to, as the global `log` alias is the logger we just created,
	// but it requires defining the alias all the time, and it can be easily
	// confused with the built-in `log`
	Trace = log.Trace
	Info = log.Info
	Warn = log.Warn
	Error = log.Error
	Fatal = log.Fatal
	Stop = log.Stop
}
