package k8s

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"

	"github.com/gosuda/ako/util/table"
)

type HelmSearchItem struct {
	URL         string `json:"url"`
	Version     string `json:"version"`
	AppVersion  string `json:"app_version"`
	Description string `json:"description"`
	Repository  struct {
		URL  string `json:"url"`
		Name string `json:"name"`
	} `json:"repository"`
}

type HelmSearchResult []HelmSearchItem

func (result HelmSearchResult) Print() {
	tbl := table.NewTableBuilder("Name", "URL", "Version", "AppVersion", "Description", "Repository URL", "Repository Name")
	for _, item := range result {
		tbl.AppendRow(filepath.Base(item.URL), item.URL, item.Version, item.AppVersion, item.Description, item.Repository.URL, item.Repository.Name)
	}
	tbl.Print()
}

func searchHelmChart(repo string, query string) (HelmSearchResult, error) {
	output, err := exec.Command("helm", "search", repo, query, "-o", "json").Output()
	if err != nil {
		return nil, err
	}

	result := HelmSearchResult{}
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func SelectHelmChartFromRepo(repo string, query string) (HelmSearchItem, error) {
	charts, err := searchHelmChart(repo, query)
	if err != nil {
		return HelmSearchItem{}, err
	}

	if len(charts) == 0 {
		return HelmSearchItem{}, fmt.Errorf("no helm chart found for %s", repo)
	}

	candidates := make([]string, len(charts))
	for i, chart := range charts {
		candidates[i] = fmt.Sprintf("%s: %s => %s", filepath.Base(chart.URL), chart.Version, chart.Description)
	}

	var selectedName string
	if err := survey.AskOne(&survey.Select{
		Message: "Select a chart",
		Options: candidates,
	}, &selectedName); err != nil {
		return HelmSearchItem{}, err
	}

	selectedIndex := -1
	for i, candidate := range candidates {
		if candidate == selectedName {
			selectedIndex = i
			break
		}
	}

	if selectedIndex == -1 {
		return HelmSearchItem{}, fmt.Errorf("invalid selection: %s", selectedName)
	}

	selectedChart := charts[selectedIndex]
	return selectedChart, nil
}

func InstallHelmChart(chart HelmSearchItem, releaseName string, namespace string) error {
	cmd := exec.Command("helm", "install", releaseName, filepath.Base(chart.URL), "--repo", chart.Repository.URL, "--namespace", namespace)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

type HelmReleaseListItem struct {
	Name       string `json:"name"`
	Namespace  string `json:"namespace"`
	Revision   string `json:"revision"`
	Updated    string `json:"updated"`
	Status     string `json:"status"`
	Chart      string `json:"chart"`
	AppVersion string `json:"app_version"`
}

type HelmReleaseListResult []HelmReleaseListItem

func ListHelmReleases(namespace string) (HelmReleaseListResult, error) {
	output, err := exec.Command("helm", "list", "--namespace", namespace, "-o", "json").Output()
	if err != nil {
		return nil, err
	}

	result := HelmReleaseListResult{}
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func SelectHelmRelease(releases HelmReleaseListResult) (HelmReleaseListItem, error) {
	if len(releases) == 0 {
		return HelmReleaseListItem{}, fmt.Errorf("no helm release found")
	}

	candidates := make([]string, len(releases))
	for i, release := range releases {
		candidates[i] = fmt.Sprintf("%s: %s => %s", release.Name, release.Chart, release.Status)
	}

	var selectedName string
	if err := survey.AskOne(&survey.Select{
		Message: "Select a release",
		Options: candidates,
	}, &selectedName); err != nil {
		return HelmReleaseListItem{}, err
	}

	selectedIndex := -1
	for i, candidate := range candidates {
		if candidate == selectedName {
			selectedIndex = i
			break
		}
	}

	if selectedIndex == -1 {
		return HelmReleaseListItem{}, fmt.Errorf("invalid selection: %s", selectedName)
	}

	selectedRelease := releases[selectedIndex]
	return selectedRelease, nil
}

func UninstallHelmChart(releaseName string, namespace string) error {
	cmd := exec.Command("helm", "uninstall", releaseName, "--namespace", namespace)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

type HelmRepoListItem struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type HelmRepoListResult []HelmRepoListItem

func (result HelmRepoListResult) Print() {
	tbl := table.NewTableBuilder("Name", "URL")
	for _, item := range result {
		tbl.AppendRow(item.Name, item.URL)
	}
	tbl.Print()
}

func ListHelmRepos() (HelmRepoListResult, error) {
	output, err := exec.Command("helm", "repo", "list", "-o", "json").Output()
	if err != nil {
		return nil, err
	}

	result := HelmRepoListResult{}
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func AddHelmRepo(name string, url string) error {
	_, err := exec.Command("helm", "repo", "add", name, url).Output()
	if err != nil {
		return err
	}
	return nil
}

func SelectHelmRepo(repos HelmRepoListResult) (HelmRepoListItem, error) {
	if len(repos) == 0 {
		return HelmRepoListItem{}, fmt.Errorf("no helm repo found")
	}

	candidates := make([]string, len(repos))
	for i, repo := range repos {
		candidates[i] = fmt.Sprintf("%s: %s", repo.Name, repo.URL)
	}

	var selectedName string
	if err := survey.AskOne(&survey.Select{
		Message: "Select a repo",
		Options: candidates,
	}, &selectedName); err != nil {
		return HelmRepoListItem{}, err
	}

	selectedIndex := -1
	for i, candidate := range candidates {
		if candidate == selectedName {
			selectedIndex = i
			break
		}
	}

	if selectedIndex == -1 {
		return HelmRepoListItem{}, fmt.Errorf("invalid selection: %s", selectedName)
	}

	selectedRepo := repos[selectedIndex]
	return selectedRepo, nil
}

func RemoveHelmRepo(name string) error {
	_, err := exec.Command("helm", "repo", "remove", name).Output()
	if err != nil {
		return err
	}
	return nil
}
