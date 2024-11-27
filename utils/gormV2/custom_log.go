package gormV2

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	g "src/global"

	gormLog "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

// 自定义日志格式, 对 gorm 自带日志进行拦截重写
func createCustomGormLog(options ...Options) gormLog.Interface {
	var (
		infoStr      = "%s\n[info] "
		warnStr      = "%s\n[warn] "
		errStr       = "%s\n[error] "
		traceStr     = "%s\n[%.3fms] [rows:%d] %s"
		traceWarnStr = "%s %s\n[%.3fms] [rows:%d] %s"
		traceErrStr  = "%s %s\n[%.3fms] [rows:%d] %s"
	)
	logConf := gormLog.Config{
		SlowThreshold:             time.Second * g.VP.GetDuration("mysql.SlowThreshold"),
		LogLevel:                  gormLog.Warn,
		IgnoreRecordNotFoundError: true, // 忽略未找到记录的这种错误
		Colorful:                  false,
	}
	log := &logger{
		Config:       logConf,
		infoStr:      infoStr,
		warnStr:      warnStr,
		errStr:       errStr,
		traceStr:     traceStr,
		traceWarnStr: traceWarnStr,
		traceErrStr:  traceErrStr,
	}
	for _, val := range options {
		val.apply(log)
	}
	return log
}
// 主要重写了这个方法，为了能够添加logid
func printf(ctx context.Context, strFormat string, args ...interface{}) {
	logRes := fmt.Sprintf(strFormat, args...)
	logFlag := "mysql日志:"

	// 优先获取ctx传过来的log，因为带有logid，方便链路追踪
	log, ok := ctx.Value(g.MysqlLogKeyName).(*g.Log)
	if !ok {
		log = g.InitLog(0) // 无请求id 的情况
	}
	if strings.HasPrefix(strFormat, "[info]") || strings.HasPrefix(strFormat, "[traceStr]") {
		log.Info(logFlag, logRes)
	} else if strings.HasPrefix(strFormat, "[error]") || strings.HasPrefix(strFormat, "[traceErr]") {
		log.Error(logFlag, errors.New(logRes))
	} else if strings.HasPrefix(strFormat, "[warn]") || strings.HasPrefix(strFormat, "[traceWarn]") {
		log.Warn(logFlag, logRes)
	}
}

// 尝试从外部重写内部相关的格式化变量
type Options interface {
	apply(*logger)
}
type OptionFunc func(log *logger)

func (f OptionFunc) apply(log *logger) {
	f(log)
}

// 定义 6 个函数修改内部变量
func SetInfoStrFormat(format string) Options {
	return OptionFunc(func(log *logger) {
		log.infoStr = format
	})
}

func SetWarnStrFormat(format string) Options {
	return OptionFunc(func(log *logger) {
		log.warnStr = format
	})
}

func SetErrStrFormat(format string) Options {
	return OptionFunc(func(log *logger) {
		log.errStr = format
	})
}

func SetTraceStrFormat(format string) Options {
	return OptionFunc(func(log *logger) {
		log.traceStr = format
	})
}
func SetTracWarnStrFormat(format string) Options {
	return OptionFunc(func(log *logger) {
		log.traceWarnStr = format
	})
}

func SetTracErrStrFormat(format string) Options {
	return OptionFunc(func(log *logger) {
		log.traceErrStr = format
	})
}

type logger struct {
	gormLog.Config
	infoStr, warnStr, errStr            string
	traceStr, traceErrStr, traceWarnStr string
}

// LogMode log mode
func (l *logger) LogMode(level gormLog.LogLevel) gormLog.Interface {
	newlogger := *l
	newlogger.LogLevel = level
	return &newlogger
}

// Info print info
func (l logger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormLog.Info {
		printf(ctx, l.infoStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

// Warn print warn messages
func (l logger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormLog.Warn {
		printf(ctx, l.warnStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

// Error print error messages
func (l logger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormLog.Error {
		printf(ctx, l.errStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

// Trace print sql message
func (l logger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= gormLog.Silent {
		return
	}
	elapsed := time.Since(begin)
	sql, rows := fc()
	if err != nil && l.LogLevel >= gormLog.Error { // 有错误时的记录格式
		printf(ctx, l.traceErrStr, utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, rows, sql)
	} else if elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= gormLog.Warn { // sql运行慢于设置的阈值
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		printf(ctx, l.traceWarnStr, utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, rows, sql)
	} else if l.LogLevel == gormLog.Info{
		printf(ctx, l.traceStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, rows, sql)
	}
}
