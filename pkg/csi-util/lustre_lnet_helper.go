package csi_util

import (
	"fmt"
	"net"
	"os/exec"
	"strings"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"

	"github.com/oracle/oci-cloud-controller-manager/pkg/util/osinfo"
)

const (
	CHROOT_BASH_COMMAND             = "chroot-bash"
	LOAD_LNET_KERNEL_MODULE_COMMAND = "modprobe lnet"
	CONFIGURE_LNET_KERNEL_SERVICE_COMMAND = "lnetctl lnet configure"
	SHOW_CONFIGURED_LNET = "lnetctl net show --net %s"
	DELETE_LNET_INTERFACE = "lnetctl net del --net %s"
	CONFIGURE_LNET_INTERFACE = "lnetctl net add --net %s --if %s --peer-timeout 180 --peer-credits 120 --credits 1024"
	RPM_PACKAGE_QUERY = "rpm -q %s"
	DPKG_PACKAGE_QEURY = "dpkg -s %s"
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
	InterfaceIPv4 string
	InterfaceName string
	LnetConfigured bool
}

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
	}
	//last part in volume handle which is fsname should start with "/"
	if !strings.HasPrefix(splits[len(splits)-1], "/") {
		return false, lnetLabel
	}
	return true, lnetLabel
}

func SetupLnet(logger *zap.SugaredLogger, lustreSubnetCIDR string, lnetLabel string)  error {

	logger.With("LustreSubnetCidr", lustreSubnetCIDR).With("LnetLabel", lnetLabel).Info("Lnet setup started.")

	//Find net interfaces in lnet subnet on worker node
	interfacesInLustreSubnet, err := getNetInterfacesInSubnet(lustreSubnetCIDR)
	if err != nil {
		return err
	}
	if len(interfacesInLustreSubnet) == 0 {
		return fmt.Errorf("No VNIC identified on worker node to configure lnet.")
	}
	logger.Infof("Net interfaces are identified for lnet configuration: %v", interfacesInLustreSubnet)

	//Lustre client installation state is kept non breaking currently and we still go ahead and try loading kernel module and start lnet kernel service
	//If any of these fail then we provide information to customer on missing packages.
	missingLustrePackages, lustreClientPacakagesInstalled := isLustreClientPackagesInstalled(logger)

	//Load lnet kernel module
	_, err = executeCommandOnWorkerNode(LOAD_LNET_KERNEL_MODULE_COMMAND)
	if err != nil {
		if !lustreClientPacakagesInstalled {
			return fmt.Errorf("Failed to load lnet kernel module with error : %v. Please make sure that following lustre client packages are installed on worker nodes : %v", err, missingLustrePackages)
		}
		return fmt.Errorf("Failed to load lnet kernel module with error %v", err)
	}

	//Configure lnet kernel service
	_, err = executeCommandOnWorkerNode(CONFIGURE_LNET_KERNEL_SERVICE_COMMAND)
	if err != nil {
		if !lustreClientPacakagesInstalled {
			return fmt.Errorf("Failed to configure lnet kernel service with error : %v. Please make sure that following lustre client packages are installed on worker nodes : %v",err, missingLustrePackages)
		}
		return fmt.Errorf("Failed to configure lnet kernel service with error : %v", err)
	}

 	//get existing lnet configuration
	var netInfo NetInfo
	netInfo, err = getLnetInfoByLnetLabel(lnetLabel)
	if err != nil {
		return err
	}

	//Configure lnet if its not configured already for requried interfaces
	err = configureLnet(logger, interfacesInLustreSubnet, lnetLabel, netInfo)
	if err != nil {
		return err
	}

	//Verify lnet configuration
	err = verifyLnetConfiguration(logger, interfacesInLustreSubnet, lnetLabel, netInfo, err)
	if err != nil {
		return err
	}

	return nil
}


