package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"gopkg.in/natefinch/lumberjack.v2"
)

func Setup(conf *Config) {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.DurationFieldUnit = time.Nanosecond
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	_ = os.MkdirAll(conf.Path, os.ModePerm)

	var level zerolog.Level
	if conf.TraceLevel {
		level = zerolog.TraceLevel
	} else {
		level = zerolog.DebugLevel
	}

	var writer io.Writer

	if conf.StdoutOnly {
		writer = os.Stdout
	} else {
		writer = zerolog.MultiLevelWriter(
			&lumberjack.Logger{
				Filename:   conf.Path + conf.Filename,
				MaxSize:    conf.MaxSize,    // megabytes
				MaxAge:     conf.MaxAge,     // days
				MaxBackups: conf.MaxBackups, // files
				Compress:   conf.Compress,
			},
			zerolog.ConsoleWriter{
				Out:        os.Stdout,
				TimeFormat: time.RFC3339Nano,
			},
		)
	}

	log.Logger = zerolog.New(writer).
		With().
		Timestamp().
		Caller().
		Logger().
		Level(level)
}
