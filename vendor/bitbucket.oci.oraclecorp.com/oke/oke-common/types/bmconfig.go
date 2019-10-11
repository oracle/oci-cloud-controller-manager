package types

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// NewBMConfigV1 reads a config file path returning the config file
// contents
func NewBMConfigV1(cfgPath string) (*BMConfig, error) {
	f, err := os.Open(resolveTilda(cfgPath))
	if err != nil {
		return nil, fmt.Errorf("failed to open config file %s: %s", cfgPath, err)
	}

	bmc := new(BMConfig)

	dir := filepath.Dir(cfgPath)

	s := bufio.NewScanner(f)
	for s.Scan() {
		txt := s.Text()
		if strings.HasPrefix(txt, "[") {
			continue
		}

		parts := strings.Split(txt, "=")
		if len(parts) != 2 {
			continue
		}
		k, v := parts[0], parts[1]
		switch k {
		case "user":
			bmc.UserOCID = v
		case "fingerprint":
			bmc.Fingerprint = v
		case "tenancy":
			bmc.Tenancy = v
		case "region":
			bmc.Region = v
		case "key_file":
			val := resolveTilda(v)
			if !strings.HasPrefix(val, "/") {
				val = filepath.Join(dir, val)
			}

			bs, err := ioutil.ReadFile(val)
			if err != nil {
				return bmc, fmt.Errorf("unable to read file %s: %s", val, err)
			}
			// kept as bytes for marshalling
			bmc.PrivateKey = bs
		default:
			log.Printf("Unrecognized key %s=%s", k, v)
		}
	}
	return bmc, nil
}

func resolveTilda(path string) string {
	homepath := os.Getenv("HOME")
	val := regexp.MustCompile("^~/").ReplaceAllString(path, homepath+"/")
	return val
}

func (src *BMConfig) ToBMCCloudAuth() *BMCCloudAuth {
	return &BMCCloudAuth{
		User:        src.UserOCID,
		Fingerprint: src.Fingerprint,
		PrivateKey:  string(src.PrivateKey),
		Tenancy:     src.Tenancy,
		Region:      src.Region,
	}
}

func (bmc *BMConfig) ToProto() *BMConfig {
	if bmc == nil {
		return &BMConfig{}
	}
	bm := new(BMConfig)

	bm.Fingerprint = bmc.Fingerprint
	bm.UserOCID = bmc.UserOCID
	bm.PrivateKey = bmc.PrivateKey
	bm.Tenancy = bmc.Tenancy
	bm.Region = bmc.Region

	return bm
}

// Encode writes the BMConfig
func (bmc *BMConfig) EncodeToPath(path string) error {
	_, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("unable to encode to %s: %s", path, err)
	}

	return bmc.encodeToPath(path)
}

// encodetoPath creates oraclebmc config file and private_key files
func (bmc *BMConfig) encodeToPath(path string) error {
	configPath := filepath.Join(path, "config")
	privKeyPath := filepath.Join(path, "bmcs_api_key.pem")
	err := ioutil.WriteFile(privKeyPath, bmc.PrivateKey, 0600)
	if err != nil {
		return fmt.Errorf("failed to write api key %s: %s", privKeyPath, err)
	}

	file, err := os.Create(configPath)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(file, `[DEFAULT]
user=%s
fingerprint=%s
key_file=%s
tenancy=%s
region=%s
`, bmc.UserOCID, bmc.Fingerprint, privKeyPath, bmc.Tenancy, bmc.Region)

	return err
}
