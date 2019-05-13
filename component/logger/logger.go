package logger

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
	"bigbigTravel/component/helper"
)

const (
	suffixFormatDay  = "20060102"
	suffixFormatHour = "2006010215"

	targetTypeStdout = "stdout"
	targetTypeFile   = "file"

	defaultLogid = "-"
)

type LogConfig struct {
	WriterTarget string
	FileSuffix   string
	InfoPath     string
	ErrorPath    string
	DebugPath    string
	WarningPath  string
	SystemPath   string
}

type LogWriter struct {
	Writer        *os.File
	Path          string
	CurrentSuffix string
	mutex         sync.Mutex
}

func (lw *LogWriter) Close() {
	if lw.Writer != nil {
		lw.Writer.Close()
		lw.Writer = nil
		lw.CurrentSuffix = ""
	}
}

var lconfig = LogConfig{
	WriterTarget: targetTypeStdout,
	FileSuffix:   suffixFormatDay,
}

var DefaultWriter io.Writer = os.Stdout

var infoWriter LogWriter
var debugWriter LogWriter
var errorWriter LogWriter
var warningWriter LogWriter
var systemWriter LogWriter

var envMode = "debug"

var timenow = func() string { return time.Now().Format("2006-01-02 15:04:05.000") }

func InitConfig(config *LogConfig, mode string) {
	if config == nil {
		return
	}
	infoWriter.Close()
	debugWriter.Close()
	errorWriter.Close()
	warningWriter.Close()
	systemWriter.Close()
	envMode = mode
	if config.WriterTarget == targetTypeStdout || config.WriterTarget == targetTypeFile {
		lconfig.WriterTarget = config.WriterTarget
	}
	if config.FileSuffix == suffixFormatDay || config.FileSuffix == suffixFormatHour {
		lconfig.FileSuffix = config.FileSuffix
	}
	infoWriter.Path = config.InfoPath
	errorWriter.Path = config.ErrorPath
	debugWriter.Path = config.DebugPath
	warningWriter.Path = config.WarningPath
	systemWriter.Path = config.SystemPath
}

