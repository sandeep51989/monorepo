package logger

import (
	"io"
	"os"
	"runtime"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

var (
	ZapLogger *zap.Logger
)

type WriteSyncer struct {
	io.Writer
}

func (ws WriteSyncer) Sync() error {
	return nil
}

func InitLogging(logName string) {
	cfg := zap.NewProductionConfig()
	cfg.DisableCaller = true
	cfg.DisableStacktrace = true
	cfg.Encoding = "json"
	cfg.EncoderConfig.TimeKey = "timestamp"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.OutputPaths = []string{logName}
	sw := getWriteSyncer(logName)

	l, err := cfg.Build(SetOutput(sw, cfg))
	if err != nil {
		panic(err)
	}
	defer l.Sync()

	ZapLogger = l
}

// SetOutput replaces existing Core with new, that writes to passed WriteSyncer.
func SetOutput(ws zapcore.WriteSyncer, conf zap.Config) zap.Option {
	var enc zapcore.Encoder
	switch conf.Encoding {
	case "json":
		enc = zapcore.NewJSONEncoder(conf.EncoderConfig)
	case "console":
		enc = zapcore.NewConsoleEncoder(conf.EncoderConfig)
	default:
		panic("unknown encoding")
	}
	if runtime.GOOS == "windows" {
		return zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			syncer := zap.CombineWriteSyncers(os.Stdout, getWriteSyncer(conf.OutputPaths[0]))
			return zapcore.NewCore(enc, syncer, conf.Level)
		})
	} else {
		return zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			return zapcore.NewCore(enc, ws, conf.Level)
		})
	}
}

func getWriteSyncer(logName string) zapcore.WriteSyncer {
	var ioWriter = &lumberjack.Logger{
		Filename:   logName,
		MaxSize:    10, // MB
		MaxBackups: 10, // number of backups
		MaxAge:     28, //days
		LocalTime:  true,
		Compress:   false, // disabled by default
	}
	var sw = WriteSyncer{
		ioWriter,
	}
	return sw
}
