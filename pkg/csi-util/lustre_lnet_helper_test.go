package csi_util

import (
	"fmt"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestValidateLustreVolumeId(t *testing.T) {
	tests := []struct {
		input                    string
		expectedValidationResult bool
		expectedLnetLable        string
	}{
		// Valid cases
		{"192.168.227.11@tcp1:192.168.227.12@tcp1:/demo", true, "tcp1"},
		{"192.168.227.11@tcp1:/demo", true, "tcp1"},
		{"192.168.227.11@tcp1 & rm -rf /:192.168.227.12@tcp1:/demo", false, ""},
		{"192.168.227.11@tcp1:192.168.227.12@tcp1:/demo & rm -rf", false, "tcp1"},
		{"192.168.227.11@tcp1:/demo", true, "tcp1"},
		// Invalid cases
		{"192.168.227.11@tcp1:192.168.227.12@tcp1", false, "tcp1"},      // No fsname provided
		{"192.168.227.11@tcp1:192.168.227.12@tcp1:demo", false, "tcp1"}, // fsname not starting with "/"
		{"invalidip@tcp1:192.168.227.12@tcp1:/demo", false, ""},         // Invalid IP address
		{"192.168.227.11@tcp1:invalidip@tcp1:/demo", false, "tcp1"},     // Invalid IP address
		{"192.168.227.11@:192.168.227.12@:tcp1/demo", false, ""},        // No Lnet label provided
		{"192.168.227.11@tcp1:192.168.227.12:/demo", false, "tcp1"},     // No Lnet label provided in 2nd MGS NID
		// Empty input
		{"", false, ""},

		// Single IP
		{"192.168.227.11", false, ""}, // Missing ":" in the input
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			validationResult, lnetLabel := ValidateLustreVolumeId(test.input)
			if validationResult != test.expectedValidationResult || lnetLabel != test.expectedLnetLable {
				t.Errorf("For input '%s', expected validationResult : %v & lnetLable : %v but got validationResult : %v & lnetLable : %v", test.input, test.expectedValidationResult, test.expectedLnetLable, validationResult, lnetLabel)
			}
		})
	}
}

// FakeConfigurator is a fake implementation of LnetConfigurator.
type FakeConfigurator struct {
	GetNetInterfacesInSubnetFunc        func(subnetCIDR string) ([]NetInterface, error)
	IsLustreClientPackagesInstalledFunc func(logger *zap.SugaredLogger) bool
	GetLnetInfoByLnetLabelFunc          func(logger *zap.SugaredLogger, lnetLabel string) (NetInfo, error)
	ConfigureLnetFunc                   func(logger *zap.SugaredLogger, ifaces []NetInterface, lnetLabel string, netInfo NetInfo) error
	VerifyLnetConfigurationFunc         func(logger *zap.SugaredLogger, ifaces []NetInterface, lnetLabel string, netInfo NetInfo, err error) error
	ExecuteCommandOnWorkerNodeFunc      func(args ...string) (string, error)
}

func (f *FakeConfigurator) GetNetInterfacesInSubnet(subnetCIDR string) ([]NetInterface, error) {
	return f.GetNetInterfacesInSubnetFunc(subnetCIDR)
}

func (f *FakeConfigurator) IsLustreClientPackagesInstalled(logger *zap.SugaredLogger) bool {
	return f.IsLustreClientPackagesInstalledFunc(logger)
}

func (f *FakeConfigurator) GetLnetInfoByLnetLabel(logger *zap.SugaredLogger, lnetLabel string) (NetInfo, error) {
	return f.GetLnetInfoByLnetLabelFunc(logger, lnetLabel)
}

func (f *FakeConfigurator) ConfigureLnet(logger *zap.SugaredLogger, ifaces []NetInterface, lnetLabel string, netInfo NetInfo) error {
	return f.ConfigureLnetFunc(logger, ifaces, lnetLabel, netInfo)
}

