package csi_util

import (
	"encoding/json"
	"fmt"
	"net"
	"os/exec"
	"strings"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

const (
	LOAD_LNET_KERNEL_MODULE_COMMAND       = "modprobe lnet"
	CONFIGURE_LNET_KERNEL_SERVICE_COMMAND = "lnetctl lnet configure"
	SHOW_CONFIGURED_LNET                  = "lnetctl net show --net %s"
	DELETE_LNET_INTERFACE                 = "lnetctl net del --net %s"
	CONFIGURE_LNET_INTERFACE              = "lnetctl net add --net %s --if %s --peer-timeout 180 --peer-credits 120 --credits 1024"
	LCTL_SET_PARAM                        = "lctl set_param %s=%s"
	LFS_VERSION_COMMAND                   = "lfs --version"
)

type NetInfo struct {
	Net []struct {
		NetType string `yaml:"net type"`
		LocalNI []struct {
			NID        string         `yaml:"nid"`
			Status     string         `yaml:"status"`
			Interfaces map[int]string `yaml:"interfaces"`
		} `yaml:"local NI(s)"`
	} `yaml:"net"`
}

type NetInterface struct {
	InterfaceIPv4  string
	InterfaceName  string
	LnetConfigured bool
}
type Parameter map[string]interface{}

/*
ValidateLustreVolumeId takes lustreVolumeId as input and returns if its valid or not along with lnetLabel
Ex. volume handle :  10.112.10.6@tcp1:/fsname
 volume handle : <MGS NID>[:<MGS NID>]:/<fsname>
*/
func ValidateLustreVolumeId(lusterVolumeId string) (bool, string) {
	const minNumOfParamsFromVolumeHandle = 2
	const numOfParamsForMGSNID = 2

	lnetLabel := ""
	splits := strings.Split(lusterVolumeId, ":")
	if len(splits) < minNumOfParamsFromVolumeHandle {
		return false, lnetLabel
	}
	for i := 0; i < len(splits)-1; i++ {
		//Ex. split[i] will be 10.112.10.6@tcp1
		parts := strings.Split(splits[i], "@")

		if len(parts) != numOfParamsForMGSNID {
			return false, lnetLabel
		}
		ip := parts[0]
		if net.ParseIP(ip) == nil {
			return false, lnetLabel
		}
		lnetLabel = parts[1]
		if !isValidShellInput(lnetLabel) {
			return false, ""
		}
	}
	//last part in volume handle which is fsname should start with "/"
	if !strings.HasPrefix(splits[len(splits)-1], "/") || !isValidShellInput(splits[len(splits)-1]) {
		return false, lnetLabel
	}
	return true, lnetLabel
}

type LnetConfigurator interface {
	GetNetInterfacesInSubnet(subnetCIDR string) ([]NetInterface, error)
	IsLustreClientPackagesInstalled(logger *zap.SugaredLogger) bool
	GetLnetInfoByLnetLabel(lnetLabel string) (NetInfo, error)
	ConfigureLnet(logger *zap.SugaredLogger, ifaces []NetInterface, lnetLabel string, netInfo NetInfo) error
	VerifyLnetConfiguration(logger *zap.SugaredLogger, ifaces []NetInterface, lnetLabel string, netInfo NetInfo, err error) error
	ExecuteCommandOnWorkerNode(args ...string) (string, error)
}

type LnetService struct {
	Configurator LnetConfigurator
}

type OCILnetConfigurator struct{}

func NewLnetService() *LnetService{
	return &LnetService{
		Configurator:  &OCILnetConfigurator{},
	}
}

func (ls *LnetService) SetupLnet(logger *zap.SugaredLogger, lustreSubnetCIDR string, lnetLabel string) error {

	logger.With("LustreSubnetCidr", lustreSubnetCIDR).With("LnetLabel", lnetLabel).Info("Lnet setup started.")

	//Find net interfaces in lnet subnet on worker node
	interfacesInLustreSubnet, err := ls.Configurator.GetNetInterfacesInSubnet(lustreSubnetCIDR)
	if err != nil {
		return err
	}
	if len(interfacesInLustreSubnet) == 0 {
		return fmt.Errorf("No VNIC identified on worker node to configure lnet.")
	}
	logger.Infof("Net interfaces are identified for lnet configuration: %v", interfacesInLustreSubnet)

	//Lustre client installation state is kept non breaking currently and we still go ahead and try loading kernel module and start lnet kernel service
	//If any of these fail then we provide information to customer on missing packages.
	lustreClientPacakagesInstalled := ls.Configurator.IsLustreClientPackagesInstalled(logger)

	//Load lnet kernel module
	_, err = ls.Configurator.ExecuteCommandOnWorkerNode(LOAD_LNET_KERNEL_MODULE_COMMAND)
	if err != nil {
		if !lustreClientPacakagesInstalled {
			return fmt.Errorf("Failed to load lnet kernel module with error : %v. Please make sure that lustre client packages are installed on worker nodes.", err)
		}
		return fmt.Errorf("Failed to load lnet kernel module with error %v", err)
	}

	//Configure lnet kernel service
	_, err = ls.Configurator.ExecuteCommandOnWorkerNode(CONFIGURE_LNET_KERNEL_SERVICE_COMMAND)
	if err != nil {
		if !lustreClientPacakagesInstalled {
			return fmt.Errorf("Failed to configure lnet kernel service with error : %v. Please make sure that lustre client packages are installed on worker nodes.", err)
		}
		return fmt.Errorf("Failed to configure lnet kernel service with error : %v", err)
	}

	//get existing lnet configuration
	var netInfo NetInfo
	netInfo, err = ls.Configurator.GetLnetInfoByLnetLabel(lnetLabel)
	if err != nil {
		return err
	}

	//Configure lnet if its not configured already for requried interfaces
	err = ls.Configurator.ConfigureLnet(logger, interfacesInLustreSubnet, lnetLabel, netInfo)
	if err != nil {
		return err
	}

	//Verify lnet configuration
	err = ls.Configurator.VerifyLnetConfiguration(logger, interfacesInLustreSubnet, lnetLabel, netInfo, err)
	if err != nil {
		return err
	}

	return nil
}

func (olc *OCILnetConfigurator) GetNetInterfacesInSubnet(subnetCIDR string) ([]NetInterface, error) {
	_, subnet, err := net.ParseCIDR(subnetCIDR)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse subnetCIDR %v  with error : %v", subnetCIDR, err)
	}
	var matchingInterfaces []NetInterface

	// Get all network interfaces
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("Failed to get net interfaces on host with error : %v", err)
	}

	for _, intf := range interfaces {
		// If some problem is there with interface, we will continue to check others
		addrs, err := intf.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			ip, _, _ := net.ParseCIDR(addr.String())
			// Check if the IP address falls within the specified subnet
			if ip.To4() != nil && subnet.Contains(ip) {
				matchingInterfaces = append(matchingInterfaces, NetInterface{
					InterfaceName: intf.Name,
					InterfaceIPv4: ip.String(),
				})
			}
		}
	}

	return matchingInterfaces, nil
}

