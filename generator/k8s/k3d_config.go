package k8s

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type K3dConfig struct {
	Cluster        string `yaml:"cluster"`
	Namespace      string `yaml:"namespace"`
	LocalRegistry  string `yaml:"localRegistry"`
	RemoteRegistry string `yaml:"remoteRegistry"`
}

var (
	GlobalConfig          K3dConfig
	globalConfigNotExists bool
)

const k3dConfigFileName = ".ako/k3d_config.yaml"

func getK3dConfigPath() string {
	return filepath.Join(k8sManifestFolder, k3dConfigFileName)
}

func init() {
	f, err := os.Open(getK3dConfigPath())
	if err != nil {
		if os.IsNotExist(err) {
			globalConfigNotExists = true
		}
		return
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	if err := decoder.Decode(&GlobalConfig); err != nil {
		return
	}
}

func isNotExistsK3dConfig() bool {
	if globalConfigNotExists {
		return true
	}

	if GlobalConfig.Cluster == "" || GlobalConfig.Namespace == "" {
		return true
	}

	return false
}

func SaveK3dConfig() error {
	if err := os.MkdirAll(k8sManifestFolder, os.ModePerm); err != nil {
		return err
	}

	f, err := os.Create(getK3dConfigPath())
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := yaml.NewEncoder(f)
	if err := encoder.Encode(GlobalConfig); err != nil {
		return err
	}

	return nil
}
