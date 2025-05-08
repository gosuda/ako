package ai

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const (
	aiCopilotTemplateFileName = ".vscode/settings.json"
	aiCopilotVsCodeTemplate   = `{
    "github.copilot.chat.commitMessageGeneration.instructions": [
        {
            "file": "commit_message_rule.txt"
        },
        {
            "text": "Use conventional commit message format."
        }
    ]
}`
)

var aiCopilotTemplateList = map[string][]map[string]string{
	"github.copilot.chat.commitMessageGeneration.instructions": {
		{
			"file": "commit_message_rule.txt",
		},
		{
			"text": "Use conventional commit message format.",
		},
	},
}

func CreateVsCodeCopilotSettings() error {
	path := ".vscode"
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}

	value := map[string]interface{}{}

	settingsFile := filepath.Join(path, "settings.json")
	f, err := os.Open(settingsFile)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}

		f, err = os.Create(settingsFile)
		if err != nil {
			return err
		}
	}
	defer func(f *os.File) {
		if err := f.Close(); err != nil {
			return
		}
	}(f)

	_ = json.NewDecoder(f).Decode(&value)

	for k, v := range aiCopilotTemplateList {
		value[k] = v
	}

	if err := os.Remove(settingsFile); err != nil {
		return err
	}

	f, err = os.Create(settingsFile)
	if err != nil {
		return err
	}
	defer func(f *os.File) {
		if err := f.Close(); err != nil {
			return
		}
	}(f)

	if err := json.NewEncoder(f).Encode(value); err != nil {
		return err
	}

	return nil
}
