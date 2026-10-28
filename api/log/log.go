// SPDX-License-Identifier: GPL-2.0-or-later
package log

import (
	"encoding/json"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gopkg.in/natefinch/lumberjack.v2"
)

func Initialize() {
	logFileLocation := os.Getenv("LOG_FILE_LOCATION")
	if logFileLocation != "" {
		log.SetOutput(&lumberjack.Logger{
			Filename:   logFileLocation,
			MaxSize:    100,
			MaxBackups: 5,
			MaxAge:     28,
			Compress:   true,
		})
	}
}

func Info(ctx *gin.Context, code int64, message string) {
	doLog(ctx, "INFO", code, message)
}

func Debug(ctx *gin.Context, code int64, message string) {
	doLog(ctx, "DEBUG", code, message)
}

func Warning(ctx *gin.Context, code int64, message string) {
	doLog(ctx, "WARNING", code, message)
}

func Error(ctx *gin.Context, code int64, message string) {
	doLog(ctx, "ERROR", code, message)
}

func Print(message string) {
	doLog(nil, "INFO", 0, message)
}

func doLog(request *gin.Context, level string, code int64, message string) {
	_, filename, line, _ := runtime.Caller(2)
	lastSlash := strings.LastIndex(filename, "/")
	filename = filename[lastSlash+1:]

	var r string
	if request != nil {
		r = request.GetString("request")
	}

	newLog := NewLog(r, level, filename, line, int(code), message)
	str, err := json.Marshal(newLog)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println(string(str))
	println(string(str))
}

type LogEntry struct {
	When    string `json:"when,omitempty"`
	Level   string `json:"level,omitempty"`
	Request string `json:"request,omitempty"`
	Where   string `json:"where,omitempty"`
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

func NewLog(request, level, filename string, line int, code int, message string) *LogEntry {
	return &LogEntry{
		When:    time.Now().Format("2006-01-02 15:04:05"),
		Request: request,
		Level:   level,
		Where:   filename + ":" + strconv.Itoa(line),
		Code:    code,
		Message: message,
	}
}
