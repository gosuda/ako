package git

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/AlecAivazis/survey/v2"
)

func InputTag() (string, error) {
	latestTag, err := ListLatestTag(1)
	if err != nil || len(latestTag) == 0 {
		latestTag = []string{"<none>"}
	}

	value := ""
	if err := survey.AskOne(&survey.Input{
		Message: fmt.Sprintf("Enter the tag name [latest: %s]:", latestTag[0]),
	}, &value, survey.WithValidator(survey.Required)); err != nil {
		return "", err
	}

	value = strings.TrimSpace(value)

	const regularExpression = `^v\d+\.\d+\.\d+(-\w+.\d+)?$`
	regex := regexp.MustCompile(regularExpression)
	if !regex.MatchString(value) {
		return "", fmt.Errorf("invalid tag name: %s", value)
	}

	return value, nil
}

func InputTagMemo() (string, error) {
	value := ""
	if err := survey.AskOne(&survey.Input{
		Message: "Enter the tag memo:",
	}, &value, survey.WithValidator(survey.Required)); err != nil {
		return "", err
	}

	value = strings.TrimSpace(value)

	return value, nil
}

func SetTag(tag string, memo string) error {
	cmd := exec.Command("git", "tag", tag, "-m", memo)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func PushTag(tag string) error {
	cmd := exec.Command("git", "push", "origin", tag)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func DeleteTag(tag string) error {
	cmd := exec.Command("git", "tag", "-d", tag)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func DeleteRemoteTag(tag string) error {
	cmd := exec.Command("git", "push", ":"+tag)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func ListLatestTag(count int) ([]string, error) {
	// git for-each-ref --format="%(refname:strip=2)" --sort=-creatordate --count=10 refs/tags
	command := exec.Command("git", "for-each-ref", "--format=%(refname:strip=2)", "--sort=-creatordate", fmt.Sprintf("--count=%d", count), "refs/tags")
	output, err := command.Output()
	if err != nil {
		return nil, err
	}

	output = bytes.TrimSpace(output)
	scanner := bufio.NewScanner(bytes.NewReader(output))
	var tags []string
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 0 {
			tags = append(tags, line)
		}
	}

	return tags, nil
}

type TagInfo struct {
	Name string
	Date string
	Memo string
}

func ListTags() ([]TagInfo, error) {
	// git for-each-ref --format="[{(%(creatordate))}],[{(%(refname:strip=2))}],[{(%(subject))}]" --sort=-creatordate --count=10 refs/tags
	command := exec.Command("git", "for-each-ref", "--format=\"[{(%(creatordate))}],[{(%(refname:strip=2))}],[{(%(subject))}]\"", "--sort=-creatordate", "refs/tags")
	output, err := command.Output()
	if err != nil {
		return nil, err
	}

	output = bytes.TrimSpace(output)
	contents := make([][]byte, 0, 4)
	state := 0
	prev := 0
	for i, b := range output {
		switch state {
		case 0:
			if b == '[' && i+2 < len(output) && output[i+1] == '{' && output[i+2] == '(' {
				prev = i + 3
				state = 1
			}
		case 1:
			if b == ')' && i+2 < len(output) && output[i+1] == '}' && output[i+2] == ']' {
				contents = append(contents, output[prev:i])
				state = 0
			}
		}
	}

	tags := make([]TagInfo, len(contents))
	for i := 0; i < len(contents); i += 3 {
		if i+2 < len(contents) {
			idx := i / 3
			tags[idx].Date = string(contents[i])
			tags[idx].Name = string(contents[i+1])
			tags[idx].Memo = string(contents[i+2])
		}
	}

	return tags, nil
}
