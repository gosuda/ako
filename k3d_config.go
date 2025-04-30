package main

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

var globalConfig K3dConfig

const k3dConfigFileName = "k3d_config.yaml"

func getK3dConfigPath() string {
	return filepath.Join(k8sManifestFolder, k3dConfigFileName)
}

func init() {
	f, err := os.Open(getK3dConfigPath())
	if err != nil {
		return
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	if err := decoder.Decode(&globalConfig); err != nil {
		return
	}
}

func saveK3dConfig() error {
	f, err := os.Create(getK3dConfigPath())
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := yaml.NewEncoder(f)
	if err := encoder.Encode(globalConfig); err != nil {
		return err
	}

	return nil
}
