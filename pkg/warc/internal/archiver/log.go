package archiver

import "github.com/sirupsen/logrus"

type logType int

const (
	infoLog logType = iota
	errorLog
	warningLog
)

// Log prints the log ended with newline.
func (arc *Archiver) Log(tp logType, msgs ...interface{}) {
	if !arc.LogEnabled {
		return
	}

	switch tp {
	case errorLog:
		logrus.Errorln(msgs...)
	case warningLog:
		logrus.Warnln(msgs...)
	default:
		logrus.Infoln(msgs...)
	}
}

// Logf print log with specified format.
func (arc *Archiver) Logf(tp logType, format string, msgs ...interface{}) {
	if !arc.LogEnabled {
		return
	}

	switch tp {
	case errorLog:
		logrus.Errorf(format, msgs...)
	case warningLog:
		logrus.Warnf(format, msgs...)
	default:
		logrus.Infof(format, msgs...)
	}
}
