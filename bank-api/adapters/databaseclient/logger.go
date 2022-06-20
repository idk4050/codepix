package databaseclient

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/go-logr/logr"
	gormlogger "gorm.io/gorm/logger"
)

type gormLogger struct {
	Logger logr.Logger
	gormlogger.Config
}

func NewLogger(logger logr.Logger) gormlogger.Interface {
	config := gormlogger.Config{
		SlowThreshold:             200 * time.Millisecond,
		LogLevel:                  gormlogger.Info,
		IgnoreRecordNotFoundError: false,
	}
	return &gormLogger{
		Logger: logger,
		Config: config,
	}
}

func (gl *gormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	newLogger := *gl
	newLogger.LogLevel = level
	return &newLogger
}

func (gl gormLogger) Info(ctx context.Context, msg string, values ...interface{}) {
	if gl.LogLevel >= gormlogger.Info {
		gl.Logger.Info(fmt.Sprintf(msg, values...))
	}
}

func (gl gormLogger) Warn(ctx context.Context, msg string, values ...interface{}) {
	if gl.LogLevel >= gormlogger.Warn {
		gl.Logger.Info(fmt.Sprintf(msg, values...))
	}
}

func (gl gormLogger) Error(ctx context.Context, msg string, values ...interface{}) {
	if gl.LogLevel >= gormlogger.Error {
		gl.Logger.Error(fmt.Errorf(msg, values...), "")
	}
}

var ddlRegex = regexp.MustCompile(`.*?((?i:CREATE|ALTER|DROP|TRUNCATE) TABLE)`)
var verbRegex = regexp.MustCompile(`.*?(?i:INSERT|SELECT 1|SELECT COUNT|SELECT|UPDATE|DELETE)`)
var tableNameRegex = regexp.MustCompile(
	".*?(?i:(?i:CREATE|ALTER|DROP|TRUNCATE) TABLE|FROM|INTO|UPDATE)\\s`?([^` ]+)")

func (gl gormLogger) Trace(ctx context.Context, begin time.Time,
	fc func() (string, int64), err error,
) {
	if gl.LogLevel <= gormlogger.Silent {
		return
	}
	sql, rows := fc()
	kvs := []any{}

	ddl := ddlRegex.FindString(sql)
	if ddl != "" {
		kvs = append(kvs, "sql", sql)
	} else {
		verb := verbRegex.FindString(sql)
		if verb == "" {
			return
		}
		verb = strings.ToUpper(verb)
		kvs = append(kvs, "sql", verb)
	}

	table := tableNameRegex.FindStringSubmatch(sql)
	if len(table) == 0 {
		return
	}

	tableName := strings.ReplaceAll(table[1], `"`, "")
	elapsed := time.Since(begin)

	kvs = append(kvs,
		"table", tableName,
		"rows", rows,
		"elapsed", elapsed.String(),
	)
	switch {
	case err != nil && gl.LogLevel >= gormlogger.Error:
		if errors.Is(err, gormlogger.ErrRecordNotFound) {
			if !gl.IgnoreRecordNotFoundError {
				gl.Logger.Error(err, "trace error", kvs...)
			}
		} else {
			gl.Logger.Error(err, "trace error", kvs...)
		}
	case elapsed > gl.SlowThreshold && gl.SlowThreshold != 0 && gl.LogLevel >= gormlogger.Warn:
		gl.Logger.Info("trace warn: slow SQL", kvs...)
	case gl.LogLevel >= gormlogger.Info:
		gl.Logger.Info("trace", kvs...)
	}
}