func (f *FakeConfigurator) VerifyLnetConfiguration(logger *zap.SugaredLogger, ifaces []NetInterface, lnetLabel string, netInfo NetInfo, err error) error {
	return f.VerifyLnetConfigurationFunc(logger, ifaces, lnetLabel, netInfo, err)
}

func (f *FakeConfigurator) ExecuteCommandOnWorkerNode(args ...string) (string, error) {
	return f.ExecuteCommandOnWorkerNodeFunc(args...)
}

func TestLnetService_SetupLnet_TableDriven(t *testing.T) {
	logger := zap.NewExample().Sugar()

	dummyInterfaces := []NetInterface{
		{InterfaceName: "eth0", InterfaceIPv4: "10.244.0.10", LnetConfigured: false},
	}
	dummyNetInfo := NetInfo{
		Net: []struct {
			NetType string "yaml:\"net type\""
			LocalNI []struct {
				NID        string         "yaml:\"nid\""
				Status     string         "yaml:\"status\""
				Interfaces map[int]string "yaml:\"interfaces\""
			} "yaml:\"local NI(s)\""
		}{
			{
				NetType: "tcp",
				LocalNI: []struct {
					NID        string         "yaml:\"nid\""
					Status     string         "yaml:\"status\""
					Interfaces map[int]string "yaml:\"interfaces\""
				}{
					{
						NID:        "10.244.0.10@tcp1",
						Status:     "up",
						Interfaces: map[int]string{0: "eth0"},
					},
				},
			},
		},
	}

	tests := []struct {
		name              string
		lustreSubnetCIDR  string
		lnetLabel         string
		fakeCfg           *FakeConfigurator
		expectedErrSubstr string
	}{
		{
			name:             "No Interfaces Found",
			lustreSubnetCIDR: "10.244.0.0/24",
			lnetLabel:        "tcp1",
			fakeCfg: &FakeConfigurator{
				GetNetInterfacesInSubnetFunc: func(subnetCIDR string) ([]NetInterface, error) {
					return []NetInterface{}, nil
				},
				// These functions are not used in this case.
				IsLustreClientPackagesInstalledFunc: func(logger *zap.SugaredLogger) bool { return true },
				ExecuteCommandOnWorkerNodeFunc:      func(args ...string) (string, error) { return "ok", nil },
				GetLnetInfoByLnetLabelFunc:          func(logger *zap.SugaredLogger, lnetLabel string) (NetInfo, error) { return NetInfo{}, nil },
				ConfigureLnetFunc: func(logger *zap.SugaredLogger, ifaces []NetInterface, lnetLabel string, netInfo NetInfo) error {
					return nil
				},
				VerifyLnetConfigurationFunc: func(logger *zap.SugaredLogger, ifaces []NetInterface, lnetLabel string, netInfo NetInfo, err error) error {
					return nil
				},
			},
			expectedErrSubstr: "No VNIC identified on worker node to configure lnet.",
		},
		{
			name:             "Network Interface identified in provided CIDR and Successful Lnet Configuration",
			lustreSubnetCIDR: "10.244.0.0/24",
			lnetLabel:        "tcp1",
			fakeCfg: &FakeConfigurator{
				GetNetInterfacesInSubnetFunc: func(subnetCIDR string) ([]NetInterface, error) {
					return dummyInterfaces, nil
				},
				IsLustreClientPackagesInstalledFunc: func(logger *zap.SugaredLogger) bool { return true },
				ExecuteCommandOnWorkerNodeFunc: func(args ...string) (string, error) {
					if args[0] == LOAD_LNET_KERNEL_MODULE_COMMAND || args[0] == CONFIGURE_LNET_KERNEL_SERVICE_COMMAND {
						return "ok", nil
					}
					if strings.HasPrefix(args[0], "lnetctl net show") {
						return "net: []", nil
					}
					return "ok", nil
				},
				GetLnetInfoByLnetLabelFunc: func(logger *zap.SugaredLogger, lnetLabel string) (NetInfo, error) {
					return dummyNetInfo, nil
				},
				ConfigureLnetFunc: func(logger *zap.SugaredLogger, ifaces []NetInterface, lnetLabel string, netInfo NetInfo) error {
					return nil
				},
				VerifyLnetConfigurationFunc: func(logger *zap.SugaredLogger, ifaces []NetInterface, lnetLabel string, netInfo NetInfo, err error) error {
					return nil
				},
			},
			expectedErrSubstr: "",
		},
		{
			name:             "Load Kernel Module Fails",
			lustreSubnetCIDR: "10.244.0.0/24",
			lnetLabel:        "tcp1",
			fakeCfg: &FakeConfigurator{
				GetNetInterfacesInSubnetFunc: func(subnetCIDR string) ([]NetInterface, error) {
					return dummyInterfaces, nil
				},
				IsLustreClientPackagesInstalledFunc: func(logger *zap.SugaredLogger) bool { return false },
				ExecuteCommandOnWorkerNodeFunc: func(args ...string) (string, error) {
					if args[0] == LOAD_LNET_KERNEL_MODULE_COMMAND {
						return "error", errors.New("loading lnet module failed")
					}
					return "ok", nil
				},
				GetLnetInfoByLnetLabelFunc: func(logger *zap.SugaredLogger, lnetLabel string) (NetInfo, error) { return NetInfo{}, nil },
				ConfigureLnetFunc: func(logger *zap.SugaredLogger, ifaces []NetInterface, lnetLabel string, netInfo NetInfo) error {
					return nil
				},
				VerifyLnetConfigurationFunc: func(logger *zap.SugaredLogger, ifaces []NetInterface, lnetLabel string, netInfo NetInfo, err error) error {
					return nil
				},
			},
			expectedErrSubstr: "Failed to load lnet kernel module with error : loading lnet module failed. Please make sure that lustre client packages are installed on worker nodes.",
		},
		{
			name:             "Lnet Service configuration fails",
			lustreSubnetCIDR: "10.244.0.0/24",
			lnetLabel:        "tcp1",
			fakeCfg: &FakeConfigurator{
				GetNetInterfacesInSubnetFunc: func(subnetCIDR string) ([]NetInterface, error) {
					return dummyInterfaces, nil
				},
				IsLustreClientPackagesInstalledFunc: func(logger *zap.SugaredLogger) bool { return false },
				ExecuteCommandOnWorkerNodeFunc: func(args ...string) (string, error) {
					if args[0] == CONFIGURE_LNET_KERNEL_SERVICE_COMMAND {
						return "error", errors.New("lnet service configuration failed")
					}
					return "ok", nil
				},
				GetLnetInfoByLnetLabelFunc: func(logger *zap.SugaredLogger, lnetLabel string) (NetInfo, error) { return NetInfo{}, nil },
				ConfigureLnetFunc: func(logger *zap.SugaredLogger, ifaces []NetInterface, lnetLabel string, netInfo NetInfo) error {
					return nil
				},
				VerifyLnetConfigurationFunc: func(logger *zap.SugaredLogger, ifaces []NetInterface, lnetLabel string, netInfo NetInfo, err error) error {
					return nil
				},
			},
			expectedErrSubstr: "Failed to configure lnet kernel service with error : lnet service configuration failed. Please make sure that lustre client packages are installed on worker nodes.",
		},
		{
			name:             "Failure during Get Lnet Info By Lnet Label",
			lustreSubnetCIDR: "10.244.0.0/24",
			lnetLabel:        "tcp1",
			fakeCfg: &FakeConfigurator{
				GetNetInterfacesInSubnetFunc: func(subnetCIDR string) ([]NetInterface, error) {
					return dummyInterfaces, nil
				},
				IsLustreClientPackagesInstalledFunc: func(logger *zap.SugaredLogger) bool { return true },
				ExecuteCommandOnWorkerNodeFunc: func(args ...string) (string, error) {
					// Both kernel module load and kernel service config succeed.
					return "ok", nil
				},
				GetLnetInfoByLnetLabelFunc: func(logger *zap.SugaredLogger, lnetLabel string) (NetInfo, error) {
					return NetInfo{}, errors.New("get lnet info error")
				},
			},
			expectedErrSubstr: "get lnet info error",
		},
		{
			name:             "Failed happens during Configure Lnet",
			lustreSubnetCIDR: "10.244.0.0/24",
			lnetLabel:        "tcp1",
			fakeCfg: &FakeConfigurator{
				GetNetInterfacesInSubnetFunc: func(subnetCIDR string) ([]NetInterface, error) {
					return dummyInterfaces, nil
				},
				IsLustreClientPackagesInstalledFunc: func(logger *zap.SugaredLogger) bool { return true },
				ExecuteCommandOnWorkerNodeFunc: func(args ...string) (string, error) {
					return "ok", nil
				},
				GetLnetInfoByLnetLabelFunc: func(logger *zap.SugaredLogger, lnetLabel string) (NetInfo, error) {
					return dummyNetInfo, nil
				},
				ConfigureLnetFunc: func(logger *zap.SugaredLogger, ifaces []NetInterface, lnetLabel string, netInfo NetInfo) error {
					return errors.New("configure lnet failed")
				},
			},
			expectedErrSubstr: "configure lnet failed",
		},
		{
			name:             "Failure to VerifyLnetConfiguration",
			lustreSubnetCIDR: "10.244.0.0/24",
			lnetLabel:        "tcp1",
			fakeCfg: &FakeConfigurator{
				GetNetInterfacesInSubnetFunc: func(subnetCIDR string) ([]NetInterface, error) {
					return dummyInterfaces, nil
				},
				IsLustreClientPackagesInstalledFunc: func(logger *zap.SugaredLogger) bool { return true },
				ExecuteCommandOnWorkerNodeFunc: func(args ...string) (string, error) {
					return "ok", nil
				},
				GetLnetInfoByLnetLabelFunc: func(logger *zap.SugaredLogger, lnetLabel string) (NetInfo, error) {
					return dummyNetInfo, nil
				},
				ConfigureLnetFunc: func(logger *zap.SugaredLogger, ifaces []NetInterface, lnetLabel string, netInfo NetInfo) error {
					return nil
				},
				VerifyLnetConfigurationFunc: func(logger *zap.SugaredLogger, ifaces []NetInterface, lnetLabel string, netInfo NetInfo, err error) error {
					return errors.New("verify lnet failed")
				},
			},
			expectedErrSubstr: "verify lnet failed",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := LnetService{Configurator: tc.fakeCfg}
			err := svc.SetupLnet(logger, tc.lustreSubnetCIDR, tc.lnetLabel)
			if tc.expectedErrSubstr == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErrSubstr)
			}
		})
	}
}

