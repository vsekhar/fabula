package cloudenv

// https://levelup.gitconnected.com/4-tips-for-logging-on-gcp-using-golang-and-logrus-239baf3b1ac2

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"runtime"

	"cloud.google.com/go/errorreporting"
	"cloud.google.com/go/logging"
	"github.com/sirupsen/logrus"
)

type stackDriverHook struct {
	client      *logging.Client
	errorClient *errorreporting.Client
	logger      *logging.Logger
	logName     string
}

var logLevelMappings = map[logrus.Level]logging.Severity{
	logrus.TraceLevel: logging.Default,
	logrus.DebugLevel: logging.Debug,
	logrus.InfoLevel:  logging.Info,
	logrus.WarnLevel:  logging.Warning,
	logrus.ErrorLevel: logging.Error,
	logrus.FatalLevel: logging.Critical,
	logrus.PanicLevel: logging.Critical,
}

func newStackDriverHook(logName string, project string) (*stackDriverHook, error) {
	ctx := context.Background()
	client, err := logging.NewClient(ctx, project)
	if err != nil {
		return nil, err
	}
	errorClient, err := errorreporting.NewClient(ctx, project, errorreporting.Config{
		ServiceName: logName,
		OnError: func(err error) {
			log.Printf("Could not log error: %v", err)
		},
	})
	if err != nil {
		return nil, err
	}
	return &stackDriverHook{
		client:      client,
		errorClient: errorClient,
		logger:      client.Logger(logName),
	}, nil
}
func (sh *stackDriverHook) Close() {
	sh.client.Close()
	sh.errorClient.Close()
}
func (sh *stackDriverHook) Levels() []logrus.Level {
	return logrus.AllLevels
}
func (sh *stackDriverHook) Fire(entry *logrus.Entry) error {
	entry.Data["message"] = entry.Message
	if entry.Context != nil {
		entry.Data["context"] = entry.Context
	}
	e := logging.Entry{
		Timestamp: entry.Time,
		Severity:  logLevelMappings[entry.Level],
		Payload:   entry.Data,
	}
	sh.logger.Log(e)
	if int(e.Severity) >= int(logging.Error) {
		err := getError(entry)
		if err == nil {
			errData, e := json.Marshal(e.Payload)
			if e != nil {
				fmt.Printf("Error %v", e)
			}
			fmt.Print(string(errData))
			err = fmt.Errorf(string(errData))
		}
		fmt.Println(err.Error())
		sh.errorClient.Report(errorreporting.Entry{
			Error: err,
			Stack: sh.getStackTrace(),
		})
	}
	return nil
}
func (sh *stackDriverHook) getStackTrace() []byte {
	stackSlice := make([]byte, 2048)
	length := runtime.Stack(stackSlice, false)
	stack := string(stackSlice[0:length])
	re := regexp.MustCompile("[\r\n].*logrus.*")
	res := re.ReplaceAllString(stack, "")
	return []byte(res)
}

type stackDriverError struct {
	Err         interface{}
	Code        interface{}
	Description interface{}
	Message     interface{}
	Env         interface{}
}

func (e stackDriverError) Error() string {
	return fmt.Sprintf("%v - %v - %v - %v - %v", e.Code, e.Description, e.Message, e.Err, e.Env)
}
func getError(entry *logrus.Entry) error {
	errData := entry.Data["error"]
	env := entry.Data["env"]
	code := entry.Data["ErrCode"]
	desc := entry.Data["ErrDescription"]
	msg := entry.Message
	err := stackDriverError{
		Err:         errData,
		Code:        code,
		Message:     msg,
		Description: desc,
		Env:         env,
	}
	return err
}
func (sh *stackDriverHook) Wait() {}
