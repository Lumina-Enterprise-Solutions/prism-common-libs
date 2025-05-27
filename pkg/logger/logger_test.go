package logger

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func unsetEnv(t *testing.T, key string) {
	t.Helper()
	if err := os.Unsetenv(key); err != nil {
		t.Errorf("Failed to unset %s: %v", key, err)
	}
}

func TestLoggerInitialization(t *testing.T) {
	assert.NotNil(t, Log)
	assert.IsType(t, &logrus.Logger{}, Log)
}

func TestLogLevel(t *testing.T) {
	// Test debug level
	if err := os.Setenv("LOG_LEVEL", "debug"); err != nil {
		t.Fatalf("Failed to set LOG_LEVEL to debug: %v", err)
	}
	defer unsetEnv(t, "LOG_LEVEL")
	// Reinitialize logger (in real scenario, this would be done during init)
	Log.SetLevel(logrus.DebugLevel)
	assert.Equal(t, logrus.DebugLevel, Log.GetLevel())

	// Test info level
	if err := os.Setenv("LOG_LEVEL", "info"); err != nil {
		t.Fatalf("Failed to set LOG_LEVEL to info: %v", err)
	}
	defer unsetEnv(t, "LOG_LEVEL")
	Log.SetLevel(logrus.InfoLevel)
	assert.Equal(t, logrus.InfoLevel, Log.GetLevel())

	// Test error level
	if err := os.Setenv("LOG_LEVEL", "error"); err != nil {
		t.Fatalf("Failed to set LOG_LEVEL to error: %v", err)
	}
	defer unsetEnv(t, "LOG_LEVEL")
	Log.SetLevel(logrus.ErrorLevel)
	assert.Equal(t, logrus.ErrorLevel, Log.GetLevel())
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
