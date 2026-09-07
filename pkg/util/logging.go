package util

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

// ConfigureLogging sets logrus level and dual output (stderr + rotating file).
// LOG_FILE defaults to logs/talesmud.log. Rotation keeps ~7 days (MaxAge).
// Path is documented in docs/development/LOGGING.md.
func ConfigureLogging() string {
	logLevel := strings.ToLower(strings.TrimSpace(os.Getenv("LOG_LEVEL")))
	if logLevel == "" {
		logLevel = "info"
	}
	level, err := log.ParseLevel(logLevel)
	if err != nil {
		log.WithField("LOG_LEVEL", logLevel).Warn("Invalid LOG_LEVEL, defaulting to info")
		level = log.InfoLevel
	}
	log.SetLevel(level)
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	logFile := strings.TrimSpace(os.Getenv("LOG_FILE"))
	if logFile == "" {
		logFile = "logs/talesmud.log"
	}
	if err := os.MkdirAll(filepath.Dir(logFile), 0o755); err != nil {
		log.WithError(err).WithField("logFile", logFile).Warn("Could not create log directory; file logging disabled")
		return ""
	}

	rotator := &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    50, // megabytes
		MaxBackups: 14,
		MaxAge:     7, // days
		Compress:   true,
	}
	log.SetOutput(io.MultiWriter(os.Stderr, rotator))
	log.WithFields(log.Fields{
		"logFile":    logFile,
		"maxAgeDays": 7,
		"level":      level.String(),
	}).Info("Logging configured (stderr + rotating file, 7-day retention)")
	return logFile
}

// RequestPathWithoutQuery returns the URL path with any query string removed.
func RequestPathWithoutQuery(path string) string {
	if i := strings.Index(path, "?"); i >= 0 {
		return path[:i]
	}
	return path
}

// RedactAccessToken strips access_token and ticket query values from URLs for safe logging.
func RedactAccessToken(path string) string {
	if path == "" {
		return path
	}
	out := path
	for _, key := range []string{"access_token=", "ticket="} {
		out = redactQueryValue(out, key)
	}
	return out
}

func redactQueryValue(path, key string) string {
	lower := strings.ToLower(path)
	keyLower := strings.ToLower(key)
	idx := strings.Index(lower, keyLower)
	if idx < 0 {
		return path
	}
	start := idx + len(key)
	end := start
	for end < len(path) && path[end] != '&' && path[end] != ' ' {
		end++
	}
	return path[:start] + "[REDACTED]" + path[end:]
}
