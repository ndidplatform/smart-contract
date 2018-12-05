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
	// "os"
	// "strconv"
	// "time"

	"github.com/spf13/cobra"
	// "github.com/spf13/viper"

	// "github.com/tendermint/tendermint/cmd/tendermint/commands"
	// cfg "github.com/tendermint/tendermint/config"
	// tmflags "github.com/tendermint/tendermint/libs/cli/flags"
	// "github.com/tendermint/tendermint/libs/log"
	// "github.com/tendermint/tmlibs/cli"

	"github.com/ndidplatform/smart-contract/abci/version"
)

var abciVersionCmd = &cobra.Command{
	Use:   "abci_app_version",
	Short: "Show DID ABCI app version info",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version.Version)
	},
}

// func customTMRootCmdPersistentPreRunE(cmd *cobra.Command, args []string) (err error) {
// 	if cmd.Name() == commands.VersionCmd.Name() {
// 		return nil
// 	}
// 	config, err := commands.ParseConfig()
// 	if err != nil {
// 		return err
// 	}

// 	var logTarget = getEnv("TENDERMINT_LOG_TARGET", logTargetConsole)
// 	var logger log.Logger

// 	if logTarget == logTargetConsole {
// 		if config.LogFormat == cfg.LogFormatJSON {
// 			logger = log.NewTMJSONLogger(log.NewSyncWriter(os.Stdout))
// 		} else {
// 			logger = log.NewTMLogger(log.NewSyncWriter(os.Stdout))
// 		}
// 	} else if logTarget == logTargetFile {
// 		currentTime := time.Now()
// 		currentTimeStr := currentTime.Format(fileDatetimeFormat)

// 		var logFilePath = getEnv("TENDERMINT_LOG_FILE_PATH", "./tm-"+strconv.Itoa(os.Getpid())+"-"+currentTimeStr+".log")
// 		logFile, _ := os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)

// 		if config.LogFormat == cfg.LogFormatJSON {
// 			logger = log.NewTMJSONLogger(log.NewSyncWriter(logFile))
// 		} else {
// 			logger = log.NewTMLogger(log.NewSyncWriter(logFile))
// 		}
// 	} else {
// 		panic(fmt.Errorf("Unknown log target: \"%s\". Only \"console\" and \"file\" are allowed", logTarget))
// 	}

// 	logger, err = tmflags.ParseLogLevel(config.LogLevel, logger, cfg.DefaultLogLevel())
// 	if err != nil {
// 		return err
// 	}
// 	if viper.GetBool(cli.TraceFlag) {
// 		logger = log.NewTracingLogger(logger)
// 	}
// 	logger = logger.With("module", "main")
// 	return nil
// }
