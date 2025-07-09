package logger

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

// LogLevel represents the severity of the log message
type LogLevel string

const (
	LevelDebug LogLevel = "DEBUG"
	LevelInfo  LogLevel = "INFO"
	LevelWarn  LogLevel = "WARN"
	LevelError LogLevel = "ERROR"
	LevelFatal LogLevel = "FATAL"
)

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp  time.Time              `json:"timestamp"`
	Level      LogLevel               `json:"level"`
	Message    string                 `json:"message"`
	Module     string                 `json:"module,omitempty"`
	Function   string                 `json:"function,omitempty"`
	File       string                 `json:"file,omitempty"`
	Line       int                    `json:"line,omitempty"`
	RequestID  string                 `json:"request_id,omitempty"`
	UserID     string                 `json:"user_id,omitempty"`
	Component  string                 `json:"component,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Error      string                 `json:"error,omitempty"`
	StackTrace string                 `json:"stack_trace,omitempty"`
}

// Logger represents the enhanced logger
type Logger struct {
	level        LogLevel
	enableJSON   bool
	enableCaller bool
	component    string
}

var (
	defaultLogger *Logger

	// Backward compatibility - keeping original loggers
	Info  *log.Logger
	Error *log.Logger
)

func init() {
	// Initialize enhanced logger
	defaultLogger = NewLogger()

	// Maintain backward compatibility
	Info = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	Error = log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
}

// NewLogger creates a new enhanced logger instance
func NewLogger() *Logger {
	level := LevelInfo
	if envLevel := os.Getenv("LOG_LEVEL"); envLevel != "" {
		level = LogLevel(strings.ToUpper(envLevel))
	}

	enableJSON := os.Getenv("LOG_FORMAT") == "json"
	enableCaller := os.Getenv("LOG_CALLER") == "true"

	return &Logger{
		level:        level,
		enableJSON:   enableJSON,
		enableCaller: enableCaller,
		component:    "iudx-sandbox-backend",
	}
}

// WithComponent creates a logger instance with a specific component name
func WithComponent(component string) *Logger {
	l := *defaultLogger
	l.component = component
	return &l
}

// shouldLog checks if a message should be logged based on the current level
func (l *Logger) shouldLog(level LogLevel) bool {
	levels := map[LogLevel]int{
		LevelDebug: 0,
		LevelInfo:  1,
		LevelWarn:  2,
		LevelError: 3,
		LevelFatal: 4,
	}
	return levels[level] >= levels[l.level]
}

// getCaller returns caller information
func (l *Logger) getCaller(skip int) (string, string, int) {
	if !l.enableCaller {
		return "", "", 0
	}

	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "", "", 0
	}

	fn := runtime.FuncForPC(pc)
	var funcName string
	if fn != nil {
		funcName = fn.Name()
		// Extract just the function name (not the full path)
		if idx := strings.LastIndex(funcName, "."); idx != -1 {
			funcName = funcName[idx+1:]
		}
	}

	// Extract just the filename (not the full path)
	if idx := strings.LastIndex(file, "/"); idx != -1 {
		file = file[idx+1:]
	}

	return funcName, file, line
}

// log is the core logging function
func (l *Logger) log(level LogLevel, ctx context.Context, msg string, metadata map[string]interface{}, err error) {
	if !l.shouldLog(level) {
		return
	}

	entry := LogEntry{
		Timestamp: time.Now().UTC(),
		Level:     level,
		Message:   msg,
		Component: l.component,
		Metadata:  metadata,
	}

	// Add caller information
	if l.enableCaller {
		funcName, file, line := l.getCaller(3)
		entry.Function = funcName
		entry.File = file
		entry.Line = line
	}

	// Add context information
	if ctx != nil {
		if requestID := ctx.Value("request_id"); requestID != nil {
			if id, ok := requestID.(string); ok {
				entry.RequestID = id
			}
		}
		if userID := ctx.Value("user_id"); userID != nil {
			if id, ok := userID.(string); ok {
				entry.UserID = id
			}
		}
	}

	// Add error information
	if err != nil {
		entry.Error = err.Error()
		if level == LevelError || level == LevelFatal {
			entry.StackTrace = getStackTrace()
		}
	}

	l.output(entry)
}

// getStackTrace returns a formatted stack trace
func getStackTrace() string {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

// output writes the log entry to the appropriate destination
func (l *Logger) output(entry LogEntry) {
	var output string

	if l.enableJSON {
		jsonBytes, err := json.Marshal(entry)
		if err != nil {
			// Fallback to simple format if JSON marshaling fails
			output = l.formatSimple(entry)
		} else {
			output = string(jsonBytes)
		}
	} else {
		output = l.formatSimple(entry)
	}

	// Write to appropriate stream
	writer := os.Stdout
	if entry.Level == LevelError || entry.Level == LevelFatal {
		writer = os.Stderr
	}

	writer.WriteString(output + "\n")
}

// formatSimple creates a human-readable log format
func (l *Logger) formatSimple(entry LogEntry) string {
	msg := entry.Timestamp.Format("2006-01-02 15:04:05") + " " + string(entry.Level) + " " + entry.Message

	if entry.Component != "" {
		msg += " [" + entry.Component + "]"
	}

	if entry.RequestID != "" {
		msg += " req_id=" + entry.RequestID
	}

	if entry.UserID != "" {
		msg += " user_id=" + entry.UserID
	}

	if entry.File != "" && entry.Line > 0 {
		msg += " " + entry.File + ":" + string(rune(entry.Line))
	}

	if entry.Error != "" {
		msg += " error=" + entry.Error
	}

	// Add metadata as JSON if present
	if len(entry.Metadata) > 0 {
		meta, err := json.Marshal(entry.Metadata)
		if err == nil {
			msg += " " + string(meta)
		}
	}

	return msg
}

// Public logging methods
func Debug(msg string) {
	defaultLogger.log(LevelDebug, nil, msg, nil, nil)
}

func DebugWithContext(ctx context.Context, msg string) {
	defaultLogger.log(LevelDebug, ctx, msg, nil, nil)
}

func DebugWithMetadata(msg string, metadata map[string]interface{}) {
	defaultLogger.log(LevelDebug, nil, msg, metadata, nil)
}

func InfoMsg(msg string) {
	defaultLogger.log(LevelInfo, nil, msg, nil, nil)
}

func InfoWithContext(ctx context.Context, msg string) {
	defaultLogger.log(LevelInfo, ctx, msg, nil, nil)
}

func InfoWithMetadata(msg string, metadata map[string]interface{}) {
	defaultLogger.log(LevelInfo, nil, msg, metadata, nil)
}

func Warn(msg string) {
	defaultLogger.log(LevelWarn, nil, msg, nil, nil)
}

func WarnWithContext(ctx context.Context, msg string) {
	defaultLogger.log(LevelWarn, ctx, msg, nil, nil)
}

func WarnWithError(msg string, err error) {
	defaultLogger.log(LevelWarn, nil, msg, nil, err)
}

func WarnWithMetadata(msg string, metadata map[string]interface{}) {
	defaultLogger.log(LevelWarn, nil, msg, metadata, nil)
}

func ErrorMsg(msg string) {
	defaultLogger.log(LevelError, nil, msg, nil, nil)
}

func ErrorWithContext(ctx context.Context, msg string) {
	defaultLogger.log(LevelError, ctx, msg, nil, nil)
}

func ErrorWithError(msg string, err error) {
	defaultLogger.log(LevelError, nil, msg, nil, err)
}

func ErrorWithMetadata(msg string, metadata map[string]interface{}, err error) {
	defaultLogger.log(LevelError, nil, msg, metadata, err)
}

func Fatal(msg string) {
	defaultLogger.log(LevelFatal, nil, msg, nil, nil)
	os.Exit(1)
}

func FatalWithError(msg string, err error) {
	defaultLogger.log(LevelFatal, nil, msg, nil, err)
	os.Exit(1)
}

// Audit logging for security events
func AuditLog(ctx context.Context, action, resource string, metadata map[string]interface{}) {
	auditData := map[string]interface{}{
		"audit":    true,
		"action":   action,
		"resource": resource,
	}

	// Merge with provided metadata
	for k, v := range metadata {
		auditData[k] = v
	}

	defaultLogger.log(LevelInfo, ctx, "AUDIT: "+action+" on "+resource, auditData, nil)
}

// Security logging for authentication/authorization events
func SecurityLog(ctx context.Context, event string, success bool, metadata map[string]interface{}) {
	securityData := map[string]interface{}{
		"security": true,
		"event":    event,
		"success":  success,
	}

	for k, v := range metadata {
		securityData[k] = v
	}

	level := LevelInfo
	if !success {
		level = LevelWarn
	}

	defaultLogger.log(level, ctx, "SECURITY: "+event, securityData, nil)
}
