// Copyright 2018 Oracle and/or its affiliates. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logging

import (
	"os"
	"strings"
	"sync"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

var (
	lvl         = zapcore.InfoLevel
	logJSON     = false
	logfilePath = ""
	config      *zap.Config
	mu          sync.Mutex
)

// Options holds the zap logger configuration.
type Options struct {
	LogLevel *zapcore.Level
	Config   *zap.Config
}

// Level gets the current log level.
func Level() *zap.AtomicLevel {
	return &config.Level
}

// Logger builds a new logger based on the given flags.
func Logger() *zap.Logger {
	return logger(logfilePath)
}

// FileLogger builds a new logger which logs to the given path.
func FileLogger(path string) *zap.Logger {
	return logger(path)
}

func logger(path string) *zap.Logger {
	mu.Lock()
	defer mu.Unlock()

	var cfg zap.Config

	setFlags()

	if !logJSON {
		cfg = zap.NewDevelopmentConfig()
	} else {
		cfg = zap.NewProductionConfig()
	}

	// Extract log fields from environment variables.
	envFields := FieldsFromEnv(os.Environ())

	options := []zap.Option{
		zap.AddStacktrace(zapcore.FatalLevel),
		zap.WrapCore(func(c zapcore.Core) zapcore.Core {
			return c.With(envFields)
		}),
	}

	if len(path) > 0 {
		w := zapcore.AddSync(&lumberjack.Logger{
			Filename:   path,
			MaxSize:    10, // megabytes
			MaxBackups: 3,
			MaxAge:     28, // days
		})
		var enc zapcore.Encoder
		if logJSON {
			enc = zapcore.NewJSONEncoder(cfg.EncoderConfig)
		} else {
			enc = zapcore.NewConsoleEncoder(cfg.EncoderConfig)
		}
		core := zapcore.NewCore(enc, w, lvl)
		options = append(options, zap.WrapCore(func(zapcore.Core) zapcore.Core {
			return core
		}))
	}

	if config == nil {
		config = &cfg
		config.Level.SetLevel(lvl)
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	logger, err := config.Build(
		// We handle this via errors package for 99% of the stuff so only
		// enable this at the fatal/panic level.
		options...,
	)
	if err != nil {
		panic(err)
	}

	return logger
}

func setFlags() {
	logLvl := viper.GetInt("log-level")
	logJSON = viper.GetBool("log-json")
	lvl = zapcore.Level(logLvl)
	logfilePath = viper.GetString("logfile-path")
}

// FieldsFromEnv extracts log fields from environment variables.
// If an environment variable starts with LOG_FIELD_, the suffix is extracted
// and split on =. The first part is used for the name and the second for the
// value.
// For example, LOG_FIELD_foo=bar would result in a field named "foo" with the
// value "bar".
func FieldsFromEnv(env []string) []zapcore.Field {
	const logfieldPrefix = "LOG_FIELD_"

	fields := []zapcore.Field{}
	for _, s := range env {
		if !strings.HasPrefix(s, logfieldPrefix) || len(s) < (len(logfieldPrefix)+1) {
			continue
		}
		s = s[len(logfieldPrefix):]
		parts := strings.SplitN(s, "=", 2)
		if len(parts) != 2 {
			continue
		}
		fields = append(fields, zap.String(parts[0], parts[1]))
	}
	return fields
}
