package logging

import (
	"fmt"
	"io"
	logging "log"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
)

var (
	log *logrus.Logger
)

func init() {

	if _, err := os.Stat("logs"); os.IsNotExist(err) {
		os.Mkdir("logs", 0777)
	}

	f, err := os.OpenFile(fmt.Sprintf("logs/cowboysindians_%d.log", time.Now().Unix()), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {
		logging.Fatalf("error opening file: %v", err)
	}

	log = logrus.New()

	mw := io.MultiWriter(f)
	log.SetOutput(mw)
}

func Info(format string, v ...interface{}) {
	log.Infof(format, v...)
}

func Debug(format string, v ...interface{}) {
	log.Debugf(format, v...)
}