func generateWriter(logWriter *LogWriter) {
	if logWriter.Path == "" {
		return
	}

	shouldLogSuffix := time.Now().Format(lconfig.FileSuffix)

	if logWriter.CurrentSuffix != shouldLogSuffix {
		file, err := os.OpenFile(fmt.Sprintf("%s.%s", logWriter.Path, shouldLogSuffix), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
		if err != nil {
			return
		}
		logWriter.CurrentSuffix = shouldLogSuffix
		if logWriter.Writer != nil {
			logWriter.Writer.Close()
		}
		logWriter.Writer = file
	}
}

func Info(category string, logid string, message ...interface{}) {
	msg := helper.ToSting(message...)
	if "" == strings.Trim(msg, "\"") {
		return
	}
	if lconfig.WriterTarget == targetTypeStdout {
		fmt.Fprintf(DefaultWriter, msg)
		return
	}
	infoWriter.mutex.Lock()
	generateWriter(&infoWriter)
	if infoWriter.Writer != nil {

		fmt.Fprintf(infoWriter.Writer, "%v [%s][%s][%s]\t%s\n",
			timenow(),
			"info",
			logid,
			category,
			msg,
		)
	}
	infoWriter.mutex.Unlock()
}

func Error(category string, logid string, message ...interface{}) {
	msg := helper.ToSting(message...)
	if "" == strings.Trim(msg, "\"") {
		return
	}
	if lconfig.WriterTarget == targetTypeStdout {
		fmt.Fprintf(DefaultWriter, msg)
		return
	}
	errorWriter.mutex.Lock()
	generateWriter(&errorWriter)
	if errorWriter.Writer != nil {
		fmt.Fprintf(errorWriter.Writer, "%v [%s][%s][%s]\t%s\n",
			timenow(),
			"error",
			logid,
			category,
			msg,
		)
	}
	errorWriter.mutex.Unlock()
}

func Warning(category string, logid string, message ...interface{}) {

	msg := helper.ToSting(message...)
	if "" == strings.Trim(msg, "\"") {
		return
	}
	if lconfig.WriterTarget == targetTypeStdout {
		fmt.Fprintf(DefaultWriter, msg)
		return
	}
	warningWriter.mutex.Lock()
	generateWriter(&warningWriter)
	if warningWriter.Writer != nil {
		fmt.Fprintf(warningWriter.Writer, "%v [%s][%s][%s]\t%s\n",
			timenow(),
			"warning",
			logid,
			category,
			msg,
		)
	}
	warningWriter.mutex.Unlock()
}

func Debug(category string, logid string, message ...interface{}) {
	msg := helper.ToSting(message...)
	if envMode == "release" || "" == strings.Trim(msg, "\"") {
		return
	}
	if lconfig.WriterTarget == targetTypeStdout {
		fmt.Fprintf(DefaultWriter, msg)
		return
	}
	debugWriter.mutex.Lock()
	generateWriter(&debugWriter)
	if debugWriter.Writer != nil {
		fmt.Fprintf(debugWriter.Writer, "%v [%s][%s][%s]\t%s\n",
			timenow(),
			"debug",
			logid,
			category,
			msg,
		)
	}
	debugWriter.mutex.Unlock()
}

func System(category string, message ...interface{}) {
	msg := helper.ToSting(message...)
	if "" == strings.Trim(msg, "\"") {
		return
	}
	if lconfig.WriterTarget == targetTypeStdout {
		fmt.Fprintf(DefaultWriter, msg)
		return
	}
	if systemWriter.Path == "" {
		return
	}

	if systemWriter.Writer == nil {
		file, err := os.OpenFile(fmt.Sprintf("%s", systemWriter.Path), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
		if err != nil {
			return
		}
		systemWriter.Writer = file
	}
	fmt.Fprintf(systemWriter.Writer, "%v [%s][%d]\t%s\n",
		timenow(),
		category,
		os.Getpid(),
		msg,
	)
}

func GetSystemOutFile() *os.File {
	return systemWriter.Writer
}

type Fields struct {
	FieldsMap map[string]interface{}
}

func GenFields() *Fields {
	newFields := new(Fields)
	newFields.FieldsMap = make(map[string]interface{})
	return newFields
}

func (f *Fields) AddField(name string, value interface{}) *Fields {
	f.FieldsMap[name] = value
	return f
}

func (f *Fields) FindFieldInt(name string) int {
	if value, ok := f.FieldsMap[name]; ok {
		return value.(int)
	}
	return 0
}

func (f *Fields) Marshal() string {
	result := ""
	for k, v := range f.FieldsMap {
		result += k + ":" + fmt.Sprint(v) + " "
	}
	return result
}

func (f *Fields) Monitor(category string, logId string, message ...interface{}) {
	classId := f.FindFieldInt("classId")
	courseId := f.FindFieldInt("courseId")
	userId := f.FindFieldInt("userId")
	actionId := f.FindFieldInt("actionId")

	appendMsg := helper.ToSting(message...)
	fieldsMsg := f.Marshal()
	msg := fieldsMsg + appendMsg
	if "" == strings.Trim(msg, "\"") {
		return
	}
	if lconfig.WriterTarget == targetTypeStdout {
		fmt.Fprintf(DefaultWriter, msg)
		return
	}
	infoWriter.mutex.Lock()
	generateWriter(&infoWriter)
	if infoWriter.Writer != nil {

		fmt.Fprintf(infoWriter.Writer, "%v [%s][%s][%s][%d][%d][%d][%d]\t%s\n",
			timenow(),
			"monitor",
			logId,
			category,
			classId,
			courseId,
			userId,
			actionId,
			msg,
		)
	}
	infoWriter.mutex.Unlock()
}


func (f *Fields) DataAnalysis(category string, logId string, message ...interface{}) {
	connId := f.FindFieldInt("connId")
	termId := f.FindFieldInt("termId")
	bigClassId := f.FindFieldInt("bigClassId")
	classId := f.FindFieldInt("classId")
	courseId := f.FindFieldInt("courseId")
	userId := f.FindFieldInt("userId")
	actionId := f.FindFieldInt("actionId")

	appendMsg := helper.ToSting(message...)
	fieldsMsg := f.Marshal()
	msg := fieldsMsg + appendMsg
	if "" == strings.Trim(msg, "\"") {
		return
	}
	if lconfig.WriterTarget == targetTypeStdout {
		fmt.Fprintf(DefaultWriter, msg)
		return
	}
	infoWriter.mutex.Lock()
	generateWriter(&infoWriter)
	if infoWriter.Writer != nil {

		fmt.Fprintf(infoWriter.Writer, "%v [%s][%s][%s] connId:%d userId:%d bigClassId:%d miniClassId:%d termId:%d courseId:%d action:%d\t%s\n",
			timenow(),
			"dataAnalysis",
			logId,
			category,
			connId,
			userId,
			bigClassId,
			classId,
			termId,
			courseId,
			actionId,
			msg,
		)
	}
	infoWriter.mutex.Unlock()
}