func TestParseLnetInfo(t *testing.T) {
	tests := []struct {
		name              string
		commandOutput     string
		commandErr        error
		expectedErrSubstr string
		expectedNetType   string
		expectedLogSubstr string
	}{
		{
			name: "Lustre 2.15.90 network down response returns empty net info",
			commandOutput: `---
show:
-     net:
      errno: -3
      descr: ! "Network is down"
...
`,
			commandErr:        errors.New("exit status 234"),
			expectedLogSubstr: "Expected failure observed",
		},
		{
			name: "network down response without errno returns empty net info",
			commandOutput: `---
show:
-     net:
      descr: ! "Network is down"
...
`,
			commandErr:        errors.New("exit status 1"),
			expectedLogSubstr: "Expected failure observed",
		},
		{
			name: "valid lnet output is parsed",
			commandOutput: `net:
-     net type: tcp
      local NI(s):
      -     nid: 10.244.0.10@tcp
            status: up
            interfaces:
                  0: eth0
`,
			expectedNetType: "tcp",
		},
		{
			name:              "unrelated command error is returned",
			commandOutput:     "opening /dev/lnet failed: No such device",
			commandErr:        errors.New("exit status 1"),
			expectedErrSubstr: "Failed to get existing configured lnet information",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			core, recorded := observer.New(zap.InfoLevel)
			logger := zap.New(core).Sugar()

			netInfo, err := parseLnetInfo(logger, tc.commandOutput, tc.commandErr)

			if tc.expectedErrSubstr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErrSubstr)
				return
			}

			assert.NoError(t, err)
			if tc.expectedLogSubstr != "" {
				assert.Equal(t, 1, recorded.FilterMessageSnippet(tc.expectedLogSubstr).Len())
			}
			if tc.expectedNetType == "" {
				assert.Empty(t, netInfo.Net)
				return
			}
			if assert.Len(t, netInfo.Net, 1) {
				assert.Equal(t, tc.expectedNetType, netInfo.Net[0].NetType)
			}
		})
	}
}

