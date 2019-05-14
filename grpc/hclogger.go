package grpc

import (
	"bytes"
	"fmt"
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/pcore/px"
	"github.com/lyraproj/pcore/utils"
)

type hclogLogger struct {
	log hclog.Logger
}

func NewHclogLogger(log hclog.Logger) px.Logger {
	return &hclogLogger{log}
}

func (l *hclogLogger) Log(level px.LogLevel, args ...px.Value) {
	w := bytes.NewBufferString(``)
	for _, arg := range args {
		px.ToString3(arg, w)
	}
	l.hcLog(level, w.String())
}

func (l *hclogLogger) Logf(level px.LogLevel, format string, args ...interface{}) {
	l.hcLog(level, format, args...)
}

func (l *hclogLogger) hcLog(level px.LogLevel, format string, args ...interface{}) {
	switch level {
	case px.ERR, px.ALERT, px.CRIT, px.EMERG:
		if l.log.IsError() {
			l.log.Error(fmt.Sprintf(format, args...))
		}
	case px.WARNING:
		if l.log.IsWarn() {
			l.log.Warn(fmt.Sprintf(format, args...))
		}
	case px.INFO, px.NOTICE:
		if l.log.IsInfo() {
			l.log.Info(fmt.Sprintf(format, args...))
		}
	case px.DEBUG:
		if l.log.IsDebug() {
			l.log.Debug(fmt.Sprintf(format, args...))
		}
	}
}

func (l *hclogLogger) LogIssue(issue issue.Reported) {
	utils.Fprintln(os.Stderr, issue.String())
}
