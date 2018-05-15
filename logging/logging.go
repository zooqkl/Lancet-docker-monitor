/*
Copyright Yunphant Corp. All Rights Reserved.
*/

package lancetlogging

// lancet日志输出的接口，切换日志输出模块需要在logger.go中实现此接口
type Logger interface {
	Debug(args ...interface{})

	Debugf(format string, args ...interface{})

	Info(args ...interface{})

	Infof(format string, args ...interface{})

	Notice(args ...interface{})

	Noticef(format string, args ...interface{})

	Warn(args ...interface{})

	Warnf(format string, args ...interface{})

	Error(args ...interface{})

	Errorf(format string, args ...interface{})

	Fatal(args ...interface{})

	Fatalf(format string, args ...interface{})
}