func TestLnetService_SetupLnet_ConfiguresWhenExistingLnetInfoIsEmpty(t *testing.T) {
	logger := zap.NewExample().Sugar()
	configureCalled := false
	ifaces := []NetInterface{{InterfaceName: "eth0", InterfaceIPv4: "10.244.0.10"}}
	fakeCfg := &FakeConfigurator{
		GetNetInterfacesInSubnetFunc: func(subnetCIDR string) ([]NetInterface, error) {
			return ifaces, nil
		},
		IsLustreClientPackagesInstalledFunc: func(logger *zap.SugaredLogger) bool {
			return true
		},
		ExecuteCommandOnWorkerNodeFunc: func(args ...string) (string, error) {
			return "ok", nil
		},
		GetLnetInfoByLnetLabelFunc: func(logger *zap.SugaredLogger, lnetLabel string) (NetInfo, error) {
			return NetInfo{}, nil
		},
		ConfigureLnetFunc: func(logger *zap.SugaredLogger, gotIfaces []NetInterface, lnetLabel string, netInfo NetInfo) error {
			configureCalled = true
			assert.Equal(t, ifaces, gotIfaces)
			assert.Equal(t, "tcp", lnetLabel)
			assert.Empty(t, netInfo.Net)
			return nil
		},
		VerifyLnetConfigurationFunc: func(logger *zap.SugaredLogger, ifaces []NetInterface, lnetLabel string, netInfo NetInfo, err error) error {
			return nil
		},
	}

	svc := LnetService{Configurator: fakeCfg}

	err := svc.SetupLnet(logger, "10.244.0.0/24", "tcp")

	assert.NoError(t, err)
	assert.True(t, configureCalled)
}

