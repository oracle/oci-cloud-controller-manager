// Copyright 2017 Oracle and/or its affiliates. All rights reserved.
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

package disk

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
	"k8s.io/mount-utils"
	utilexec "k8s.io/utils/exec"
)

func TestGetMountPointForPath(t *testing.T) {
	testCases := []struct {
		name     string
		mps      []mount.MountPoint
		path     string
		err      error
		expected mount.MountPoint
	}{
		{
			name: "single",
			mps: []mount.MountPoint{
				{Path: "/tmp/my-mountpoint"},
			},
			path:     "/tmp/my-mountpoint",
			err:      nil,
			expected: mount.MountPoint{Path: "/tmp/my-mountpoint"},
		}, {
			name: "multiple",
			mps: []mount.MountPoint{
				{Path: "/tmp/my-other-mountpoint"},
				{Path: "/tmp/my-mountpoint"},
			},
			path:     "/tmp/my-mountpoint",
			err:      nil,
			expected: mount.MountPoint{Path: "/tmp/my-mountpoint"},
		}, {
			name: "missing",
			mps: []mount.MountPoint{
				{Path: "/tmp/my-other-mountpoint"},
			},
			path:     "/tmp/my-mountpoint",
			err:      ErrMountPointNotFound,
			expected: mount.MountPoint{},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			mock := mount.NewFakeMounter(tt.mps)
			result, err := getMountPointForPath(mock, tt.path)
			if err != tt.err {
				t.Fatalf("getMountPointForPath(mockLister, %q) => error: %v; expected %v", tt.path, err, tt.err)
			}
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("getMountPointForPath(mockLister, %q) =>\n%+v\nExpected: %+v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestISCSIMounterLogin(t *testing.T) {
	loginErr := errors.New("exit status 15")
	disk := &Disk{
		IQN:     "iqn.2015-12.com.oracleiaas:volume",
		IscsiIp: "169.254.2.2",
		Port:    3260,
	}

	tests := []struct {
		name                      string
		disk                      *Disk
		responses                 []fakeExecResponse
		wantErr                   string
		wantCalls                 int
		wantSessionCombinedOutput bool
	}{
		{
			name: "login succeeds",
			responses: []fakeExecResponse{
				{},
			},
			wantCalls: 1,
		},
		{
			name: "failed login succeeds when target session is active",
			responses: []fakeExecResponse{
				{err: loginErr},
				{combinedOutput: "tcp: [2] 169.254.2.2:3260,1 iqn.2015-12.com.oracleiaas:volume (non-flash)\n"},
			},
			wantCalls:                 2,
			wantSessionCombinedOutput: true,
		},
		{
			name: "failed login returns original error when session query succeeds with no output",
			responses: []fakeExecResponse{
				{err: loginErr},
				{},
			},
			wantErr:   "iscsi: error logging in target: exit status 15",
			wantCalls: 2,
		},
		{
			name: "failed login succeeds when a later session matches",
			responses: []fakeExecResponse{
				{err: loginErr},
				{combinedOutput: "tcp: [2] 169.254.2.3:3260,1 iqn.2015-12.com.oracleiaas:another-volume (non-flash)\n" +
					"tcp: [3] 169.254.2.2:3260,1 iqn.2015-12.com.oracleiaas:volume (non-flash)\n"},
			},
			wantCalls: 2,
		},
		{
			name: "failed login succeeds when IPv6 target session already exists",
			disk: &Disk{
				IQN:     "iqn.2015-12.com.oracleiaas:volume",
				IscsiIp: "fd00:c1::a9fe:202",
				Port:    3260,
			},
			responses: []fakeExecResponse{
				{err: loginErr},
				{combinedOutput: "tcp: [2] [fd00:00c1:0000:0000:0000:0000:a9fe:0202]:3260,1 iqn.2015-12.com.oracleiaas:volume (non-flash)\n"},
			},
			wantCalls: 2,
		},
		{
			name: "failed login returns error when only another target session is active",
			responses: []fakeExecResponse{
				{err: loginErr},
				{combinedOutput: "tcp: [2] 169.254.2.2:3260,1 iqn.2015-12.com.oracleiaas:another-volume (non-flash)\n"},
			},
			wantErr:   "iscsi: error logging in target: exit status 15",
			wantCalls: 2,
		},
		{
			name: "failed login returns error when target session uses another portal",
			responses: []fakeExecResponse{
				{err: loginErr},
				{combinedOutput: "tcp: [2] 169.254.2.3:3260,1 iqn.2015-12.com.oracleiaas:volume (non-flash)\n"},
			},
			wantErr:   "iscsi: error logging in target: exit status 15",
			wantCalls: 2,
		},
		{
			name: "failed login returns error when session lookup fails",
			responses: []fakeExecResponse{
				{err: loginErr},
				{combinedOutput: "iscsiadm: session lookup failed", err: errors.New("session lookup failed")},
			},
			wantErr:   "iscsi: error logging in target: exit status 15",
			wantCalls: 2,
		},
		{
			name: "failed session query with matching output returns original login error",
			responses: []fakeExecResponse{
				{err: loginErr},
				{combinedOutput: "tcp: [2] 169.254.2.2:3260,1 iqn.2015-12.com.oracleiaas:volume (non-flash)\n", err: errors.New("session lookup failed")},
			},
			wantErr:   "iscsi: error logging in target: exit status 15",
			wantCalls: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &fakeExec{responses: tt.responses}
			testDisk := disk
			if tt.disk != nil {
				testDisk = tt.disk
			}
			mounter := &iSCSIMounter{
				disk:         testDisk,
				runner:       runner,
				iscsiadmPath: "/usr/sbin/iscsiadm",
				logger:       zap.NewNop().Sugar(),
			}

			err := mounter.Login()
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("Login() error = %v, want nil", err)
				}
			} else if err == nil || err.Error() != tt.wantErr {
				t.Fatalf("Login() error = %v, want %q", err, tt.wantErr)
			}
			if len(runner.calls) != tt.wantCalls {
				t.Fatalf("Login() made %d commands, want %d", len(runner.calls), tt.wantCalls)
			}
			if tt.wantCalls == 2 {
				wantSessionCommand := []string{"/usr/sbin/iscsiadm", "-m", "session"}
				if !reflect.DeepEqual(runner.calls[1], wantSessionCommand) {
					t.Fatalf("Login() session command = %v, want %v", runner.calls[1], wantSessionCommand)
				}
				if tt.wantSessionCombinedOutput && runner.outputMethods[1] != "CombinedOutput" {
					t.Fatalf("Login() session command used %s, want CombinedOutput", runner.outputMethods[1])
				}
			}
		})
	}
}

