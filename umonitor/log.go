package umonitor

import (
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func InitLogger(file, level string) (logger *zap.Logger) {
	writeSyncer := getLogWriter(file)
	encoder := getEncoder()
	logLevel := zapcore.InfoLevel
	if level == "debug" {
		logLevel = zapcore.DebugLevel
	}

	core := zapcore.NewCore(encoder, writeSyncer, logLevel)

	logger = zap.New(core, zap.AddCaller())
	return logger
	//sugarLogger = logger.Sugar()
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	return zapcore.NewConsoleEncoder(encoderConfig)
}

func getLogWriter(log string) zapcore.WriteSyncer {
	/*
		Filename: 日志文件的位置
		MaxSize：在进行切割之前，日志文件的最大大小（以MB为单位）
		MaxBackups：保留旧文件的最大个数
		MaxAges：保留旧文件的最大天数
		Compress：是否压缩/归档旧文件
	*/
	lumberJackLogger := &lumberjack.Logger{
		Filename:   log,
		MaxSize:    10,
		MaxBackups: 5,
		MaxAge:     30,
		LocalTime:  true,
		Compress:   true,
	}
	return zapcore.AddSync(lumberJackLogger)
}
