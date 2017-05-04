package logparse

import (
	"time"

	"github.com/upgear/gokit/log"
)

type Line struct {
	Time    time.Time
	Level   log.Level
	Message string
	Values  map[string]string
}

func ParseLine(ln string) Line {
	return Line{}
}
