package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"github.com/AlecAivazis/survey/v2"
)

const (
	gitBranchPrefixRelease  = "release"
	gitBranchPrefixStaging  = "staging"
	gitBranchPrefixDevelop  = "develop"
	gitBranchPrefixEpic     = "epic"
	gitBranchPrefixFeature  = "feature"
	gitBranchPrefixHotfix   = "hotfix"
	gitBranchPrefixPatch    = "patch"
	gitBranchPrefixBreak    = "break"
	gitBranchPrefixProposal = "proposal"
)

func initGit(initialBranchName string) error {
	cmd := exec.Command("git", "init", "-b", initialBranchName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func getGitBranchName() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(bytes.TrimSpace(output)), nil
}

func getGitBranchesWithPrefixSuffix(prefix, suffix string) ([]string, error) {
	cmd := exec.Command("git", "branch", "--list", fmt.Sprintf("'%v*%v'", prefix, suffix))
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	branches := bytes.Split(output, []byte("\n"))
	var result []string
	for _, branch := range branches {
		if len(branch) > 0 {
			result = append(result, string(bytes.TrimSpace(branch)))
		}
	}

	return result, nil
}

func constructSubBranchName(prefix string, superName string, name string) (string, error) {
	if name == "" {
		return prefix, nil
	}

	if superName == "" {
		return fmt.Sprintf("%s/%s", prefix, name), nil
	}

	return fmt.Sprintf("%s/%s/%s", prefix, superName, name), nil
}

func deconstructSubBranchName(branchName string) (string, string, string) {
	parts := bytes.Split([]byte(branchName), []byte("/"))
	switch len(parts) {
	case 3:
		return string(parts[0]), string(parts[1]), string(parts[2])
	case 2:
		return string(parts[0]), "", string(parts[1])
	case 1:
		return string(parts[0]), "", ""
	}

	return "", "", ""
}

func getGitSubBranchPrefix(prefix string) (string, error) {
	switch prefix {
	case gitBranchPrefixRelease:
		candidates := []string{gitBranchPrefixStaging, gitBranchPrefixHotfix}
		var subPrefix string
		if err := survey.AskOne(&survey.Select{
			Message: "Select sub branch type",
			Options: candidates,
		}, &subPrefix, survey.WithValidator(survey.Required)); err != nil {
			return "", err
		}

		return subPrefix, nil
	case gitBranchPrefixStaging:
		return gitBranchPrefixDevelop, nil
	case gitBranchPrefixDevelop:
		return gitBranchPrefixEpic, nil
	case gitBranchPrefixEpic:
		candidates := []string{gitBranchPrefixFeature, gitBranchPrefixPatch, gitBranchPrefixBreak}
		var subPrefix string
		if err := survey.AskOne(&survey.Select{
			Message: "Select sub branch type",
			Options: candidates,
		}, &subPrefix, survey.WithValidator(survey.Required)); err != nil {
			return "", err
		}

		return subPrefix, nil
	case gitBranchPrefixFeature, gitBranchPrefixPatch, gitBranchPrefixBreak:
		return gitBranchPrefixProposal, nil
	}

	return "", fmt.Errorf("invalid prefix: %s", prefix)
}

func makeGitSubBranchName(branchName string) (string, error) {
	prefix, _, name := deconstructSubBranchName(branchName)
	if prefix == "" {
		return "", fmt.Errorf("invalid branch name: %s", branchName)
	}

	subPrefix, err := getGitSubBranchPrefix(prefix)
	if err != nil {
		return "", err
	}

	switch prefix {
	case gitBranchPrefixRelease:
		return constructSubBranchName(subPrefix, "", "")
	case gitBranchPrefixStaging:
		return constructSubBranchName(subPrefix, "", "")
	case gitBranchPrefixDevelop:
		var scopeName string
		if err := survey.AskOne(&survey.Input{
			Message: "Enter the scope name:",
		}, &scopeName, survey.WithValidator(survey.Required)); err != nil {
			return "", err
		}

		return constructSubBranchName(subPrefix, "", scopeName)
	case gitBranchPrefixEpic:
		var featureName string
		if err := survey.AskOne(&survey.Input{
			Message: "Enter the work name:",
		}, &featureName, survey.WithValidator(survey.Required)); err != nil {
			return "", err
		}

		return constructSubBranchName(subPrefix, name, featureName)
	case gitBranchPrefixFeature, gitBranchPrefixPatch, gitBranchPrefixBreak:
		var featureName string
		if err := survey.AskOne(&survey.Input{
			Message: "Enter the proposal name:",
		}, &featureName, survey.WithValidator(survey.Required)); err != nil {
			return "", err
		}

		return constructSubBranchName(subPrefix, name, featureName)
	case gitBranchPrefixProposal:
		return "", fmt.Errorf("proposal branch cannot create sub branch")
	}

	return "", fmt.Errorf("invalid prefix: %s", prefix)
}

func switchGitBranchTo(branchName string) error {
	cmd := exec.Command("git", "switch", "-C", branchName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func getParentBranchName() ([]string, error) {
	current, err := getGitBranchName()
	if err != nil {
		return nil, err
	}

	currentPrefix, superName, _ := deconstructSubBranchName(current)
	parentPrefix := make([]string, 0)
	switch currentPrefix {
	case gitBranchPrefixRelease:
		return nil, fmt.Errorf("release branch cannot have parent branch")
	case gitBranchPrefixStaging:
		parentPrefix = []string{gitBranchPrefixRelease}
	case gitBranchPrefixDevelop:
		parentPrefix = []string{gitBranchPrefixStaging}
	case gitBranchPrefixEpic:
		parentPrefix = []string{gitBranchPrefixDevelop}
	case gitBranchPrefixFeature, gitBranchPrefixPatch, gitBranchPrefixBreak:
		parentPrefix = []string{gitBranchPrefixEpic}
	case gitBranchPrefixProposal:
		parentPrefix = []string{gitBranchPrefixFeature, gitBranchPrefixPatch, gitBranchPrefixBreak}
	default:
		return nil, fmt.Errorf("invalid prefix: %s", currentPrefix)
	}

	branches := make([]string, 0)
	for _, prefix := range parentPrefix {
		b, err := getGitBranchesWithPrefixSuffix(prefix, superName)
		if err != nil {
			return nil, err
		}

		branches = append(branches, b...)
	}

	return branches, nil
}

func getChildrenBranchName() ([]string, error) {
	current, err := getGitBranchName()
	if err != nil {
		return nil, err
	}

	currentPrefix, _, name := deconstructSubBranchName(current)
	childrenPrefix := make([]string, 0)
	switch currentPrefix {
	case gitBranchPrefixRelease:
		childrenPrefix = []string{gitBranchPrefixStaging}
	case gitBranchPrefixStaging:
		childrenPrefix = []string{gitBranchPrefixDevelop}
	case gitBranchPrefixDevelop:
		childrenPrefix = []string{gitBranchPrefixEpic}
	case gitBranchPrefixEpic:
		childrenPrefix = []string{gitBranchPrefixFeature, gitBranchPrefixPatch, gitBranchPrefixBreak}
	case gitBranchPrefixFeature, gitBranchPrefixPatch, gitBranchPrefixBreak:
		childrenPrefix = []string{gitBranchPrefixProposal}
	case gitBranchPrefixProposal:
		return nil, fmt.Errorf("proposal branch cannot have children branch")
	default:
		return nil, fmt.Errorf("invalid prefix: %s", currentPrefix)
	}

	branches := make([]string, 0)
	for _, prefix := range childrenPrefix {
		b, err := getGitBranchesWithPrefixSuffix(prefix+"/"+name, "")
		if err != nil {
			return nil, err
		}

		branches = append(branches, b...)
	}

	return branches, nil
}

func addGitFiles(files ...string) error {
	files = append([]string{"add"}, files...)
	cmd := exec.Command("git", files...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func commitGitFiles(message string) error {
	cmd := exec.Command("git", "commit", "-m", message)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func pushGitFiles() error {
	cmd := exec.Command("git", "push")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func pullGitFiles() error {
	cmd := exec.Command("git", "pull")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

/*
5. 커밋 메시지 컨벤션
프로젝트는 Conventional Commits 규칙을 따릅니다. 모든 커밋 메시지는 다음 형식을 준수합니다.

<타입>[선택적 범위][!]: <설명>

<타입> (필수): feat, fix, build, chore, ci, docs, style, refactor, perf, test 등 커밋의 성격을 나타냅니다.

[선택적 범위]: 커밋이 영향을 미치는 코드 영역(패키지, 모듈 등)을 명시합니다. 모노레포 환경에서는 범위를 사용하여 변경된 패키지나 영역을 명시하는 것이 충돌 예방 및 변경 추적에 도움이 될 수 있습니다. (예: feat(auth): ..., fix(ui-kit): ...). epic의 이름을 범위로 사용하는 것도 고려할 수 있습니다 (feat(new-payment-system): ...).

[!]: (선택적, 중요): 실제로 하위 호환성을 깨뜨리는 변경(Breaking Change) 이 있음을 나타냅니다. 타입 또는 범위 바로 뒤에 !를 추가하며, MAJOR 버전 변경을 유발합니다.

<설명> (필수): 변경 사항에 대한 간결한 설명 (현재형, 명령형, 첫 글자 소문자, 마침표 없음).
*/
func buildGitCommitMessage() (string, error) {
	commitType := ""
	if err := survey.AskOne(&survey.Select{
		Message: "Select commit type",
		Options: []string{"feat", "fix", "build", "chore", "ci", "docs", "style", "refactor", "perf", "test"},
	}, &commitType, survey.WithValidator(survey.Required)); err != nil {
		return "", err
	}

	var scope string
	if err := survey.AskOne(&survey.Input{
		Message: "Enter the scope name (optional):",
	}, &scope, survey.WithValidator(survey.Required)); err != nil {
		return "", err
	}

	var breakingChange bool
	if err := survey.AskOne(&survey.Confirm{
		Message: "Is this a breaking change?",
		Default: false,
	}, &breakingChange); err != nil {
		return "", err
	}

	var description string
	if err := survey.AskOne(&survey.Input{
		Message: "Enter the commit message:",
	}, &description, survey.WithValidator(survey.Required)); err != nil {
		return "", err
	}

	if breakingChange {
		scope = fmt.Sprintf("%s!", scope)
	}

	if scope == "" {
		return fmt.Sprintf("%s: %s", commitType, description), nil
	}

	return fmt.Sprintf("%s(%s): %s", commitType, scope, description), nil
}
