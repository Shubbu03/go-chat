package pkg

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
	Caller    string                 `json:"caller,omitempty"`
	Error     string                 `json:"error,omitempty"`
}

type Logger struct {
	level      LogLevel
	structured bool
	logger     *log.Logger
}

var globalLogger *Logger

func InitLogger(level LogLevel, structured bool) {
	globalLogger = &Logger{
		level:      level,
		structured: structured,
		logger:     log.New(os.Stdout, "", 0),
	}
}

func GetLogger() *Logger {
	if globalLogger == nil {
		InitLogger(INFO, true)
	}
	return globalLogger
}

func (l *Logger) log(level LogLevel, message string, fields map[string]interface{}, err error) {
	if level < l.level {
		return
	}

	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     level.String(),
		Message:   message,
		Fields:    fields,
	}

	if level >= ERROR {
		if _, file, line, ok := runtime.Caller(3); ok {
			entry.Caller = fmt.Sprintf("%s:%d", getFileName(file), line)
		}
	}

	if err != nil {
		entry.Error = err.Error()
	}

	if l.structured {
		l.outputStructured(entry)
	} else {
		l.outputText(entry)
	}
}

func (l *Logger) outputStructured(entry LogEntry) {
	jsonData, err := json.Marshal(entry)
	if err != nil {
		l.logger.Printf("[%s] %s %s", entry.Level, entry.Timestamp, entry.Message)
		return
	}
	l.logger.Println(string(jsonData))
}

func (l *Logger) outputText(entry LogEntry) {
	var parts []string
	
	parts = append(parts, fmt.Sprintf("[%s] %s %s", entry.Level, entry.Timestamp, entry.Message))

	if len(entry.Fields) > 0 {
		fieldParts := make([]string, 0, len(entry.Fields))
		for key, value := range entry.Fields {
			fieldParts = append(fieldParts, fmt.Sprintf("%s=%v", key, value))
		}
		parts = append(parts, fmt.Sprintf("fields={%s}", strings.Join(fieldParts, ", ")))
	}

	if entry.Caller != "" {
		parts = append(parts, fmt.Sprintf("caller=%s", entry.Caller))
	}

	if entry.Error != "" {
		parts = append(parts, fmt.Sprintf("error=%s", entry.Error))
	}

	l.logger.Println(strings.Join(parts, " "))
}

func (l *Logger) Debug(message string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(DEBUG, message, f, nil)
}

func (l *Logger) Info(message string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(INFO, message, f, nil)
}

func (l *Logger) Warn(message string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(WARN, message, f, nil)
}

func (l *Logger) Error(message string, err error, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(ERROR, message, f, err)
}

func (l *Logger) Fatal(message string, err error, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(FATAL, message, f, err)
	os.Exit(1)
}

func Debug(message string, fields ...map[string]interface{}) {
	GetLogger().Debug(message, fields...)
}

func Info(message string, fields ...map[string]interface{}) {
	GetLogger().Info(message, fields...)
}

func Warn(message string, fields ...map[string]interface{}) {
	GetLogger().Warn(message, fields...)
}

func Error(message string, err error, fields ...map[string]interface{}) {
	GetLogger().Error(message, err, fields...)
}

func Fatal(message string, err error, fields ...map[string]interface{}) {
	GetLogger().Fatal(message, err, fields...)
}

func LogHTTPRequest(method, path, userAgent, clientIP string, userID uint, duration time.Duration) {
	Info("HTTP request", map[string]interface{}{
		"method":     method,
		"path":       path,
		"user_agent": userAgent,
		"client_ip":  clientIP,
		"user_id":    userID,
		"duration":   duration.String(),
	})
}

func LogHTTPError(method, path, clientIP string, statusCode int, err error, userID uint) {
	Error("HTTP error", err, map[string]interface{}{
		"method":      method,
		"path":        path,
		"client_ip":   clientIP,
		"status_code": statusCode,
		"user_id":     userID,
	})
}

func LogAuthEvent(event, email, clientIP string, userID uint, success bool) {
	level := INFO
	if !success {
		level = WARN
	}

	GetLogger().log(level, fmt.Sprintf("Auth event: %s", event), map[string]interface{}{
		"event":     event,
		"email":     email,
		"client_ip": clientIP,
		"user_id":   userID,
		"success":   success,
	}, nil)
}

func LogDatabaseOperation(operation, table string, duration time.Duration, err error) {
	if err != nil {
		Error("Database operation failed", err, map[string]interface{}{
			"operation": operation,
			"table":     table,
			"duration":  duration.String(),
		})
	} else {
		Debug("Database operation", map[string]interface{}{
			"operation": operation,
			"table":     table,
			"duration":  duration.String(),
		})
	}
}

func LogWebSocketEvent(event, clientIP string, userID uint, connectionID string) {
	Info("WebSocket event", map[string]interface{}{
		"event":         event,
		"client_ip":     clientIP,
		"user_id":       userID,
		"connection_id": connectionID,
	})
}

func LogRateLimitEvent(clientIP string, userID uint, endpoint string, exceeded bool) {
	level := DEBUG
	if exceeded {
		level = WARN
	}

	GetLogger().log(level, "Rate limit check", map[string]interface{}{
		"client_ip": clientIP,
		"user_id":   userID,
		"endpoint":  endpoint,
		"exceeded":  exceeded,
	}, nil)
}

func getFileName(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			return path[i+1:]
		}
	}
	return path
}

func SetLogLevel(level LogLevel) {
	if globalLogger != nil {
		globalLogger.level = level
	}
}

func SetStructuredLogging(enabled bool) {
	if globalLogger != nil {
		globalLogger.structured = enabled
	}
} 