package sse

// Log 常用日志结构体注入接口
type Log interface {
	Info(args ...interface{})
	Infoln(args ...interface{})
	Infof(format string, args ...interface{})

	Debug(args ...interface{})
	Debugln(args ...interface{})
	Debugf(format string, args ...interface{})

	Warn(args ...interface{})
	Warnln(args ...interface{})
	Warnf(format string, args ...interface{})

	Error(args ...interface{})
	Errorln(args ...interface{})
	Errorf(format string, args ...interface{})
}
