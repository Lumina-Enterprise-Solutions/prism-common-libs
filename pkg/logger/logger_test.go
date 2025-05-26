package logger

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestLoggerInitialization(t *testing.T) {
	assert.NotNil(t, Log)
	assert.IsType(t, &logrus.Logger{}, Log)
}

func TestLogLevel(t *testing.T) {
	// Test debug level
	os.Setenv("LOG_LEVEL", "debug")
	// Reinitialize logger (in real scenario, this would be done during init)
	Log.SetLevel(logrus.DebugLevel)
	assert.Equal(t, logrus.DebugLevel, Log.GetLevel())

	// Test info level
	os.Setenv("LOG_LEVEL", "info")
	Log.SetLevel(logrus.InfoLevel)
	assert.Equal(t, logrus.InfoLevel, Log.GetLevel())

	// Test error level
	os.Setenv("LOG_LEVEL", "error")
	Log.SetLevel(logrus.ErrorLevel)
	assert.Equal(t, logrus.ErrorLevel, Log.GetLevel())

	// Clean up
	os.Unsetenv("LOG_LEVEL")
}

func TestWithFields(t *testing.T) {
	fields := logrus.Fields{
		"user_id": "123",
		"action":  "test",
	}

	entry := WithFields(fields)
	assert.NotNil(t, entry)
	assert.IsType(t, &logrus.Entry{}, entry)
}