func TestISCSIMounterLoginLogsCombinedSessionOutput(t *testing.T) {
	const sessionErrorOutput = "iscsiadm: session lookup failed"
	core, logs := observer.New(zap.InfoLevel)
	mounter := &iSCSIMounter{
		disk: &Disk{
			IQN:     "iqn.2015-12.com.oracleiaas:volume",
			IscsiIp: "169.254.2.2",
			Port:    3260,
		},
		runner: &fakeExec{responses: []fakeExecResponse{
			{err: errors.New("exit status 15")},
			{combinedOutput: sessionErrorOutput, err: errors.New("session lookup failed")},
		}},
		iscsiadmPath: "/usr/sbin/iscsiadm",
		logger:       zap.New(core).Sugar(),
	}

	if err := mounter.Login(); err == nil {
		t.Fatal("Login() error = nil, want error")
	}

	entries := logs.FilterMessage("Unable to list active iSCSI sessions.").All()
	if len(entries) != 1 {
		t.Fatalf("session lookup log entries = %d, want 1", len(entries))
	}
	if got := entries[0].ContextMap()["output"]; got != sessionErrorOutput {
		t.Fatalf("session lookup output = %v, want %q", got, sessionErrorOutput)
	}
}

func TestISCSISessionOutputContainsTarget(t *testing.T) {
	const iqn = "iqn.2015-12.com.oracleiaas:volume"
	const ipv4Target = "169.254.2.2:3260"
	const ipv4Session = "tcp: [2] 169.254.2.2:3260,1 " + iqn + " (non-flash)\n"

	tests := []struct {
		name   string
		output string
		target string
		iqn    string
		want   bool
	}{
		{
			name:   "skips blank and malformed lines before a match",
			output: "\ninvalid session output\n\n" + ipv4Session,
			target: ipv4Target,
			iqn:    iqn,
			want:   true,
		},
		{
			name:   "rejects invalid target endpoint",
			output: ipv4Session,
			target: "invalid-target",
			iqn:    iqn,
		},
		{
			name:   "skips invalid session endpoint",
			output: "tcp: [2] invalid-portal,1 " + iqn + " (non-flash)\n",
			target: ipv4Target,
			iqn:    iqn,
		},
		{
			name:   "rejects session without IQN",
			output: "tcp: [2] 169.254.2.2:3260,1\n",
			target: ipv4Target,
			iqn:    iqn,
		},
		{
			name:   "strips TPGT suffix",
			output: "tcp: [2] 169.254.2.2:3260,42 " + iqn + " (non-flash)\n",
			target: ipv4Target,
			iqn:    iqn,
			want:   true,
		},
		{
			name:   "matches exact IPv6 endpoint from node",
			output: "tcp: [1] [fd00:c1::a9fe:202]:3260,1 " + iqn + " (non-flash)\n",
			target: "[fd00:c1::a9fe:202]:3260",
			iqn:    iqn,
			want:   true,
		},
		{
			name:   "matches equivalent IPv6 endpoint representation",
			output: "tcp: [1] [fd00:c1::a9fe:202]:3260,1 " + iqn + " (non-flash)\n",
			target: "[fd00:00c1:0000:0000:0000:0000:a9fe:0202]:3260",
			iqn:    iqn,
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isIscsiSessionOutputContainsTarget(tt.output, tt.target, tt.iqn); got != tt.want {
				t.Fatalf("isIscsiSessionOutputContainsTarget() = %t, want %t", got, tt.want)
			}
		})
	}
}

