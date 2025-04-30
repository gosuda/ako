package main

import (
	"github.com/fatih/color"
	"github.com/rodaine/table"
)

type TableBuilder struct {
	tbl table.Table
}

func NewTableBuilder(column ...string) *TableBuilder {
	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()

	tbl := table.New(column)
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

	return &TableBuilder{
		tbl: tbl,
	}
}

func (tb *TableBuilder) AppendRow(data ...any) {
	tb.tbl.AddRow(data...)
}

func (tb *TableBuilder) Print() {
	tb.tbl.Print()
}
