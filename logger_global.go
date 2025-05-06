package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
)

const (
	loggerDependencyLumberjack = `gopkg.in/natefinch/lumberjack.v2`
	loggerDependencyZerolog    = `github.com/rs/zerolog/log`
	loggerDependencyZap        = `go.uber.org/zap`
)

const (
	loggerLibraryZerolog = "zerolog"
	loggerLibraryZap     = "zap"
	loggerLibrarySlog    = "slog"
)

func selectLoggerLibrary() (string, error) {
	candidates := []string{
		loggerLibraryZerolog,
		loggerLibraryZap,
		loggerLibrarySlog,
	}

	var loggerLibrary string
	if err := survey.AskOne(&survey.Select{
		Message: "Select the prefer logger library:",
		Options: candidates,
	}, &loggerLibrary); err != nil {
		return "", err
	}

	return loggerLibrary, nil
}

const (
	loggerWriterFilename = "logger_writer.go"
	loggerWriterTemplate = `package logger

import (
	"bufio"
	"io"
	"os"

	"gopkg.in/natefinch/lumberjack.v2"
)

func NewMultiplexerWriter(writers ...io.Writer) io.Writer {
	if len(writers) == 0 {
		return io.Discard
	}

	if len(writers) == 1 {
		return writers[0]
	}

	multiWriter := io.MultiWriter(writers...)
	return multiWriter
}

func NewBufferedStdoutWriter() *BufferedWriter {
	return NewBufferedWriter(os.Stdout)
}

func NewBufferedStderrWriter() *BufferedWriter {
	return NewBufferedWriter(os.Stderr)
}

type BufferedWriter struct {
	underlying io.WriteCloser
	writer     *bufio.Writer
}

func NewBufferedWriter(writer io.WriteCloser) *BufferedWriter {
	bw := bufio.NewWriter(writer)
	return &BufferedWriter{
		underlying: writer,
		writer:     bw,
	}
}

func (bw *BufferedWriter) Write(p []byte) (n int, err error) {
	n, err = bw.writer.Write(p)
	if err != nil {
		return n, err
	}

	return n, nil
}

func (bw *BufferedWriter) Sync() error {
	if err := bw.writer.Flush(); err != nil {
		return err
	}
	return nil
}

func (bw *BufferedWriter) Close() error {
	if err := bw.writer.Flush(); err != nil {
		return err
	}
	if err := bw.underlying.Close(); err != nil {
		return err
	}
	return nil
}

func NewRotationFileWriter() *lumberjack.Logger {
	const (
		logFile    = "app.log"
		MaxSize    = 50 // megabytes
		MaxBackups = 3
		MaxAge     = 28 // days
	)
	writer := &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    MaxSize,
		MaxBackups: MaxBackups,
		MaxAge:     MaxAge,
		Compress:   true,
	}

	return writer
}
`
	loggerInitializerFilename    = "logger_initializer.go"
	loggerZapInitializerTemplate = `package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func init() {
	writer := NewBufferedStderrWriter()
	logger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		writer,
		zapcore.DebugLevel,
	))
	zap.ReplaceGlobals(logger)
}
`
	loggerZerologInitializerTemplate = `package logger

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	writer := NewBufferedStderrWriter()
	log.Logger = zerolog.New(writer).
		Level(zerolog.DebugLevel).
		With().Caller().Timestamp().Logger()
}
`
	loggerSlogInitializerTemplate = `package logger

import (
	"log/slog"
)

func init() {
	writer := NewBufferedStderrWriter()
	logger := slog.New(slog.NewJSONHandler(writer, nil))
	slog.SetDefault(logger)
}
`
)

func createLoggerWriterFile(selectedLoggerLibrary string) error {
	if err := os.MkdirAll(filepath.Join("pkg", "global", "logger"), 0755); err != nil {
		return err
	}

	writerFilePath := filepath.Join("pkg", "global", "logger", loggerWriterFilename)
	if err := os.WriteFile(writerFilePath, []byte(loggerWriterTemplate), 0644); err != nil {
		return err
	}

	initFilePath := filepath.Join("pkg", "global", "logger", loggerInitializerFilename)
	var initTemplate string
	var dependency string
	switch selectedLoggerLibrary {
	case loggerLibraryZerolog:
		initTemplate = loggerZerologInitializerTemplate
		dependency = loggerDependencyZerolog
	case loggerLibraryZap:
		initTemplate = loggerZapInitializerTemplate
		dependency = loggerDependencyZap
	case loggerLibrarySlog:
		initTemplate = loggerSlogInitializerTemplate
		dependency = ""
	default:
		return fmt.Errorf("unsupported logger library: %s", selectedLoggerLibrary)
	}
	if err := os.WriteFile(initFilePath, []byte(initTemplate), 0644); err != nil {
		return err
	}

	if err := getGoModule(loggerDependencyLumberjack); err != nil {
		return fmt.Errorf("getGoModule: %w", err)
	}

	if len(dependency) > 0 {
		if err := getGoModule(dependency); err != nil {
			return fmt.Errorf("getGoModule: %w", err)
		}
	}

	return nil
}
