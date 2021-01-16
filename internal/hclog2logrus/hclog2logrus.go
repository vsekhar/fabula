// Package hclog2logrus redirects hclog log entries (produced by Serf) to the
// default logrus logger.
package hclog2logrus

import (
	"io/ioutil"
	stdlog "log"

	hclog "github.com/hashicorp/go-hclog"
	log "github.com/sirupsen/logrus"
)

var hclogToLogrusLevels map[hclog.Level]log.Level = map[hclog.Level]log.Level{
	hclog.Trace: log.TraceLevel,
	hclog.Debug: log.DebugLevel,
	hclog.Info:  log.InfoLevel,
	hclog.Warn:  log.WarnLevel,
	hclog.Error: log.ErrorLevel,
}

func init() {
	log.Info("initing")
	hclog.DefaultOptions.Output = ioutil.Discard
	hclog.DefaultOutput = ioutil.Discard
	l := hclog.NewInterceptLogger(hclog.DefaultOptions)
	l.RegisterSink(new(logrusSink))
	hclog.SetDefault(l)
}

type logrusSink struct{}

// Accept accepts a new log entry and emits it to logrus.
func (*logrusSink) Accept(name string, level hclog.Level, msg string, args ...interface{}) {
	if len(args)%2 != 0 {
		log.Error("expected pairs of args")
	}
	fields := make(log.Fields)
	if name != "" {
		fields["logname"] = name
	}
	for i := 0; i < len(args); i += 2 {
		fields[(args[i]).(string)] = args[i+1]
	}
	log.WithFields(fields).Log(hclogToLogrusLevels[level], msg)
}

// New returns a new stdlib logger for use with Serf in the Config object. This
// logger sends log entries via the hclog chain which is then intercepted and
// sent to Logrus.
func New() *stdlog.Logger {
	hopts := &hclog.StandardLoggerOptions{
		InferLevels: true,
	}
	return hclog.Default().StandardLogger(hopts)
}
