// Package log aims to provide simple structured logging.
//
// The focus is on ease of use over performance.
//
// In practice, logs are commonly read by people so simple key-value logging is
// used to provide a happy medium between human readable and machine parsable.
package log

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

const (
	// PanicLevel level, highest level of severity. Logs and then calls panic.
	LevelPanic Level = iota
	// FatalLevel level. Logs and then calls `os.Exit(1)`.
	LevelFatal
	// ErrorLevel level. Logs. Used for errors that should definitely be noted.
	LevelError
	// WarnLevel level. Non-critical entries that deserve eyes.
	LevelWarn
	// InfoLevel level. General operational entries about what's going on inside the
	// application.
	LevelInfo
	// DebugLevel level. Usually only enabled when debugging. Very verbose logging.
	LevelDebug
)

// GlobalLevel is set at init() time using the `LOG_LEVEL` env variable.
// Misuse of this variable can lead to race conditions.
var GlobalLevel Level

func init() {
	GlobalLevel = stringToLevel(strings.ToLower(os.Getenv("LOG_LEVEL")))
	log.SetFlags(0)
}

// M is a convenience map type to prevent more typing (pun intended)
type M map[string]interface{}

func (kv M) String() string {
	var s string
	for k, v := range kv {
		s = fmt.Sprintf("%s %s", s, kvToString(k, v))
	}
	return s
}

type Level uint8

// Convert the Level to a string. E.g. PanicLevel becomes "panic".
func (level Level) String() string {
	switch level {
	case LevelDebug:
		return "debug"
	case LevelInfo:
		return "info"
	case LevelWarn:
		return "warning"
	case LevelError:
		return "error"
	case LevelFatal:
		return "fatal"
	case LevelPanic:
		return "panic"
	}

	return "unknown"
}

func stringToLevel(s string) Level {
	switch s {
	case "panic":
		return LevelPanic
	case "fatal":
		return LevelFatal
	case "error":
		return LevelError
	case "warn":
		return LevelWarn
	case "info":
		return LevelInfo
	default:
		return LevelDebug
	}
}

func Debug(msg interface{}, kvs ...M) {
	if GlobalLevel >= LevelDebug {
		prnt(LevelDebug, msg, kvs...)
	}
}

func Info(msg interface{}, kvs ...M) {
	if GlobalLevel >= LevelInfo {
		prnt(LevelInfo, msg, kvs...)

	}
}

func Warn(msg interface{}, kvs ...M) {
	if GlobalLevel >= LevelWarn {
		prnt(LevelWarn, msg, kvs...)
	}
}

func Error(msg interface{}, kvs ...M) {
	if GlobalLevel >= LevelError {
		prnt(LevelError, msg, kvs...)
	}
}

func Fatal(msg interface{}, kvs ...M) {
	if GlobalLevel >= LevelFatal {
		prnt(LevelFatal, msg, kvs...)
		os.Exit(1)
	}
}

func Panic(msg interface{}, kvs ...M) {
	if GlobalLevel >= LevelPanic {
		prnt(LevelPanic, msg, kvs...)
		panic(msg)
	}
}

func prnt(lvl Level, msg interface{}, kvs ...M) {
	var kvStr string
	for _, kv := range kvs {
		kvStr = kvStr + kv.String()
	}
	log.Printf("%s %s %s%s\n",
		kvToString("ts", time.Now().Format(time.RFC3339)),
		kvToString("lvl", lvl),
		kvToString("msg", msg),
		kvStr,
	)
}

func kvToString(k string, v interface{}) string {
	if len(k) == 0 {
		return ""
	}

	if s := fmt.Sprint(v); needsQuoting(s) {
		return fmt.Sprintf("%s=%q", k, s)
	}
	return fmt.Sprintf("%s=%v", k, v)
}

func needsQuoting(text string) bool {
	if len(text) == 0 {
		return true
	}
	for _, ch := range text {
		if !((ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') ||
			ch == '-' || ch == '.') {
			return true
		}
	}
	return false
}
