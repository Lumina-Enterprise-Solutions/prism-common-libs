package enhanced_logger

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Color definitions for different log levels
var (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorGreen  = "\033[32m"
	colorWhite  = "\033[37m"
	colorGray   = "\033[90m"
	colorBold   = "\033[1m"
	colorDim    = "\033[2m"
)

// Emoji mappings for log levels
var levelEmojis = map[zerolog.Level]string{
	zerolog.TraceLevel: "🔍",
	zerolog.DebugLevel: "🐛",
	zerolog.InfoLevel:  "ℹ️",
	zerolog.WarnLevel:  "⚠️",
	zerolog.ErrorLevel: "❌",
	zerolog.FatalLevel: "💀",
	zerolog.PanicLevel: "🚨",
}

// Service name colors for better visual distinction
var serviceColors = []string{
	"\033[94m", // Light Blue
	"\033[92m", // Light Green
	"\033[93m", // Light Yellow
	"\033[95m", // Light Magenta
	"\033[96m", // Light Cyan
	"\033[91m", // Light Red
}

// CustomConsoleWriter provides enhanced console output
type CustomConsoleWriter struct {
	Out        *os.File
	TimeFormat string
	NoColor    bool
}

// Write implements the io.Writer interface
func (w CustomConsoleWriter) Write(p []byte) (n int, err error) {
	if w.NoColor {
		return w.Out.Write(p)
	}

	// Parse the JSON log entry
	logData := make(map[string]interface{})
	if err := json.Unmarshal(p, &logData); err != nil {
		// Fallback to raw output if parsing fails
		return w.Out.Write(p)
	}

	formatted := w.formatLogEntry(logData)
	return w.Out.Write([]byte(formatted))
}

func (w CustomConsoleWriter) formatLogEntry(logData map[string]interface{}) string {
	var parts []string

	// 1. Timestamp with gradient effect
	if timestamp, ok := logData["time"].(string); ok {
		if parsedTime, err := time.Parse(time.RFC3339, timestamp); err == nil {
			timeStr := parsedTime.Format("15:04:05.000")
			parts = append(parts, fmt.Sprintf("%s%s%s%s", colorDim, colorCyan, timeStr, colorReset))
		}
	}

	// 2. Service name with distinctive color
	serviceName := "unknown"
	if service, ok := logData["service"].(string); ok {
		serviceName = service
	}
	serviceColor := w.getServiceColor(serviceName)
	serviceBox := fmt.Sprintf("%s[%s%s%s%s]%s", colorBold, serviceColor, serviceName, colorReset, colorBold, colorReset)
	parts = append(parts, serviceBox)

	// 3. Log level with emoji and color
	level := zerolog.InfoLevel
	if levelStr, ok := logData["level"].(string); ok {
		if parsedLevel, err := zerolog.ParseLevel(levelStr); err == nil {
			level = parsedLevel
		}
	}

	emoji := levelEmojis[level]
	levelColor := w.getLevelColor(level)
	levelText := strings.ToUpper(level.String())
	levelFormatted := fmt.Sprintf("%s %s%s%s%s%s", emoji, colorBold, levelColor, levelText, colorReset, colorReset)
	parts = append(parts, levelFormatted)

	// 4. Caller information (file:line)
	if caller, ok := logData["caller"].(string); ok {
		callerParts := strings.Split(caller, ":")
		if len(callerParts) >= 2 {
			filename := filepath.Base(callerParts[0])
			line := callerParts[1]
			callerInfo := fmt.Sprintf("%s%s%s:%s%s", colorGray, filename, colorWhite, line, colorReset)
			parts = append(parts, fmt.Sprintf("(%s)", callerInfo))
		}
	}

	// 5. Main message with proper formatting
	message := ""
	if msg, ok := logData["message"].(string); ok {
		message = msg
	} else if msg, ok := logData["msg"].(string); ok {
		message = msg
	}

	if message != "" {
		messageColor := w.getMessageColor(level)
		formattedMessage := fmt.Sprintf("%s%s%s", messageColor, message, colorReset)
		parts = append(parts, formattedMessage)
	}

	// 6. Additional fields formatting
	additionalFields := w.formatAdditionalFields(logData)
	if additionalFields != "" {
		parts = append(parts, additionalFields)
	}

	// 7. Error details (if present)
	if errorMsg, ok := logData["error"].(string); ok {
		errorFormatted := fmt.Sprintf("%s%s🔥 ERROR: %s%s", colorBold, colorRed, errorMsg, colorReset)
		parts = append(parts, errorFormatted)
	}

	// Join all parts and add decorative elements
	result := strings.Join(parts, " ")

	// Add decorative border for important logs
	if level >= zerolog.WarnLevel {
		border := w.getBorder(level)
		result = fmt.Sprintf("%s\n%s\n%s", border, result, border)
	}

	return result + "\n"
}

func (w CustomConsoleWriter) getServiceColor(serviceName string) string {
	// Generate consistent color based on service name hash
	hash := 0
	for _, char := range serviceName {
		hash += int(char)
	}
	return serviceColors[hash%len(serviceColors)]
}

func (w CustomConsoleWriter) getLevelColor(level zerolog.Level) string {
	switch level {
	case zerolog.TraceLevel:
		return colorPurple
	case zerolog.DebugLevel:
		return colorCyan
	case zerolog.InfoLevel:
		return colorGreen
	case zerolog.WarnLevel:
		return colorYellow
	case zerolog.ErrorLevel:
		return colorRed
	case zerolog.FatalLevel:
		return colorRed + colorBold
	case zerolog.PanicLevel:
		return colorRed + colorBold
	default:
		return colorWhite
	}
}

