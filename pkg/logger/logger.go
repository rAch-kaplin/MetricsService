package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

type LogLevel int

const (
	DebugLevel LogLevel = iota
	InfoLevel
	ErrorLevel
)

var levelMap = map[LogLevel]string{
	DebugLevel: "[DEBUG]",
	InfoLevel:  "[INFO]",
	ErrorLevel: "[ERROR]",
}

type logger struct {
	level LogLevel
	out   io.Writer
}

var (
	instance *logger
	once     sync.Once
)

func getLogger() *logger {
	once.Do(func() {
		instance = &logger{
			level: DebugLevel,
			out:   os.Stdout,
		}
	})
	return instance
}

func Init(level LogLevel, logFileName string) error {
	log := getLogger()
	log.level = level

	if logFileName != "" {
		file, err := os.Create(logFileName)
		if err != nil {
			return fmt.Errorf("failed to open log file %s: %w", logFileName, err)
		}
		log.out = file
	} else {
		log.out = os.Stdout
	}

	return nil
}

func Destroy() {
	log := getLogger()
	if file, ok := log.out.(*os.File); ok {
		file.Close()
	}
}

func (log *logger) log(level LogLevel, format string, args ...interface{}) {
	if log.level > level {
		return
	}

	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file, line = "---", 0
	}
	time := time.Now().Format("2006-01-02 15:04:05")
	levelStr := levelMap[level]
	msg := fmt.Sprintf(format, args...)

	fmt.Fprintf(log.out, "[%s]%s[%s:%d]: %s \n", time, levelStr, filepath.Base(file), line, msg)
}

func Debug(format string, args ...interface{}) {
	getLogger().log(DebugLevel, format, args...)
}

func Info(format string, args ...interface{}) {
	getLogger().log(InfoLevel, format, args...)
}

func Error(format string, args ...interface{}) {
	getLogger().log(ErrorLevel, format, args...)
}
