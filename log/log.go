package log

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
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

func CreateContext() *context.Context {
	ctx := context.WithValue(context.Background(), "request", uuid.New().String())
	return &ctx
}

func Info(ctx *context.Context, code int64, message string) {
	doLog(ctx, "INFO", code, message)
}

func Debug(ctx *context.Context, code int64, message string) {
	doLog(ctx, "DEBUG", code, message)
}

func Warning(ctx *context.Context, code int64, message string) {
	doLog(ctx, "WARNING", code, message)
}

func Error(ctx *context.Context, code int64, message string) {
	doLog(ctx, "ERROR", code, message)
}

func Print(message string) {
	doLog(nil, "INFO", 0, message)
}

func doLog(ctx *context.Context, level string, code int64, message string) {
	var request string
	var user string
	if ctx != nil {
		request, _ = (*ctx).Value("request").(string)
		user, _ = (*ctx).Value("user").(string)
	}

	_, filename, line, _ := runtime.Caller(2)
	lastSlash := strings.LastIndex(filename, "/")
	filename = filename[lastSlash+1:]

	newLog := NewLog(request, user, level, filename, line, int(code), message)
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
	User    string `json:"user,omitempty"`
	Where   string `json:"where,omitempty"`
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

func NewLog(request, user, level, filename string, line int, code int, message string) *LogEntry {
	return &LogEntry{
		When:    time.Now().Format("2006-01-02 15:04:05"),
		Request: request,
		User:    user,
		Level:   level,
		Where:   filename + ":" + strconv.Itoa(line),
		Code:    code,
		Message: message,
	}
}