func (w CustomConsoleWriter) getMessageColor(level zerolog.Level) string {
	switch level {
	case zerolog.ErrorLevel, zerolog.FatalLevel, zerolog.PanicLevel:
		return colorRed
	case zerolog.WarnLevel:
		return colorYellow
	case zerolog.InfoLevel:
		return colorWhite
	default:
		return colorGray
	}
}

func (w CustomConsoleWriter) formatAdditionalFields(logData map[string]interface{}) string {
	var fields []string

	// Skip standard fields
	skipFields := map[string]bool{
		"time": true, "level": true, "message": true, "msg": true,
		"service": true, "caller": true, "error": true,
	}

	for key, value := range logData {
		if skipFields[key] {
			continue
		}

		// Format different types appropriately
		var formattedValue string
		switch v := value.(type) {
		case string:
			if key == "port" || key == "grpc_port" {
				formattedValue = fmt.Sprintf("%s%s:%s%s", colorCyan, key, colorWhite, v)
			} else {
				formattedValue = fmt.Sprintf("%s%s:%s%s", colorBlue, key, colorWhite, v)
			}
		case float64:
			if v == float64(int64(v)) {
				formattedValue = fmt.Sprintf("%s%s:%s%d", colorBlue, key, colorWhite, int64(v))
			} else {
				formattedValue = fmt.Sprintf("%s%s:%s%.2f", colorBlue, key, colorWhite, v)
			}
		case bool:
			boolColor := colorGreen
			if !v {
				boolColor = colorRed
			}
			formattedValue = fmt.Sprintf("%s%s:%s%s%t", colorBlue, key, boolColor, colorReset, v)
		default:
			formattedValue = fmt.Sprintf("%s%s:%s%v", colorBlue, key, colorWhite, v)
		}

		fields = append(fields, formattedValue+colorReset)
	}

	if len(fields) > 0 {
		return fmt.Sprintf("%s[%s]%s", colorGray, strings.Join(fields, " "), colorReset)
	}
	return ""
}

func (w CustomConsoleWriter) getBorder(level zerolog.Level) string {
	var symbol string
	var color string

	switch level {
	case zerolog.WarnLevel:
		symbol = "⚠"
		color = colorYellow
	case zerolog.ErrorLevel:
		symbol = "❌"
		color = colorRed
	case zerolog.FatalLevel, zerolog.PanicLevel:
		symbol = "💀"
		color = colorRed + colorBold
	default:
		return ""
	}

	border := strings.Repeat(symbol, 50)
	return fmt.Sprintf("%s%s%s%s", colorBold, color, border, colorReset)
}

// Enhanced Init function with better configuration
func Init() {
	logLevelStr := strings.ToLower(os.Getenv("LOG_LEVEL"))
	var level zerolog.Level

	switch logLevelStr {
	case "trace":
		level = zerolog.TraceLevel
	case "debug":
		level = zerolog.DebugLevel
	case "warn", "warning":
		level = zerolog.WarnLevel
	case "error":
		level = zerolog.ErrorLevel
	case "fatal":
		level = zerolog.FatalLevel
	case "panic":
		level = zerolog.PanicLevel
	default:
		level = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(level)

	// Enable caller information for better debugging
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		return filepath.Base(file) + ":" + strconv.Itoa(line)
	}

	// Configure the logger based on environment
	if isTTY() && !isDisableColor() {
		// Enhanced console output for TTY
		log.Logger = zerolog.New(CustomConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: time.RFC3339,
			NoColor:    false,
		}).With().Timestamp().Caller().Logger()

		// Print startup banner
		printStartupBanner()
	} else {
		// JSON output for production/Docker
		log.Logger = zerolog.New(os.Stdout).With().Timestamp().Caller().Logger()
	}
}

func printStartupBanner() {
	banner := fmt.Sprintf(`
%s%s╔══════════════════════════════════════════════════════════════════════╗%s
%s%s║                    🚀 PRISM MICROSERVICES PLATFORM 🚀                ║%s
%s%s║                         Enhanced Logging System                       ║%s
%s%s╚══════════════════════════════════════════════════════════════════════╝%s
`, colorBold, colorCyan, colorReset,
		colorBold, colorBlue, colorReset,
		colorBold, colorGreen, colorReset,
		colorBold, colorCyan, colorReset)

	fmt.Fprint(os.Stderr, banner)
}

func isTTY() bool {
	fileInfo, _ := os.Stdout.Stat()
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

func isDisableColor() bool {
	return os.Getenv("NO_COLOR") != "" || os.Getenv("DISABLE_COLOR") != ""
}

// Additional helper functions for enhanced logging
func WithService(serviceName string) zerolog.Logger {
	return log.With().Str("service", serviceName).Logger()
}

func WithRequestID(requestID string) zerolog.Logger {
	return log.With().Str("request_id", requestID).Logger()
}

func WithUserID(userID string) zerolog.Logger {
	return log.With().Str("user_id", userID).Logger()
}

func LogStartup(serviceName string, port int, additionalInfo map[string]interface{}) {
	logger := log.With().
		Str("service", serviceName).
		Int("port", port).
		Str("status", "starting").
		Logger()

	for key, value := range additionalInfo {
		logger = logger.With().Interface(key, value).Logger()
	}

	logger.Info().Msg("Service initialization completed successfully")
}

func LogShutdown(serviceName string) {
	l := log.With().
		Str("service", serviceName).
		Str("status", "shutdown").
		Logger()
	l.Info().Msg("Service shutdown completed gracefully")
}