func TestLnetService_IsLnetActive_TableDriven(t *testing.T) {
	logger := zap.NewExample().Sugar()
	tests := []struct {
		name           string
		lnetLabel      string
		fakeGetInfo    func(logger *zap.SugaredLogger, lnetLabel string) (NetInfo, error)
		expectedActive bool
	}{
		{
			name:      "Active Lnet",
			lnetLabel: "tcp1",
			fakeGetInfo: func(logger *zap.SugaredLogger, lnetLabel string) (NetInfo, error) {
				return NetInfo{
					Net: []struct {
						NetType string "yaml:\"net type\""
						LocalNI []struct {
							NID        string         "yaml:\"nid\""
							Status     string         "yaml:\"status\""
							Interfaces map[int]string "yaml:\"interfaces\""
						} "yaml:\"local NI(s)\""
					}{
						{
							NetType: "tcp",
							LocalNI: []struct {
								NID        string         "yaml:\"nid\""
								Status     string         "yaml:\"status\""
								Interfaces map[int]string "yaml:\"interfaces\""
							}{
								{NID: "10.244.0.10@tcp1", Status: "up", Interfaces: map[int]string{0: "eth0"}},
							},
						},
					},
				}, nil
			},
			expectedActive: true,
		},
		{
			name:      "Inactive Lnet",
			lnetLabel: "tcp1",
			fakeGetInfo: func(logger *zap.SugaredLogger, lnetLabel string) (NetInfo, error) {
				return NetInfo{
					Net: []struct {
						NetType string "yaml:\"net type\""
						LocalNI []struct {
							NID        string         "yaml:\"nid\""
							Status     string         "yaml:\"status\""
							Interfaces map[int]string "yaml:\"interfaces\""
						} "yaml:\"local NI(s)\""
					}{
						{
							NetType: "tcp",
							LocalNI: []struct {
								NID        string         "yaml:\"nid\""
								Status     string         "yaml:\"status\""
								Interfaces map[int]string "yaml:\"interfaces\""
							}{
								{NID: "10.244.0.10@tcp1", Status: "down", Interfaces: map[int]string{0: "eth0"}},
							},
						},
					},
				}, nil
			},
			expectedActive: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fakeCfg := &FakeConfigurator{
				GetLnetInfoByLnetLabelFunc: tc.fakeGetInfo,
			}
			svc := LnetService{Configurator: fakeCfg}
			active := svc.IsLnetActive(logger, tc.lnetLabel)
			assert.Equal(t, tc.expectedActive, active)
		})
	}
}

