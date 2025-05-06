package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
)

const (
	K3dClusterPrefix = "k3d-"
)

func inputK3dRegistryName() (string, error) {
	name := ""
	if err := survey.AskOne(&survey.Input{
		Message: "Enter the name of the registry",
		Default: "my-registry",
	}, &name, survey.WithValidator(survey.Required)); err != nil {
		return "", err
	}

	if name == "" {
		return "", fmt.Errorf("registry name cannot be empty")
	}

	return name, nil
}

func selectK3dRegistryNames() ([]string, error) {
	registries, err := getK3dRegistries()
	if err != nil {
		return nil, err
	}

	candidates := make([]string, 0, len(registries))
	for _, registry := range registries {
		candidates = append(candidates, registry.Name)
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no registries found")
	}

	var selected []string
	if err := survey.AskOne(&survey.MultiSelect{
		Message: "Select the registry(s) to delete",
		Options: candidates,
	}, &selected, survey.WithValidator(survey.Required)); err != nil {
		return nil, err
	}

	if len(selected) == 0 {
		return nil, fmt.Errorf("no registries selected")
	}

	return selected, nil
}

func selectK3dRegistryName() (string, error) {
	registries, err := getK3dRegistries()
	if err != nil {
		return "", err
	}

	candidates := make([]string, 0, len(registries))
	for _, registry := range registries {
		candidates = append(candidates, registry.Name)
	}

	if len(candidates) == 0 {
		return "", fmt.Errorf("no registries found")
	}

	var selected string
	if err := survey.AskOne(&survey.Select{
		Message: "Select the registry to delete",
		Options: candidates,
	}, &selected, survey.WithValidator(survey.Required)); err != nil {
		return "", err
	}

	return selected, nil
}

func selectK3dRegistryForCluster() (string, error) {
	registries, err := getK3dRegistries()
	if err != nil {
		return "", err
	}

	candidates := make([]string, 0, len(registries))
	for _, registry := range registries {
		candidates = append(candidates, registry.Name)
	}

	if len(candidates) == 0 {
		return "", fmt.Errorf("no registries found")
	}

	var selected string
	if err := survey.AskOne(&survey.Select{
		Message: "Select the registry to use for the cluster",
		Options: candidates,
	}, &selected, survey.WithValidator(survey.Required)); err != nil {
		return "", err
	}

	value := ""
	for _, registry := range registries {
		if registry.Name == selected {
			value = registry.Name + ".localhost:" + registry.PortMappings.Five000TCP[0].HostPort
			break
		}
	}

	return value, nil
}

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

func inputK3dClusterName() (string, error) {
	name := ""
	if err := survey.AskOne(&survey.Input{
		Message: "Enter the name of the cluster",
		Default: "my-cluster",
	}, &name, survey.WithValidator(survey.Required)); err != nil {
		return "", err
	}

	if name == "" {
		return "", fmt.Errorf("cluster name cannot be empty")
	}

	return name, nil
}

func inputK3dClusterAgents() (int, error) {
	agents := 1
	if err := survey.AskOne(&survey.Input{
		Message: "Enter the number of agents",
		Default: "1",
	}, &agents, survey.WithValidator(survey.Required)); err != nil {
		return 0, err
	}

	if agents < 0 {
		return 0, fmt.Errorf("number of agents cannot be negative")
	}

	return agents, nil
}

func inputK3dClusterLoadBalancerPortMap() (map[int]int, error) {
	loadBalancerPortMapInput := ""
	if err := survey.AskOne(&survey.Input{
		Message: "Enter the load balancer port mapping (hostPort:containerPort, comma separated)",
		Default: "80:80,443:443",
	}, &loadBalancerPortMapInput, survey.WithValidator(survey.Required)); err != nil {
		return nil, err
	}

	loadBalancerPortMap := make(map[int]int)
	pairs := strings.Split(loadBalancerPortMapInput, ",")
	for _, pair := range pairs {
		ports := strings.Split(pair, ":")
		if len(ports) != 2 {
			return nil, fmt.Errorf("invalid port mapping: %s", pair)
		}

		hostPort := 0
		containerPort := 0
		if _, err := fmt.Sscanf(ports[0], "%d", &hostPort); err != nil {
			return nil, fmt.Errorf("invalid host port: %s", ports[0])
		}
		if _, err := fmt.Sscanf(ports[1], "%d", &containerPort); err != nil {
			return nil, fmt.Errorf("invalid container port: %s", ports[1])
		}

		loadBalancerPortMap[hostPort] = containerPort
	}

	return loadBalancerPortMap, nil
}

