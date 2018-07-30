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

package log

import (
	"flag"
	"fmt"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	lvlString     = "info"
	logJSON       = false
	logfilePath   = ""
	config        *zap.Config
	mu            sync.Mutex
	glogVerbosity string
)

func init() {
	flag.StringVar(&lvlString, "log-level", lvlString, "Adjusts the level of the logs that will be omitted.")
	flag.BoolVar(&logJSON, "log-json", logJSON, "Log in json format.")
	flag.StringVar(&logfilePath, "logfile-path", "", "If specified, write log messages to a file at this path.")
	_ = flag.String("v", "", "For glog backwards compat. Does nothing.")
}

// Options holds the zap logger configuration.
type Options struct {
	LogLevel *zapcore.Level
	Config   *zap.Config
}

// Level gets the current log level.
func Level() *zap.AtomicLevel {
	return &config.Level
}

func levelFromString(level string) (*zapcore.Level, error) {
	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(level)); err != nil {
		return nil, fmt.Errorf("invalid logging level: %v", level)
	}
	return &zapLevel, nil
}

// Logger builds a new logger based on the given flags.
func Logger() *zap.Logger {
	mu.Lock()
	defer mu.Unlock()

	var cfg zap.Config

	if !logJSON {
		cfg = zap.NewDevelopmentConfig()
	} else {
		cfg = zap.NewProductionConfig()
	}

	if len(logfilePath) > 0 {
		cfg.OutputPaths = append(cfg.OutputPaths, logfilePath)
	}

	if config == nil {
		config = &cfg
		if level, err := levelFromString(lvlString); err == nil {
			config.Level = zap.NewAtomicLevelAt(*level)
		}
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	logger, err := config.Build(
		// We handle this via errors package for 99% of the stuff so only
		// enable this at the fatal/panic level.
		zap.AddStacktrace(zapcore.FatalLevel),
	)
	if err != nil {
		panic(err)
	}

	return logger
}
