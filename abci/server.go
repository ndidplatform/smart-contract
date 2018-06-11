package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/ndidplatform/smart-contract/abci/did"
	"github.com/sirupsen/logrus"
	server "github.com/tendermint/abci/server"
	"github.com/tendermint/abci/types"
	cmn "github.com/tendermint/tmlibs/common"
	tdmLog "github.com/tendermint/tmlibs/log"
)

type loggerWriter struct{}

var log *logrus.Entry

func init() {
	// Set default logrus
	logFile, _ := os.OpenFile("DID.log", os.O_CREATE|os.O_WRONLY, 0666)
	// TODO: add evironment for write log
	// Set write log to file
	if false {
		logrus.SetOutput(logFile)
	} else {
		logrus.SetOutput(os.Stdout)
	}
	logrus.SetLevel(logrus.DebugLevel)
	customFormatter := new(logrus.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	customFormatter.FullTimestamp = true
	logrus.SetFormatter(customFormatter)
	log = logrus.WithFields(logrus.Fields{"module": "abci-app"})
}

func main() {
	runABCIServer(os.Args)
}

func runABCIServer(args []string) {
	address := args[1]

	var app types.Application
	app = did.NewDIDApplication()

	writer := newLoggerWriter()
	logger := tdmLog.NewTMLogger(tdmLog.NewSyncWriter(writer))

	// Start the listener
	srv, err := server.NewServer(address, "socket", app)
	if err != nil {
		fmt.Println(err.Error())
	}
	srv.SetLogger(logger.With("module", "abci-server"))
	if err := srv.Start(); err != nil {
		fmt.Println(err.Error())
	}

	// Wait forever
	cmn.TrapSignal(func() {
		srv.Stop()
	})
}

func newLoggerWriter() *loggerWriter {
	return &loggerWriter{}
}

func (w *loggerWriter) Write(p []byte) (int, error) {
	allMsg := strings.Fields(string(p))
	charType := allMsg[0][0]

	keyValues := make(map[string]interface{})
	newMsg := ""

	for index, msg := range allMsg {
		if index > 0 {
			if strings.Contains(msg, "=") {
				kv := strings.Split(msg, "=")
				keyValues[kv[0]] = kv[1]
			} else {
				newMsg += msg + " "
			}
		}
	}

	switch string(charType) {
	case "D":
		log.WithFields(keyValues).Debug(newMsg)
	case "E":
		log.WithFields(keyValues).Error(newMsg)
	default:
		log.WithFields(keyValues).Info(newMsg)
	}
	return 0, nil
}
