package log

import (
    golog "log"
    "fmt"
)

type LogLevel int
const (
    LogLevelDebug LogLevel = iota
    LogLevelInfo
    LogLevelWarn
    LogLevelError
    LogLevelDisabled
)

var Level LogLevel = LogLevelInfo

func doLog(level LogLevel, prefix string, s string, v ...any) {
    if Level >= level {
        golog.Printf("%s %s", prefix, fmt.Sprintf(s, v...))
    }
}

func Info(s string, v ...any) {
    doLog(LogLevelInfo, "[INFO]", s, v...)
}

func Debug(s string, v ...any) {
    doLog(LogLevelDebug, "[DEBUG]", s, v...)
}

func Warn(s string, v ...any) {
    doLog(LogLevelWarn, "[WARN]", s, v...)
}

func Error(s string, v ...any) {
    doLog(LogLevelError, "[ERROR]", s, v...)
}

func SetLevel(level LogLevel) {
    Level = level
}
