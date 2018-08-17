// Copyright 2018 Istio Authors
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

package glog

import (
	"bytes"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestAll(t *testing.T) {
	// just making sure this stuff doesn't crash...

	Errorf("%s %s", "One", "Two")
	Error("One", "Two")
	Errorln("One", "Two")
	ErrorDepth(2, "One", "Two")

	Warningf("%s %s", "One", "Two")
	Warning("One", "Two")
	Warningln("One", "Two")
	WarningDepth(2, "One", "Two")

	Infof("%s %s", "One", "Two")
	Info("One", "Two")
	Infoln("One", "Two")
	InfoDepth(2, "One", "Two")

	for i := 0; i < 10; i++ {
		V(Level(i)).Infof("%s %s", "One", "Two")
		V(Level(i)).Info("One", "Two")
		V(Level(i)).Infoln("One", "Two")
	}

	Flush()
}

func assertNumEntries(t *testing.T, b []byte, expected int) {
	t.Helper()
	count := bytes.Count(b, []byte{'\n'})
	if count != expected {
		t.Errorf("Expected %d logger entries but got %d", expected, count)
	}
}

func TestReplaceGlobals(t *testing.T) {
	b := new(bytes.Buffer)
	logger := zap.New(
		zapcore.NewCore(
			zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
			zapcore.AddSync(b),
			zap.InfoLevel,
		),
	)

	// Log to global logger.
	Info("foo")

	assertNumEntries(t, b.Bytes(), 0)

	zap.ReplaceGlobals(logger)

	// Log to logger
	Info("bar")

	assertNumEntries(t, b.Bytes(), 1)

	b2 := new(bytes.Buffer)
	logger2 := zap.New(
		zapcore.NewCore(
			zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
			zapcore.AddSync(b2),
			zap.InfoLevel,
		),
	)

	zap.ReplaceGlobals(logger2)

	// Log to logger2
	Info("baz")

	assertNumEntries(t, b.Bytes(), 1)
	assertNumEntries(t, b2.Bytes(), 1)
}

func BenchmarkSkipLogger(b *testing.B) {
	for n := 0; n < b.N; n++ {
		skipLogger()
	}
}