func TestISCSISessionActiveForDiskPath(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("requires a POSIX shell script")
	}

	const iqn = "iqn.2015-12.com.oracleiaas:volume"
	tests := []struct {
		name          string
		diskPath      string
		sessionOutput string
		stderr        string
		exitCode      int
		wantActive    bool
	}{
		{
			name:          "IPv4",
			diskPath:      "/dev/disk/by-path/ip-169.254.2.2:3260-iscsi-" + iqn + "-lun-1",
			sessionOutput: "tcp: [2] 169.254.2.2:3260,1 " + iqn + " (non-flash)",
			wantActive:    true,
		},
		{
			name:          "IPv6 equivalent endpoint representation",
			diskPath:      "/dev/disk/by-path/ip-fd00:00c1::a9fe:202:3260-iscsi-" + iqn + "-lun-1",
			sessionOutput: "tcp: [2] [fd00:c1::a9fe:202]:3260,1 " + iqn + " (non-flash)",
			wantActive:    true,
		},
		{
			name:          "IPv6 exact endpoint from node",
			diskPath:      "/dev/disk/by-path/ip-fd00:c1::a9fe:202:3260-iscsi-" + iqn + "-lun-2",
			sessionOutput: "tcp: [1] [fd00:c1::a9fe:202]:3260,1 " + iqn + " (non-flash)",
			wantActive:    true,
		},
		{
			name:       "no active session",
			diskPath:   "/dev/disk/by-path/ip-169.254.2.2:3260-iscsi-" + iqn + "-lun-1",
			stderr:     "iscsiadm: No active sessions.",
			exitCode:   21,
			wantActive: false,
		},
		{
			name:          "same portal different IQN",
			diskPath:      "/dev/disk/by-path/ip-169.254.2.2:3260-iscsi-" + iqn + "-lun-1",
			sessionOutput: "tcp: [2] 169.254.2.2:3260,1 iqn.2015-12.com.oracleiaas:another-volume (non-flash)",
			wantActive:    false,
		},
		{
			name:          "same IQN different portal",
			diskPath:      "/dev/disk/by-path/ip-169.254.2.2:3260-iscsi-" + iqn + "-lun-1",
			sessionOutput: "tcp: [2] 169.254.2.3:3260,1 " + iqn + " (non-flash)",
			wantActive:    false,
		},
		{
			name:          "malformed disk path",
			diskPath:      "/dev/disk/by-path/not-an-iscsi-path",
			sessionOutput: "tcp: [2] 169.254.2.2:3260,1 " + iqn + " (non-flash)",
			wantActive:    false,
		},
		{
			name:       "generic session query failure",
			diskPath:   "/dev/disk/by-path/ip-169.254.2.2:3260-iscsi-" + iqn + "-lun-1",
			stderr:     "iscsiadm: permission denied",
			exitCode:   1,
			wantActive: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			iscsiadm := filepath.Join(t.TempDir(), "iscsiadm")
			script := "#!/bin/sh\nprintf '%s\\n' '" + tt.sessionOutput + "'\n"
			if tt.stderr != "" {
				script += "printf '%s\\n' '" + tt.stderr + "' >&2\n"
			}
			if tt.exitCode != 0 {
				script += "exit " + strconv.Itoa(tt.exitCode) + "\n"
			}
			if err := os.WriteFile(iscsiadm, []byte(script), 0o755); err != nil {
				t.Fatalf("WriteFile(%q): %v", iscsiadm, err)
			}
			t.Setenv("PATH", filepath.Dir(iscsiadm))

			if got := isISCSISessionActiveForDiskPath(tt.diskPath, zap.NewNop().Sugar()); got != tt.wantActive {
				t.Fatalf("isISCSISessionActiveForDiskPath() = %t, want %t", got, tt.wantActive)
			}
		})
	}
}

