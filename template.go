package main

import (
	"os"
	"text/template"
)

func writeTemplate2File(filename string, tmp string, data map[string]interface{}) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	t, err := template.New("template").Parse(tmp)
	if err != nil {
		return err
	}

	if err := t.Execute(file, data); err != nil {
		return err
	}

	if err := file.Sync(); err != nil {
		return err
	}

	return nil
}