func (olc *OCILnetConfigurator) GetLnetInfoByLnetLabel(lnetLabel string) (NetInfo, error) {
	var netInfo NetInfo
	existingConfiguredLnetInfo, err := olc.ExecuteCommandOnWorkerNode(fmt.Sprintf(SHOW_CONFIGURED_LNET, lnetLabel))
	if err != nil {
		return netInfo, fmt.Errorf("Failed to get existing configured lnet information with error : %v", err)
	}

	err = yaml.Unmarshal([]byte(existingConfiguredLnetInfo), &netInfo)
	if err != nil {
		return netInfo, fmt.Errorf("Failed to parse lnet information with error : %v", err)
	}
	return netInfo, nil
}

func (olc *OCILnetConfigurator) ConfigureLnet(logger *zap.SugaredLogger, interfacesInLustreSubnet []NetInterface, lnetLabel string, netInfo NetInfo) error {
	logger.Infof("Existing lnet information : %v", netInfo)

	//Find Already active and stale interfaces
	var staleLnetInterfaces []string

	for i, interfaceInLustreSubnet := range interfacesInLustreSubnet {
		for _, net := range netInfo.Net {
			for _, localNI := range net.LocalNI {
				if existingInterfaceName, ok := localNI.Interfaces[0]; ok && localNI.Status == "up" &&
					existingInterfaceName == interfaceInLustreSubnet.InterfaceName {

					if localNI.NID == fmt.Sprintf("%s@%s", interfaceInLustreSubnet.InterfaceIPv4, lnetLabel) {
						interfacesInLustreSubnet[i].LnetConfigured = true
					} else {
						//interface name matches but NIDS dont match, this needs to be deleted before new entry can be added with correct nid
						staleLnetInterfaces = append(staleLnetInterfaces, existingInterfaceName)
					}

				} else if existingInterfaceName, ok = localNI.Interfaces[0]; ok && localNI.Status == "down" {
					//Lnet interface is down, can happen when VNIC is detached
					staleLnetInterfaces = append(staleLnetInterfaces, existingInterfaceName)
				}
			}
		}
	}

	if len(staleLnetInterfaces) > 0 {
		logger.Infof("Deleting stale lnet interfaces identified : %v", staleLnetInterfaces)

		_, err := olc.ExecuteCommandOnWorkerNode(fmt.Sprintf(DELETE_LNET_INTERFACE, lnetLabel))
		if err != nil {
			return fmt.Errorf("Failed to delete stale lnet interface %v", staleLnetInterfaces)
		}
	}
	for _, interfaceInLustreSubnet := range interfacesInLustreSubnet {
		//Lnet configuration is needed if its not already configured or we have deleted lnet in previous step because of stale interfaces.
		if !interfaceInLustreSubnet.LnetConfigured || len(staleLnetInterfaces) > 0 {
			_, err := olc.ExecuteCommandOnWorkerNode(fmt.Sprintf(CONFIGURE_LNET_INTERFACE, lnetLabel, interfaceInLustreSubnet.InterfaceName))
			if err != nil {
				return fmt.Errorf("Lnet configuration failed for interface %s.", interfaceInLustreSubnet.InterfaceName)
			}
		}
	}
	return nil
}

