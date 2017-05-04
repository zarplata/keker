package main

import (
	"fmt"
	"strconv"
	"time"

	docopt "github.com/docopt/docopt-go"
	"github.com/kovetskiy/lorg"
)

var (
	logger  *lorg.Log
	version string = "[manual build]"
)

func main() {
	usage := `zarplata-keker

Usage:
	zarplata-keker [options]

Options:
    -l --listen <address>     Address which service will be listen 
                              for incoming requests [default: 0.0.0.0:3498]
                                                                        
    -r --read-buffer <size>   Max size of messages [default: 51200]
    -w --write-buffer <size>  Max size of messages [default: 102400]
    -t --timeout <seconds>    Timeout for read and write deadlines and for 
                              handling PING-PONG frames [default: 60]
    -c --cache <time>         Set TTL for message in cache [default: 5s]
					                        
	-v --verbose <level>      Logging level [default: ERROR].
`

	cmdArguments, err := docopt.Parse(usage, nil, true, version, false)
	if err != nil {
		panic(err)
	}

	logger = setupLogger(cmdArguments["--verbose"].(string))

	subscribers := make(map[string][]chan []byte)

	timeout, err := time.ParseDuration(
		fmt.Sprintf("%ss", cmdArguments["--timeout"].(string)),
	)
	if err != nil {
		logger.Fatalf(
			"unable to parse duration: %s",
			err.Error(),
		)
	}

	readBufferSize, err := strconv.Atoi(
		cmdArguments["--read-buffer"].(string),
	)
	if err != nil {
		logger.Fatalf(
			"unable to parse read buffer size: %s",
			err.Error(),
		)
	}

	writeBufferSize, err := strconv.Atoi(
		cmdArguments["--write-buffer"].(string),
	)
	if err != nil {
		logger.Fatalf(
			"unable to parse write buffer size: %s",
			err.Error(),
		)
	}

	cachedMessageTTL, err := time.ParseDuration(
		cmdArguments["--cache"].(string),
	)
	if err != nil {
		logger.Fatalf(
			"unable to parse TTL for caced messages, reason %s",
			err.Error(),
		)
	}

	service := newService(
		subscribers,
		readBufferSize,
		writeBufferSize,
		timeout,
		cmdArguments["--verbose"].(string),
		newCache(cachedMessageTTL),
	)

	service.run(
		cmdArguments["--listen"].(string),
	)
}

func setupLogger(
	logLevel string,
) *lorg.Log {
	level := lorg.LevelInfo

	switch logLevel {
	case "ERROR":
		level = lorg.LevelError

	case "DEBUG":
		level = lorg.LevelDebug

	default:
		level = lorg.LevelInfo
	}

	newLogger := lorg.NewLog()

	defaultFormatting := lorg.NewFormat(
		`${time:15:04:05} ${level} %s [${file}:${line}]`,
	)
	newLogger.SetFormat(defaultFormatting)
	newLogger.SetLevel(level)

	return newLogger
}