func createK3dCluster(name string, agents int, registry string, loadBalancerPortMap map[int]int) error {
	args := []string{"cluster", "create", name, "--registry-use", registry, "--agents", fmt.Sprintf("%d", agents)}
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
	cmd := exec.Command("k3d", "cluster", "edit", name, "--port-add", fmt.Sprintf("%d:%d@loadbalancer", hostPort, containerPort))
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

type K3dClusterInfo struct {
	Name    string `json:"name"`
	Network struct {
		Name       string `json:"name"`
		ID         string `json:"id"`
		IsExternal bool   `json:"isExternal"`
		Ipam       struct {
			IPPrefix string `json:"ipPrefix"`
			Managed  bool   `json:"Managed"`
		} `json:"ipam"`
		Members any `json:"Members"`
	} `json:"network"`
	Nodes []struct {
		Name         string   `json:"name"`
		Role         string   `json:"role"`
		Image        string   `json:"image"`
		Volumes      []string `json:"volumes"`
		Env          []string `json:"env"`
		Cmd          any      `json:"Cmd"`
		PortMappings map[string][]struct {
			HostIP   string `json:"HostIp"`
			HostPort string `json:"HostPort"`
		} `json:"portMappings,omitempty"`
		Restart       bool      `json:"restart"`
		Created       time.Time `json:"created"`
		RuntimeLabels struct {
			K3DCluster                string `json:"k3d.cluster"`
			K3DClusterImageVolume     string `json:"k3d.cluster.imageVolume"`
			K3DClusterNetwork         string `json:"k3d.cluster.network"`
			K3DClusterNetworkExternal string `json:"k3d.cluster.network.external"`
			K3DClusterNetworkID       string `json:"k3d.cluster.network.id"`
			K3DClusterNetworkIprange  string `json:"k3d.cluster.network.iprange"`
			K3DClusterToken           string `json:"k3d.cluster.token"`
			K3DClusterURL             string `json:"k3d.cluster.url"`
			K3DRole                   string `json:"k3d.role"`
			K3DServerLoadbalancer     string `json:"k3d.server.loadbalancer"`
			K3DVersion                string `json:"k3d.version"`
		} `json:"runtimeLabels,omitempty"`
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
		K3DEntrypoint  bool `json:"K3dEntrypoint"`
		RuntimeLabels0 struct {
			K3DCluster                string `json:"k3d.cluster"`
			K3DClusterImageVolume     string `json:"k3d.cluster.imageVolume"`
			K3DClusterNetwork         string `json:"k3d.cluster.network"`
			K3DClusterNetworkExternal string `json:"k3d.cluster.network.external"`
			K3DClusterNetworkID       string `json:"k3d.cluster.network.id"`
			K3DClusterNetworkIprange  string `json:"k3d.cluster.network.iprange"`
			K3DClusterToken           string `json:"k3d.cluster.token"`
			K3DClusterURL             string `json:"k3d.cluster.url"`
			K3DRole                   string `json:"k3d.role"`
			K3DServerAPIHost          string `json:"k3d.server.api.host"`
			K3DServerAPIHostIP        string `json:"k3d.server.api.hostIP"`
			K3DServerAPIPort          string `json:"k3d.server.api.port"`
			K3DServerLoadbalancer     string `json:"k3d.server.loadbalancer"`
			K3DVersion                string `json:"k3d.version"`
		} `json:"runtimeLabels,omitempty"`
	} `json:"nodes"`
	InitNode        any    `json:"InitNode"`
	ImageVolume     string `json:"imageVolume"`
	ServersRunning  int    `json:"serversRunning"`
	ServersCount    int    `json:"serversCount"`
	AgentsRunning   int    `json:"agentsRunning"`
	AgentsCount     int    `json:"agentsCount"`
	HasLoadbalancer bool   `json:"hasLoadbalancer"`
}

type K3dClusterInfoList []K3dClusterInfo

func SelectK3dClusterName() (string, error) {
	clusters, err := getK3dClusters()
	if err != nil {
		return "", err
	}

	candidates := make([]string, 0, len(clusters))
	for _, cluster := range clusters {
		candidates = append(candidates, cluster.Name)
	}

	if len(candidates) == 0 {
		return "", fmt.Errorf("no clusters found")
	}

	var selected string
	if err := survey.AskOne(&survey.Select{
		Message: "Select the cluster to delete",
		Options: candidates,
	}, &selected, survey.WithValidator(survey.Required)); err != nil {
		return "", err
	}

	return selected, nil
}

func selectK3dClusterNames() ([]string, error) {
	clusters, err := getK3dClusters()
	if err != nil {
		return nil, err
	}

	candidates := make([]string, 0, len(clusters))
	for _, cluster := range clusters {
		candidates = append(candidates, cluster.Name)
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no clusters found")
	}

	var selected []string
	if err := survey.AskOne(&survey.MultiSelect{
		Message: "Select the cluster(s) to delete",
		Options: candidates,
	}, &selected, survey.WithValidator(survey.Required)); err != nil {
		return nil, err
	}

	if len(selected) == 0 {
		return nil, fmt.Errorf("no clusters selected")
	}

	return selected, nil
}

func getK3dClusterInfo(name string) (*K3dClusterInfo, error) {
	cmd := exec.Command("k3d", "cluster", "get", name, "-o", "json")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var clusters K3dClusterInfoList
	if err := json.Unmarshal(output, &clusters); err != nil {
		return nil, err
	}

	if len(clusters) == 0 {
		return nil, fmt.Errorf("cluster %s not found", name)
	}

	cluster := clusters[0]

	return &cluster, nil
}

func getK3dClusters() (K3dClusterInfoList, error) {
	cmd := exec.Command("k3d", "cluster", "ls", "-o", "json")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var clusters K3dClusterInfoList
	if err := json.Unmarshal(output, &clusters); err != nil {
		return nil, err
	}

	return clusters, nil
}
