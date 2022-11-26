package cmd

import (
	"fmt"
	"io"
	_log "log"
	"os"
	"path"
	"time"
)

type logger struct {
	file        *os.File
	error       *_log.Logger
	info        *_log.Logger
	conn        *_log.Logger
	connEnabled bool
}

func (l *logger) init() error {
	var output io.Writer
	var connOutput io.Writer

	if config.log.enabled {
		if !folderExists(config.log.dir) {
			if err := os.MkdirAll(config.log.dir, 0750); err != nil && !os.IsExist(err) {
				return err
			}
		}

		year, month, day := time.Now().Date()
		logFileName := fmt.Sprintf("%v_%v_%v.log", year, int(month), day)

		logFilePath := path.Join(config.log.dir, logFileName)
		if !fileExists(logFilePath) {
			if _, err := os.Create(logFilePath); err != nil {
				return err
			}
		}

		var err error
		l.file, err = os.OpenFile(logFilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			return err
		}

		if config.quiet {
			output = l.file
		} else {
			output = io.MultiWriter(os.Stdout, l.file)
		}
	} else if config.quiet {
		output = io.Discard
	} else {
		output = os.Stdout
	}

	l.connEnabled = config.log.connections
	if l.connEnabled {
		connOutput = output
	} else {
		connOutput = io.Discard
	}

	flags := _log.Ldate | _log.Ltime | _log.Lmsgprefix
	l.error = _log.New(output, "[error] ", flags)
	l.info = _log.New(output, "[info] ", flags)
	l.conn = _log.New(connOutput, "[connection] ", flags)

	return nil
}

func (l *logger) close() {
	if l.file != nil {
		if err := l.file.Close(); err != nil {
			fmt.Printf("error while closing log file: %v\n", err)
		}
	}
}
