package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/ndidplatform/ndid/abci/did"
	server "github.com/tendermint/abci/server"
	"github.com/tendermint/abci/types"
	cmn "github.com/tendermint/tmlibs/common"
	tdmLog "github.com/tendermint/tmlibs/log"
)

func main() {
	runABCIServer(os.Args)
}

func runABCIServer(args []string) {
	address := args[1]
	fmt.Println(address)

	logger := tdmLog.NewTMLogger(tdmLog.NewSyncWriter(os.Stdout))

	var app types.Application
	app = did.NewDIDApplication()

	// Start the listener
	srv, err := server.NewServer(address, "socket", app)
	if err != nil {
		color.Red("%s", err)
	}
	srv.SetLogger(logger.With("module", "abci-server"))
	if err := srv.Start(); err != nil {
		color.Red("%s", err)
	}

	// Wait forever
	cmn.TrapSignal(func() {
		srv.Stop()
	})
}
