package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"
)

func createK3dRegistry(name string) error {
	cmd := exec.Command("k3d", "registry", "create", name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

type K3dRegistry struct {
	Name         string   `json:"name"`
	Role         string   `json:"role"`
	Image        string   `json:"image"`
	Env          []string `json:"env"`
	Cmd          []string `json:"Cmd"`
	PortMappings struct {
		Five000TCP []struct {
			HostIP   string `json:"HostIp"`
			HostPort string `json:"HostPort"`
		} `json:"5000/tcp"`
	} `json:"portMappings"`
	Restart       bool      `json:"restart"`
	Created       time.Time `json:"created"`
	RuntimeLabels struct {
		K3DCluster        string `json:"k3d.cluster"`
		K3DRegistryHost   string `json:"k3d.registry.host"`
		K3DRegistryHostIP string `json:"k3d.registry.hostIP"`
		K3DRole           string `json:"k3d.role"`
		K3DVersion        string `json:"k3d.version"`
	} `json:"runtimeLabels"`
	Networks   []string `json:"Networks"`
	ExtraHosts any      `json:"ExtraHosts"`
	ServerOpts struct {
		KubeAPI struct {
			Port    string `json:"Port"`
			Binding struct {
				HostIP   string `json:"HostIp"`
				HostPort string `json:"HostPort"`
			} `json:"Binding"`
		} `json:"kubeAPI"`
	} `json:"serverOpts"`
	AgentOpts struct {
	} `json:"agentOpts"`
	GPURequest string `json:"GPURequest"`
	Memory     string `json:"Memory"`
	State      struct {
		Running bool   `json:"Running"`
		Status  string `json:"Status"`
		Started string `json:"Started"`
	} `json:"State"`
	IP struct {
		IP     string `json:"IP"`
		Static bool   `json:"Static"`
	} `json:"IP"`
	K3DEntrypoint bool `json:"K3dEntrypoint"`
}

type K3dRegistryList []K3dRegistry

func getK3dRegistry(name string) (*K3dRegistry, error) {
	cmd := exec.Command("k3d", "registry", "get", name, "-o", "json")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var registries K3dRegistryList
	if err := json.Unmarshal(output, &registries); err != nil {
		return nil, err
	}

	if len(registries) == 0 {
		return nil, fmt.Errorf("registry %s not found", name)
	}

	registry := registries[0]

	return &registry, nil
}

func getK3dRegistries() (K3dRegistryList, error) {
	cmd := exec.Command("k3d", "registry", "ls", "-o", "json")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var registries K3dRegistryList
	if err := json.Unmarshal(output, &registries); err != nil {
		return nil, err
	}

	return registries, nil
}

func deleteK3dRegistry(name string) error {
	cmd := exec.Command("k3d", "registry", "delete", name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func createK3dCluster(name string, agents int, registry string, loadBalancerPortMap map[int]int) error {
	args := []string{"cluster", "create", "--name", name, "--image", registry, "--agents", fmt.Sprintf("%d", agents)}
	for hostPort, containerPort := range loadBalancerPortMap {
		args = append(args, "-p", fmt.Sprintf("%d:%d@loadbalancer", hostPort, containerPort))
	}
	cmd := exec.Command("k3d", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func addK3dClusterPort(name string, hostPort int, containerPort int) error {
	cmd := exec.Command("k3d", "cluster", "port", "edit", name, "--port-add", fmt.Sprintf("%d:%d@loadbalancer", hostPort, containerPort))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func deleteK3dCluster(name string) error {
	cmd := exec.Command("k3d", "cluster", "delete", name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
