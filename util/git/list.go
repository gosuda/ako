package git

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/AlecAivazis/survey/v2"
)

const (
	FileTypeUntracked = "untracked"
	FileTypeModified  = "modified"
	FileTypeDeleted   = "deleted"
)

var fileTypeShortCutMap = map[string]string{
	FileTypeUntracked: "NEW",
	FileTypeModified:  "MOD",
	FileTypeDeleted:   "DEL",
}

var fileTypeLongCutMap = map[string]string{
	"NEW": "untracked",
	"MOD": "modified",
	"DEL": "deleted",
}

type UnstagedFile struct {
	Type string
	Path string
}

const (
	unstagedFileStringFormat = "{%s} %s"
)

func (u *UnstagedFile) String() string {
	return fmt.Sprintf(unstagedFileStringFormat, fileTypeShortCutMap[u.Type], u.Path)
}

func ListUnstagedFiles() ([]*UnstagedFile, error) {
	cmd := exec.Command("git", "diff", "--name-only")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	files := make([]*UnstagedFile, 0, len(lines))
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}

		t := FileTypeModified
		if _, err := os.Stat(line); os.IsNotExist(err) {
			t = FileTypeDeleted
		}

		file := &UnstagedFile{
			Type: t,
			Path: line,
		}
		files = append(files, file)
	}

	return files, nil
}

func ListUntrackedFiles() ([]*UnstagedFile, error) {
	cmd := exec.Command("git", "ls-files", "--others", "--exclude-standard")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	files := make([]*UnstagedFile, 0, len(lines))
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		file := &UnstagedFile{
			Type: FileTypeUntracked,
			Path: line,
		}
		files = append(files, file)
	}

	return files, nil
}

func ListUnstagedFilesWithType() ([]*UnstagedFile, error) {
	files, err := ListUnstagedFiles()
	if err != nil {
		return nil, err
	}

	untrackedFiles, err := ListUntrackedFiles()
	if err != nil {
		return nil, err
	}

	files = append(files, untrackedFiles...)

	return files, nil
}

func SelectUnstagedFilesToStage(files []*UnstagedFile) ([]*UnstagedFile, error) {
	if len(files) == 0 {
		return nil, nil
	}

	candidates := make([]string, 0, len(files))
	for _, file := range files {
		candidates = append(candidates, file.String())
	}

	var selected []string
	if err := survey.AskOne(&survey.MultiSelect{
		Message: "Select files to stage:",
		Options: candidates,
		Help:    "Use space to select, enter to confirm",
	}, &selected); err != nil {
		return nil, err
	}

	selectedFiles := make([]*UnstagedFile, 0, len(selected))
	for _, s := range selected {
		for _, file := range files {
			if s == file.String() {
				selectedFiles = append(selectedFiles, file)
				break
			}
		}
	}

	return selectedFiles, nil
}

func StageFiles(files []*UnstagedFile) error {
	if len(files) == 0 {
		return nil
	}

	filePaths := make([]string, 0, len(files))
	for _, file := range files {
		filePaths = append(filePaths, file.Path)
	}

	if err := AddGitFiles(filePaths...); err != nil {
		return err
	}

	return nil
}

func UnstageFiles(files []*UnstagedFile) error {
	if len(files) == 0 {
		return nil
	}

	filePaths := make([]string, 0, len(files))
	for _, file := range files {
		filePaths = append(filePaths, file.Path)
	}

	if err := ResetGitFiles(filePaths...); err != nil {
		return err
	}

	return nil
}
