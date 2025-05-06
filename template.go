package main

import (
	"os"
	"strconv"
	"strings"
	"text/template"
)

var templateFuncMap = template.FuncMap{
	"toUpper": func(s string) string {
		return strings.ToUpper(s)
	},
	"toLower": func(s string) string {
		return strings.ToLower(s)
	},
	"toPascal": func(s string) string {
		return strings.ToUpper(s[:1]) + s[1:]
	},
	"quote": func(s string) string {
		return strconv.Quote(s)
	},
}

func writeTemplate2File(filename string, tmp string, data any) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	t, err := template.New("template").Funcs(templateFuncMap).Parse(tmp)
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
