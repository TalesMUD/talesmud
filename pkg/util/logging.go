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
		"logFile":  logFile,
		"maxAgeDays": 7,
		"level":    level.String(),
	}).Info("Logging configured (stderr + rotating file, 7-day retention)")
	return logFile
}

// RedactAccessToken strips secret query values from URLs for safe logging.
// access_token, code, and activate are removed. A key counts only at the
// start of the string or after ? or &, so user_code is left alone.
func RedactAccessToken(path string) string {
	if path == "" {
		return path
	}
	lower := strings.ToLower(path)
	var b strings.Builder
	b.Grow(len(path))
	for i := 0; i < len(path); {
		if n, ok := redactKeyAt(lower, i); ok {
			b.WriteString(path[i : i+n])
			b.WriteString("[REDACTED]")
			i += n
			for i < len(path) && path[i] != '&' && path[i] != ' ' {
				i++
			}
			continue
		}
		b.WriteByte(path[i])
		i++
	}
	return b.String()
}

func redactKeyAt(lower string, i int) (int, bool) {
	if i > 0 && lower[i-1] != '?' && lower[i-1] != '&' {
		return 0, false
	}
	for _, key := range []string{"access_token=", "activate=", "code="} {
		if strings.HasPrefix(lower[i:], key) {
			return len(key), true
		}
	}
	return 0, false
}
