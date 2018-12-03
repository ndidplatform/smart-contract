/**
 * Copyright (c) 2018, 2019 National Digital ID COMPANY LIMITED
 *
 * This file is part of NDID software.
 *
 * NDID is the free software: you can redistribute it and/or modify it under
 * the terms of the Affero GNU General Public License as published by the
 * Free Software Foundation, either version 3 of the License, or any later
 * version.
 *
 * NDID is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
 * See the Affero GNU General Public License for more details.
 *
 * You should have received a copy of the Affero GNU General Public License
 * along with the NDID source code. If not, see https://www.gnu.org/licenses/agpl.txt.
 *
 * Please contact info@ndid.co.th for any further questions
 *
 */

package main

import (
	"fmt"
	"os"
	"path/filepath"
	// "strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	cmd "github.com/tendermint/tendermint/cmd/tendermint/commands"
	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/libs/cli"
	"github.com/tendermint/tendermint/libs/log"
	nm "github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/p2p"
	"github.com/tendermint/tendermint/privval"
	"github.com/tendermint/tendermint/proxy"

	"github.com/tendermint/tendermint/abci/types"

	"github.com/ndidplatform/smart-contract/abci/did"
	"github.com/ndidplatform/smart-contract/abci/version"
)

type loggerWriter struct{}

// var mainLogger *logrus.Entry

var abciVersionCmd = &cobra.Command{
	Use:   "abci_app_version",
	Short: "Show DID ABCI app version info",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version.Version)
	},
}

func init() {
	// Set default logrus

	var logLevel = getEnv("LOG_LEVEL", "debug")
	var logTarget = getEnv("LOG_TARGET", "console")

	if logTarget != "console" {
		logFile, _ := os.OpenFile(logTarget, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
		logrus.SetOutput(logFile)
	} else {
		logrus.SetOutput(os.Stdout)
	}

	switch logLevel {
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	default:
		logrus.SetLevel(logrus.DebugLevel)
	}

	customFormatter := new(logrus.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	customFormatter.FullTimestamp = true
	logrus.SetFormatter(customFormatter)
	// mainLogger = logrus.WithFields(logrus.Fields{"module": "abci-app"})
}

func main() {
	rootCmd := cmd.RootCmd
	rootCmd.AddCommand(
		cmd.GenValidatorCmd,
		cmd.InitFilesCmd,
		cmd.ProbeUpnpCmd,
		cmd.LiteCmd,
		cmd.ReplayCmd,
		cmd.ReplayConsoleCmd,
		cmd.ResetAllCmd,
		cmd.ResetPrivValidatorCmd,
		cmd.ShowValidatorCmd,
		cmd.TestnetFilesCmd,
		cmd.ShowNodeIDCmd,
		cmd.GenNodeKeyCmd,
		cmd.VersionCmd,
		abciVersionCmd)

	// NOTE:
	// Users wishing to:
	//	* Use an external signer for their validators
	//	* Supply an in-proc abci app
	//	* Supply a genesis doc file from another source
	//	* Provide their own DB implementation
	// can copy this file and use something other than the
	// DefaultNewNode function
	nodeFunc := newDIDNode

	// Create & start node
	rootCmd.AddCommand(cmd.NewRunNodeCmd(nodeFunc))

	cmd := cli.PrepareBaseCmd(rootCmd, "TM", os.ExpandEnv(filepath.Join("$HOME", cfg.DefaultTendermintDir)))
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}

func newDIDNode(config *cfg.Config, logger log.Logger) (*nm.Node, error) {
	var app types.Application
	app = did.NewDIDApplicationInterface()

	// writer := newLoggerWriter()
	// logger := log.NewTMLogger(log.NewSyncWriter(writer))

	// Generate node PrivKey
	nodeKey, err := p2p.LoadOrGenNodeKey(config.NodeKeyFile())
	if err != nil {
		return nil, err
	}
	return nm.NewNode(config,
		privval.LoadOrGenFilePV(config.PrivValidatorFile()),
		nodeKey,
		proxy.NewLocalClientCreator(app),
		nm.DefaultGenesisDocProviderFunc(config),
		nm.DefaultDBProvider,
		nm.DefaultMetricsProvider(config.Instrumentation),
		logger.With("module", "node"),
	)
}

// func newLoggerWriter() *loggerWriter {
// 	return &loggerWriter{}
// }

// func (w *loggerWriter) Write(p []byte) (int, error) {
// 	allMsg := strings.Fields(string(p))
// 	charType := allMsg[0][0]

// 	keyValues := make(map[string]interface{})
// 	newMsg := ""

// 	for index, msg := range allMsg {
// 		if index > 0 {
// 			if strings.Contains(msg, "=") {
// 				kv := strings.Split(msg, "=")
// 				keyValues[kv[0]] = kv[1]
// 			} else {
// 				newMsg += msg + " "
// 			}
// 		}
// 	}

// 	switch string(charType) {
// 	case "D":
// 		mainLogger.WithFields(keyValues).Debug(newMsg)
// 	case "E":
// 		mainLogger.WithFields(keyValues).Error(newMsg)
// 	default:
// 		mainLogger.WithFields(keyValues).Info(newMsg)
// 	}
// 	return 0, nil
// }

func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = defaultValue
	}
	return value
}