func getNetInterfacesInSubnet(subnetCIDR string) ([]NetInterface, error) {
	_, subnet, err := net.ParseCIDR(subnetCIDR)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse subnetCIDR %v  with error : %v",subnetCIDR, err)
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
			ip,_,_ := net.ParseCIDR(addr.String())
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

func getLnetInfoByLnetLabel(lnetLabel string) (NetInfo, error) {
	var netInfo NetInfo
	existingConfiguredLnetInfo, err := executeCommandOnWorkerNode(fmt.Sprintf(SHOW_CONFIGURED_LNET, lnetLabel))
	if err != nil {
		return  netInfo, fmt.Errorf("Failed to get existing configured lnet information with error : %v", err)
	}

	err = yaml.Unmarshal([]byte(existingConfiguredLnetInfo), &netInfo)
	if err != nil {
		return  netInfo, fmt.Errorf("Failed to parse lnet information with error : %v", err)
	}
	return netInfo, nil
}

func configureLnet(logger *zap.SugaredLogger, interfacesInLustreSubnet []NetInterface, lnetLabel string, netInfo NetInfo) (error) {
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


	if len(staleLnetInterfaces) > 0  {
		logger.Infof("Deleting stale lnet interfaces identified : %v", staleLnetInterfaces)

		_, err := executeCommandOnWorkerNode(fmt.Sprintf(DELETE_LNET_INTERFACE, lnetLabel))
		if err != nil {
			return fmt.Errorf("Failed to delete stale lnet interface %v", staleLnetInterfaces)
		}
	}
	for _, interfaceInLustreSubnet := range interfacesInLustreSubnet {
		//Lnet configuration is needed if its not already configured or we have deleted lnet in previous step because of stale interfaces.
		if !interfaceInLustreSubnet.LnetConfigured || len(staleLnetInterfaces) > 0 {
			_, err := executeCommandOnWorkerNode(fmt.Sprintf(CONFIGURE_LNET_INTERFACE, lnetLabel, interfaceInLustreSubnet.InterfaceName))
			if err != nil {
				return fmt.Errorf("Lnet configuration failed for interface %s.", interfaceInLustreSubnet.InterfaceName)
			}
		}
	}
	return nil
}

func verifyLnetConfiguration(logger *zap.SugaredLogger, interfacesInLustreSubnet []NetInterface, lnetLabel string, netInfo NetInfo, err error) error {
	logger.Infof("Verifying lnet configuration.")
	//Get already configured lnet interfaces
	netInfo, err = getLnetInfoByLnetLabel(lnetLabel)
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

func isLustreClientPackagesInstalled(logger *zap.SugaredLogger) ([]string, bool) {

	lustrePackages := []string{ "lustre-client-modules-dkms", "lustre-client-utils"}

	var missingLustrePackages []string

	for _, pkgName := range lustrePackages {
		if !checkPackageInstalled(pkgName) {
			missingLustrePackages = append(missingLustrePackages, pkgName)
		}
	}
	if len(missingLustrePackages) > 0 {
		logger.Error("Following lustre packages are not installed on worker ndoes : %v", missingLustrePackages)
		return missingLustrePackages, false
	}
	return nil, true
}

func checkPackageInstalled(pkgName string) bool {
	var err error
	var result string

	if osinfo.IsDebianOrUbuntu() {
		result, err = executeCommandOnWorkerNode(fmt.Sprintf(DPKG_PACKAGE_QEURY, pkgName))
		if err == nil && !strings.Contains(result,"Status: install ok installed") {
			return false
		}
	} else {
		_, err = executeCommandOnWorkerNode(fmt.Sprintf(RPM_PACKAGE_QUERY, pkgName))
	}
	// When packages are not found command fails with non zero exit code and err will not be nil.
	return err == nil
}

/*
IsLnetActive takes lnetLabel (ex. tcp0, tcp1) as input and tries to check if at leaset one lnet interface is active to consider lnet as active.
It returns true when active lnet interface is identified else returns false singling down lnet.
 */
func IsLnetActive(logger *zap.SugaredLogger, lnetLabel string) bool  {
	logger.Debugf("Trying to check status of lnet")
	//Get already configured lnet interfaces
	netInfo, err := getLnetInfoByLnetLabel(lnetLabel)
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

func executeCommandOnWorkerNode(args ...string) (string, error) {
	command := exec.Command(CHROOT_BASH_COMMAND, args...)

	output, err := command.CombinedOutput()

	if err != nil {
		return string(output), fmt.Errorf("Command failed: %v\nOutput: %v\n", args, string(output))
	}
	return string(output), nil
}