type fakeExecResponse struct {
	output         string
	combinedOutput string
	err            error
}

type fakeExec struct {
	responses     []fakeExecResponse
	calls         [][]string
	outputMethods []string
}

func (f *fakeExec) Command(cmd string, args ...string) utilexec.Cmd {
	f.calls = append(f.calls, append([]string{cmd}, args...))
	response := f.responses[len(f.calls)-1]
	return &fakeCmd{response: response, runner: f}
}

func (f *fakeExec) CommandContext(_ context.Context, cmd string, args ...string) utilexec.Cmd {
	return f.Command(cmd, args...)
}

func (f *fakeExec) LookPath(_ string) (string, error) {
	return "/usr/sbin/iscsiadm", nil
}

type fakeCmd struct {
	response fakeExecResponse
	runner   *fakeExec
}

func (c *fakeCmd) Run() error { return c.response.err }
func (c *fakeCmd) CombinedOutput() ([]byte, error) {
	c.runner.outputMethods = append(c.runner.outputMethods, "CombinedOutput")
	return []byte(c.response.combinedOutput), c.response.err
}
func (c *fakeCmd) Output() ([]byte, error) {
	c.runner.outputMethods = append(c.runner.outputMethods, "Output")
	return []byte(c.response.output), c.response.err
}
func (c *fakeCmd) SetDir(_ string)       {}
func (c *fakeCmd) SetStdin(_ io.Reader)  {}
func (c *fakeCmd) SetStdout(_ io.Writer) {}
func (c *fakeCmd) SetStderr(_ io.Writer) {}
func (c *fakeCmd) SetEnv(_ []string)     {}
func (c *fakeCmd) StdoutPipe() (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("")), nil
}
func (c *fakeCmd) StderrPipe() (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("")), nil
}
func (c *fakeCmd) Start() error { return c.response.err }
func (c *fakeCmd) Wait() error  { return c.response.err }
func (c *fakeCmd) Stop()        {}
