package k8s

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/AlecAivazis/survey/v2"
)

func MakeCmdDepthToName(cmd ...string) string {
	if len(cmd) == 0 {
		return ""
	}

	if len(cmd) == 1 {
		return cmd[0]
	}

	return fmt.Sprintf("%s-%s", MakeCmdDepthToName(cmd[1:]...), cmd[0])
}

func GetK8sManifestList(prefix string) ([]string, error) {
	if prefix == "" {
		prefix = k8sManifestFolder
	}

	dir, err := os.ReadDir(prefix)
	if err != nil {
		return nil, err
	}

	k8sManifestList := map[string]struct{}{}

	for _, entry := range dir {
		if entry.IsDir() {
			entries, err := GetK8sManifestList(filepath.Join(prefix, entry.Name()))
			if err != nil {
				return nil, err
			}

			for _, entry := range entries {
				k8sManifestList[filepath.ToSlash(entry)] = struct{}{}
			}
			continue
		}

		if strings.HasSuffix(entry.Name(), ".yaml") || strings.HasSuffix(entry.Name(), ".yml") {
			k8sManifestList[filepath.ToSlash(filepath.Join(prefix, entry.Name()))] = struct{}{}
		}
	}

	delete(k8sManifestList, k8sManifestFolder)

	var k8sManifestListSlice []string
	for k8sManifest := range k8sManifestList {
		k8sManifestListSlice = append(k8sManifestListSlice, k8sManifest)
	}

	slices.Sort(k8sManifestListSlice)

	if prefix == k8sManifestFolder {
		for i := range k8sManifestListSlice {
			k8sManifestListSlice[i] = strings.TrimPrefix(k8sManifestListSlice[i], prefix+"/")
		}
	}

	return k8sManifestListSlice, nil
}

func SelectK8sManifest() ([]string, error) {
	candidates, err := GetK8sManifestList(k8sManifestFolder)
	if err != nil {
		return nil, err
	}

	selected := make([]string, 0)
	if err := survey.AskOne(&survey.MultiSelect{
		Message: "Select k8s manifests to apply:",
		Options: candidates,
		Help:    "Use space to select, enter to confirm",
	}, &selected, survey.WithValidator(survey.Required)); err != nil {
		return nil, err
	}

	if len(selected) == 0 {
		return nil, fmt.Errorf("no k8s manifests found")
	}

	for i := range selected {
		selected[i] = filepath.Join(k8sManifestFolder, selected[i])
	}

	return selected, nil
}

func ApplyK8sManifest(file string) error {
	cmd := exec.Command("kubectl", "apply", "-f", file, "--context", K3dClusterPrefix+GlobalConfig.Cluster)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func RunK8sGetPods() error {
	cmd := exec.Command("kubectl", "get", "pods", "--context", K3dClusterPrefix+GlobalConfig.Cluster, "-n", GlobalConfig.Namespace, "-o", "wide")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func RunK8sGetServices() error {
	cmd := exec.Command("kubectl", "get", "services", "--context", K3dClusterPrefix+GlobalConfig.Cluster, "-n", GlobalConfig.Namespace, "-o", "wide")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func RunK8sGetDeployments() error {
	cmd := exec.Command("kubectl", "get", "deployments", "--context", K3dClusterPrefix+GlobalConfig.Cluster, "-n", GlobalConfig.Namespace, "-o", "wide")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func RunK8sGetIngress() error {
	cmd := exec.Command("kubectl", "get", "ingress", "--context", K3dClusterPrefix+GlobalConfig.Cluster, "-n", GlobalConfig.Namespace, "-o", "wide")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
