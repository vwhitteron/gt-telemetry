package utils

import (
	"fmt"
	"log"
)

type Logger struct {
	level int
}

var logLevel = map[string]int{
	"silent": 0,
	"error":  1,
	"warn":   2,
	"info":   3,
	"debug":  4,
}

func NewLogger(levelName string) (*Logger, error) {
	level, ok := logLevel[levelName]
	if !ok {
		return nil, fmt.Errorf("invalid log level %q", levelName)
	}

	return &Logger{
		level: level,
	}, nil
}

func (l *Logger) Debug(message string) {
	if l.level < logLevel["debug"] {
		return
	}

	log.Println(message)
}

func (l *Logger) Info(message string) {
	if l.level < logLevel["info"] {
		return
	}

	log.Println(message)
}

func (l *Logger) Warn(message string) {
	if l.level < logLevel["warn"] {
		return
	}

	log.Println(message)
}

func (l *Logger) Error(message string) {
	if l.level < logLevel["error"] {
		return
	}

	log.Println(message)
}
