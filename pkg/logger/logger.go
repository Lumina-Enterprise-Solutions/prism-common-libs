package logger

import (
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

var Log *logrus.Logger

// CustomFormatter provides a clean, colorful terminal output
type CustomFormatter struct {
	TimestampFormat string
	NoColors        bool
}

// Color constants for different log levels
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorGreen  = "\033[32m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
	ColorGray   = "\033[90m"
	ColorBold   = "\033[1m"
)

// Format renders a single log entry
func (f *CustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var levelColor string
	var levelIcon string
	
	// Check if it's a success message and add green color
	if !f.NoColors {
		if successType, exists := entry.Data["type"]; exists && successType == "success" {
			levelColor = ColorGreen
			levelIcon = "✅"
		}
	}

	if !f.NoColors {
		switch entry.Level {
		case logrus.DebugLevel:
			levelColor = ColorGray
			levelIcon = "🔍"
		case logrus.InfoLevel:
			levelColor = ColorBlue
			levelIcon = "ℹ️ "
		case logrus.WarnLevel:
			levelColor = ColorYellow
			levelIcon = "⚠️ "
		case logrus.ErrorLevel:
			levelColor = ColorRed
			levelIcon = "❌"
		case logrus.FatalLevel:
			levelColor = ColorRed + ColorBold
			levelIcon = "💀"
		default:
			levelColor = ColorWhite
			levelIcon = "📝"
		}
	}

	timestampFormat := f.TimestampFormat
	if timestampFormat == "" {
		timestampFormat = "2006-01-02 15:04:05"
	}

	var output strings.Builder
	
	// Timestamp
	if !f.NoColors {
		output.WriteString(ColorGray)
	}
	output.WriteString(entry.Time.Format(timestampFormat))
	if !f.NoColors {
		output.WriteString(ColorReset)
	}
	
	// Level with icon and color
	output.WriteString(" ")
	if !f.NoColors {
		output.WriteString(levelColor)
	}
	output.WriteString(levelIcon)
	output.WriteString(" ")
	output.WriteString(strings.ToUpper(entry.Level.String()))
	if !f.NoColors {
		output.WriteString(ColorReset)
	}
	
	// Message
	output.WriteString(" ")
	if !f.NoColors && entry.Level >= logrus.ErrorLevel {
		output.WriteString(ColorBold)
	}
	output.WriteString(entry.Message)
	if !f.NoColors && entry.Level >= logrus.ErrorLevel {
		output.WriteString(ColorReset)
	}
	
	// Fields (if any)
	if len(entry.Data) > 0 {
		output.WriteString(" ")
		if !f.NoColors {
			output.WriteString(ColorCyan)
		}
		output.WriteString("[")
		
		first := true
		for key, value := range entry.Data {
			if !first {
				output.WriteString(", ")
			}
			output.WriteString(fmt.Sprintf("%s=%v", key, value))
			first = false
		}
		
		output.WriteString("]")
		if !f.NoColors {
			output.WriteString(ColorReset)
		}
	}
	
	output.WriteString("\n")
	return []byte(output.String()), nil
}

//nolint:all
func init() {
	Log = logrus.New()
	Log.SetOutput(os.Stdout)
	
	// Check if we should disable colors (for CI/CD environments)
	noColors := os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb"
	
	// Use JSON formatter in production, custom formatter in development
	env := os.Getenv("ENVIRONMENT")
	if env == "production" || env == "prod" {
		Log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z",
		})
	} else {
		Log.SetFormatter(&CustomFormatter{
			TimestampFormat: "15:04:05",
			NoColors:        noColors,
		})
	}

	// Set log level
	level := os.Getenv("LOG_LEVEL")
	switch strings.ToLower(level) {
	case "debug":
		Log.SetLevel(logrus.DebugLevel)
	case "info":
		Log.SetLevel(logrus.InfoLevel)
	case "warn", "warning":
		Log.SetLevel(logrus.WarnLevel)
	case "error":
		Log.SetLevel(logrus.ErrorLevel)
	case "fatal":
		Log.SetLevel(logrus.FatalLevel)
	default:
		Log.SetLevel(logrus.InfoLevel)
	}
}

// WithFields creates an entry with multiple fields
func WithFields(fields logrus.Fields) *CustomEntry {
	return &CustomEntry{Log.WithFields(fields)}
}

// WithField creates an entry with a single field
func WithField(key string, value interface{}) *CustomEntry {
	return &CustomEntry{Log.WithField(key, value)}
}

// WithError creates an entry with an error field
func WithError(err error) *CustomEntry {
	return &CustomEntry{Log.WithError(err)}
}

// Debug logs a debug message
func Debug(args ...interface{}) {
	Log.Debug(args...)
}

// Debugf logs a formatted debug message
func Debugf(format string, args ...interface{}) {
	Log.Debugf(format, args...)
}

// Info logs an info message
func Info(args ...interface{}) {
	Log.Info(args...)
}

// Infof logs a formatted info message
func Infof(format string, args ...interface{}) {
	Log.Infof(format, args...)
}

// Warn logs a warning message
func Warn(args ...interface{}) {
	Log.Warn(args...)
}

// Warnf logs a formatted warning message
func Warnf(format string, args ...interface{}) {
	Log.Warnf(format, args...)
}

// Error logs an error message
func Error(args ...interface{}) {
	Log.Error(args...)
}

// Errorf logs a formatted error message
func Errorf(format string, args ...interface{}) {
	Log.Errorf(format, args...)
}

// Fatal logs a fatal message and exits
func Fatal(args ...interface{}) {
	Log.Fatal(args...)
}

// Fatalf logs a formatted fatal message and exits
func Fatalf(format string, args ...interface{}) {
	Log.Fatalf(format, args...)
}

// CustomEntry extends logrus.Entry with additional methods
type CustomEntry struct {
	*logrus.Entry
}

// Success logs a success message for CustomEntry
func (e *CustomEntry) Success(args ...interface{}) {
	e.WithField("type", "success").Info(args...)
}

// Successf logs a formatted success message for CustomEntry
func (e *CustomEntry) Successf(format string, args ...interface{}) {
	e.WithField("type", "success").Infof(format, args...)
}

// Success logs a success message (using Info level with green color)
func Success(args ...interface{}) {
	Log.WithField("type", "success").Info(args...)
}

// Successf logs a formatted success message
func Successf(format string, args ...interface{}) {
	Log.WithField("type", "success").Infof(format, args...)
}

// Banner prints a stylized banner message
func Banner(message string) {
	if Log.Level >= logrus.InfoLevel {
		banner := fmt.Sprintf(`
╔══════════════════════════════════════════════════════════════╗
║  %s  ║
╚══════════════════════════════════════════════════════════════╝`, 
		padString(message, 58))
		fmt.Println(ColorCyan + banner + ColorReset)
	}
}

// padString pads a string to the specified length
func padString(s string, length int) string {
	if len(s) >= length {
		return s[:length]
	}
	padding := (length - len(s)) / 2
	return strings.Repeat(" ", padding) + s + strings.Repeat(" ", length-len(s)-padding)
}