func (olc *OCILnetConfigurator) VerifyLnetConfiguration(logger *zap.SugaredLogger, interfacesInLustreSubnet []NetInterface, lnetLabel string, netInfo NetInfo, err error) error {
	logger.Infof("Verifying lnet configuration.")
	//Get already configured lnet interfaces
	netInfo, err = olc.GetLnetInfoByLnetLabel(lnetLabel)
	if err != nil {
		return err
	}

	logger.Infof("Lnet information post configuration : %v", netInfo)

	var idenfiedActiveInterface bool
	for _, interfaceInLustreSubnet := range interfacesInLustreSubnet {
		for _, net := range netInfo.Net {
			for _, localNI := range net.LocalNI {
				if interfaceName, ok := localNI.Interfaces[0]; ok &&
					interfaceName == interfaceInLustreSubnet.InterfaceName {
					if localNI.Status == "up" {
						idenfiedActiveInterface = true
					} else {
						logger.With(zap.Error(err)).Errorf("Lnet for nid %v , interface %v is in %v status.", localNI.NID, interfaceName, localNI.Status)
					}
				}
			}
		}
	}

	//Verification step is kept non breaking currently and will provide information in case of issues during verification
	//we will allow flow to continue, if lnet doesn't become active till mount then mount will fail and retry will happen
	if idenfiedActiveInterface {
		logger.Infof("Lnet configuration completed & Verified.")
	} else {
		logger.Error("No active lnet interface is identified.")
	}
	return nil
}

func (olc *OCILnetConfigurator) IsLustreClientPackagesInstalled(logger *zap.SugaredLogger) bool {

	_, err := olc.ExecuteCommandOnWorkerNode(LFS_VERSION_COMMAND)
	if err != nil {
		logger.With(zap.Error(err)).Errorf("Error occured while performing Lustre Client package check using command %v. Error : %v", LFS_VERSION_COMMAND, err)
		return false
	}
	return true
}

