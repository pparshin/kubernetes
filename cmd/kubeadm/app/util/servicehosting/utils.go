package servicehosting

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"

	kubeadmconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"
)

type LocalConfig struct {
	KubeAPIServerAdvertiseAddressEndpoint string `yaml:"KubeAPIServerAdvertiseAddressEndpoint"`
}

var loadedCfg *LocalConfig

// IsServiceHostedControlPlane returns true if the control plane is bootstrapped on top of unix services.
func IsServiceHostedControlPlane() bool {
	_, err := os.Stat(kubeadmconstants.GetServiceHostedConfigFilepath())
	return err == nil
}

// MarkControlPlaneAsServiceHosted creates a local configuration file to mark this node as service-hosted.
func MarkControlPlaneAsServiceHosted(cfg *LocalConfig) error {
	err := writeToFile(cfg, kubeadmconstants.GetServiceHostedConfigFilepath())
	if err != nil {
		return fmt.Errorf("failed to mark the node as service-hosted: %v", err)
	}

	return nil
}

func LoadServiceHostedConfig() (*LocalConfig, error) {
	if loadedCfg != nil {
		return loadedCfg, nil
	}

	content, err := ioutil.ReadFile(kubeadmconstants.GetServiceHostedConfigFilepath())
	if err != nil {
		return nil, err
	}

	var cfg LocalConfig
	err = yaml.Unmarshal(content, &cfg)
	if err != nil {
		return nil, err
	}
	loadedCfg = &cfg

	return loadedCfg, nil
}

func writeToFile(cfg *LocalConfig, filename string) error {
	content, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	dir := filepath.Dir(filename)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0600); err != nil {
			return err
		}
	}

	if err := ioutil.WriteFile(filename, content, 0600); err != nil {
		return err
	}

	return nil
}
