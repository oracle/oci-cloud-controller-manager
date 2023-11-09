package osinfo

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

var OsName = ""

const (
	LinuxOsReleaseFile = "/host/etc/os-release"

	DebianOSName       = "Debian GNU/Linux"

	UbuntuOSName       = "Ubuntu"
)

func GetOsName() (name string) {
	if OsName != "" {
		return OsName
	}

	OsName =  parseLinuxReleaseFile(LinuxOsReleaseFile)
	return OsName;
}

func readLines(path string) ([]string, error) {
	inFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer inFile.Close()
	var lines []string
	scanner := bufio.NewScanner(inFile)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func parseLinuxReleaseFile(releaseFile string) (name string) {
	osName := ""
	lines, err := readLines(releaseFile)
	if err != nil {
		return osName
	}
	for _, line := range lines {
		if strings.HasPrefix(line, "NAME=") {
			tokens := strings.Split(line, "=")
			if len(tokens) > 1 {
				r := regexp.MustCompile(`[\w\s/]+`)
				osName = r.FindString(tokens[1])
			}
		}
	}
	return osName
}


func IsUbuntu() bool {
	return strings.EqualFold(UbuntuOSName, GetOsName())
}

func IsDebian() bool {
	return strings.EqualFold(DebianOSName, GetOsName())
}

func IsDebianOrUbuntu() bool {
	return IsUbuntu() || IsDebian()
}