/*
IsLnetActive takes lnetLabel (ex. tcp0, tcp1) as input and tries to check if at leaset one lnet interface is active to consider lnet as active.
It returns true when active lnet interface is identified else returns false singling down lnet.
*/
func (ls *LnetService) IsLnetActive(logger *zap.SugaredLogger, lnetLabel string) bool {
	logger.Debugf("Trying to check status of lnet")
	//Get already configured lnet interfaces
	netInfo, err := ls.Configurator.GetLnetInfoByLnetLabel(lnetLabel)
	if err != nil {
		logger.With(zap.Error(err)).Errorf("Failed to get lnet info for lnet :  %v", lnetLabel)
		return false
	}

	logger.Infof("Identified lnet information : %v", netInfo)

	var idenfiedActiveInterface bool
	for _, net := range netInfo.Net {
		for _, localNI := range net.LocalNI {
			if localNI.Status == "up" {
				idenfiedActiveInterface = true
			} else {
				logger.With(zap.Error(err)).Errorf("Lnet for nid %v  is in %v status.", localNI.NID, localNI.Status)
			}
		}
	}
	if !idenfiedActiveInterface {
		logger.Error("No active lnet interface is identified.")
	}

	return idenfiedActiveInterface
}

func (olc *OCILnetConfigurator) ExecuteCommandOnWorkerNode(args ...string) (string, error) {
	
	command := exec.Command("chroot-bash", args...)

	output, err := command.CombinedOutput()

	if err != nil {
		return string(output), fmt.Errorf("Command failed: %v\nOutput: %v\n", args, string(output))
	}
	return string(output), nil
}

func (ls *LnetService) ApplyLustreParameters(logger *zap.SugaredLogger, lustreParamsJson string) error {
	if lustreParamsJson == "" {
		logger.Debug("No lustre parameters specified.")
		return nil
	}
	var lustreParams []Parameter

	err := json.Unmarshal([]byte(lustreParamsJson), &lustreParams)

	if err != nil {
		return err
	}

	for _, param := range lustreParams {
		for key, value := range param {
			logger.Infof("Applying lustre param %s=%s", key, fmt.Sprintf("%v", value))
			_, err := ls.Configurator.ExecuteCommandOnWorkerNode(fmt.Sprintf(LCTL_SET_PARAM, key, fmt.Sprintf("%v", value)))
			if err != nil {
				return err
			}
		}
	}
	logger.Infof("Successfully applied lustre parameters.")
	return nil
}

func isValidShellInput(input string) bool {
	// Check for no spaces
	if strings.Contains(input, " ") {
		return false
	}
	// List of forbidden characters
	forbiddenChars := []string{";", "&", "|", "<", ">", "(", ")", "`", "'", "\"","$","!"}
	for _, char := range forbiddenChars {
		if strings.Contains(input, char) {
			return false
		}
	}
	return true
}
func  ValidateLustreParameters(logger *zap.SugaredLogger, lustreParamsJson string) error {
	if lustreParamsJson == "" {
		logger.Debug("No lustre parameters specified.")
		return nil
	}
	var lustreParams []Parameter

	err := json.Unmarshal([]byte(lustreParamsJson), &lustreParams)

	if err != nil {
		return err
	}

	var invalidParams []string

	for _, param := range lustreParams {
		for key, value := range param {
			logger.Infof("Validating lustre param %s=%s", key, fmt.Sprintf("%v", value))
			if !isValidShellInput(key) || !isValidShellInput(fmt.Sprintf("%v", value)) {
				invalidParams = append(invalidParams, fmt.Sprintf("%v=%v",key, value))
			}
		}
	}
	if len(invalidParams) > 0 {
		return fmt.Errorf("%v", strings.Join(invalidParams, ","))
	}
	logger.Infof("Successfully validated lustre parameters.")
	return nil
}

