// File: prism-common-libs/pkg/logger/logger.go
package logger

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"
	"time" // Tambahkan ini

	"github.com/gin-gonic/gin" // Tambahkan ini
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
	env := os.Getenv("ENVIRONMENT") // Anda mungkin memuat ini dari Vault sekarang
	if env == "production" || env == "prod" {
		Log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z",
			// FieldMap: logrus.FieldMap{ // Opsional: rename default fields
			//  logrus.FieldKeyTime: "@timestamp",
			//  logrus.FieldKeyMsg:  "message",
			// },
		})
	} else {
		Log.SetFormatter(&CustomFormatter{
			TimestampFormat: "15:04:05", // Format waktu di log dev
			NoColors:        noColors,
		})
	}

	// Set log level
	// Ini juga bisa diambil dari konfigurasi yang dimuat dari Vault
	logLevelEnv := os.Getenv("LOG_LEVEL")
	parsedLevel, err := logrus.ParseLevel(logLevelEnv)
	if err != nil {
		Log.SetLevel(logrus.InfoLevel) // Default jika parsing gagal
		// Log.Warnf("Invalid LOG_LEVEL '%s', defaulting to 'info'", logLevelEnv)
	} else {
		Log.SetLevel(parsedLevel)
	}
}

// GinLogger adalah middleware untuk logging request HTTP menggunakan logrus.
// Ini akan menggantikan logger default Gin.
func GinLogger(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		c.Next() // Proses request

		endTime := time.Now()
		latency := endTime.Sub(startTime)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		path := c.Request.URL.Path
		if c.Request.URL.RawQuery != "" {
			path = path + "?" + c.Request.URL.RawQuery
		}
		requestID := c.GetString("request_id") // Dari middleware RequestID Anda

		entry := logger.WithFields(logrus.Fields{
			"type":        "access_log", // Untuk membedakan dari log aplikasi biasa
			"status_code": statusCode,
			"latency_ms":  latency.Milliseconds(), // Latency dalam milidetik
			"client_ip":   clientIP,
			"method":      method,
			"path":        path,
			"request_id":  requestID, // Sertakan request ID jika ada
			// "user_agent": c.Request.UserAgent(), // Opsional
		})

		if len(c.Errors) > 0 {
			// Gabungkan error jika ada (Gin menyimpan error di c.Errors)
			entry.Error(c.Errors.ByType(gin.ErrorTypePrivate).String())
		} else {
			msg := fmt.Sprintf("%s %s %d (%s)", method, path, statusCode, latency)
			if statusCode >= http.StatusInternalServerError {
				entry.Error(msg)
			} else if statusCode >= http.StatusBadRequest {
				entry.Warn(msg)
			} else {
				entry.Info(msg)
			}
		}
	}
}

// GinRecovery adalah middleware untuk menangani panic dan log error menggunakan logrus.
// Ini akan menggantikan recovery default Gin.
func GinRecovery(logger *logrus.Logger, stack bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Periksa apakah koneksi rusak (broken pipe)
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				httpRequest, _ := httputil.DumpRequest(c.Request, false)
				requestID := c.GetString("request_id")

				if brokenPipe {
					logger.WithFields(logrus.Fields{
						"type":       "recovery_log",
						"request_id": requestID,
						"error":      err,
						"request":    string(httpRequest),
					}).Error("Recovery from broken pipe")
					// Jika koneksi rusak, kita tidak bisa mengirim response apa pun.
					// c.Error(err.(error)) // Jika Anda ingin Gin menangani ini lebih lanjut
					c.Abort()
					return
				}

				headers := ""
				if stack {
					headers = string(debug.Stack())
				}

				logger.WithFields(logrus.Fields{
					"type":       "panic_log",
					"request_id": requestID,
					"error":      err,
					"stack":      headers, // Sertakan stack trace jika diminta
					"request":    string(httpRequest),
				}).Error("[Recovery] panic recovered")

				// Kembalikan response error 500
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"message": "An internal server error occurred.",
					"error":   "internal_server_error", // Kode error generik
					// "request_id": requestID, // Bisa juga sertakan request ID di response
				})
			}
		}()
		c.Next()
	}
}

// Fungsi helper untuk parse level (digunakan di init atau jika Anda ingin mengubah level secara dinamis)
// func ParseLevel(lvl string) (logrus.Level, error) {
// 	return logrus.ParseLevel(lvl)
// }

// Fungsi helper untuk set level (jika Anda ingin mengubah level setelah init)
// func SetLevel(level logrus.Level) {
// 	Log.SetLevel(level)
// }

// --- Fungsi log yang sudah ada (WithFields, Debug, Info, dll.) tetap sama ---
// WithFields creates an entry with multiple fields
func WithFields(fields logrus.Fields) *logrus.Entry {
	return Log.WithFields(fields)
}

// WithField creates an entry with a single field
func WithField(key string, value interface{}) *logrus.Entry {
	return Log.WithField(key, value)
}

// WithError creates an entry with an error field
func WithError(err error) *logrus.Entry {
	return Log.WithError(err)
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

// Success logs a success message (using Info level with green color)
func Success(args ...interface{}) {
	entry := Log.WithField("type", "success")
	entry.Info(args...)
}

// Successf logs a formatted success message
func Successf(format string, args ...interface{}) {
	entry := Log.WithField("type", "success")
	entry.Infof(format, args...)
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