func TestLnetService_ApplyLustreParameters(t *testing.T) {
	logger := zap.NewExample().Sugar()
	tests := []struct {
		name             string
		lustreParamsJSON string
		fakeExec         func(args ...string) (string, error)
		expectedErr      bool
	}{
		{
			name:             "Valid Lustre Parameters Json Provided",
			lustreParamsJSON: `[{"failover.recovery_mode":"quorum","lnet.debug":"0x200"}]`,
			fakeExec: func(args ...string) (string, error) {
				if !strings.Contains(args[0], LCTL_SET_PARAM[:6]) {
					return "", errors.New("unexpected command")
				}
				return "ok", nil
			},
			expectedErr: false,
		},
		{
			name:             "No Lustre Parameters Provided",
			lustreParamsJSON: "",
			fakeExec: func(args ...string) (string, error) {
				return "ok", nil
			},
			expectedErr: false,
		},
		{
			name:             "Invalid Lustre Parameters Json Provided",
			lustreParamsJSON: `invalid-json`,
			fakeExec: func(args ...string) (string, error) {
				return "ok", nil
			},
			expectedErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fakeCfg := &FakeConfigurator{
				ExecuteCommandOnWorkerNodeFunc: tc.fakeExec,
			}
			svc := LnetService{Configurator: fakeCfg}
			err := svc.ApplyLustreParameters(logger, tc.lustreParamsJSON)
			if tc.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLnetService_ValidateLustreParameters(t *testing.T) {
	logger := zap.NewExample().Sugar()
	tests := []struct {
		name             string
		lustreParamsJSON string
		expectedErr      error
	}{
		{
			name:             "Valid Lustre Parameters Json Provided",
			lustreParamsJSON: `[{"failover.recovery_mode":"quorum","lnet.debug":"0x200"}]`,
			expectedErr:      nil,
		},
		{
			name:             "No Lustre Parameters Provided",
			lustreParamsJSON: "",
			expectedErr:      nil,
		},
		{
			name:             "Invalid Lustre Parameters Json Provided",
			lustreParamsJSON: `invalid-json`,
			expectedErr:      fmt.Errorf("%s", "invalid character 'i' looking for beginning of value"),
		},
		{
			name:             "Valid and Invalid Lustre Parameters Provided",
			lustreParamsJSON: `[{"failover.recovery_mode":"quorum","lnet.debug":"0x200","lnet.debug && ls -l | wc -l":"0x200 I am Invalid"}]`,
			expectedErr:      fmt.Errorf("%v", "lnet.debug && ls -l | wc -l=0x200 I am Invalid"),
		},
		{
			name:             "Invalid Lustre Parameters Provided",
			lustreParamsJSON: `[{"failover.recovery_mode;cat /var/log/cloud-init.log":"quorum; echo Hello"}]`,
			expectedErr:      fmt.Errorf("%v", "failover.recovery_mode;cat /var/log/cloud-init.log=quorum; echo Hello"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateLustreParameters(logger, tc.lustreParamsJSON)
			if err != nil && !strings.EqualFold(tc.expectedErr.Error(), err.Error()) {
				t.Errorf("ValidateLustreParameters() got = %v, want %v", err, tc.expectedErr)
			}

		})
	}
}
