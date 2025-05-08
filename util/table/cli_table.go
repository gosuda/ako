package table

import (
	"bufio"
	"os"

	"github.com/fatih/color"
	"github.com/rodaine/table"
)

type TableBuilder struct {
	buf *bufio.Writer
	tbl table.Table
}

func NewTableBuilder(column ...any) *TableBuilder {
	buffer := bufio.NewWriter(os.Stdout)

	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()

	tbl := table.New(column...)
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt).WithWriter(buffer)

	return &TableBuilder{
		buf: buffer,
		tbl: tbl,
	}
}

func (tb *TableBuilder) AppendRow(data ...any) {
	tb.tbl.AddRow(data...)
}

func (tb *TableBuilder) Print() {
	tb.tbl.Print()
	_ = tb.buf.Flush()
}
