package main

import (
	"fmt"
	"os"

	"github.com/ndidplatform/smart-contract/abci/did"
	server "github.com/tendermint/abci/server"
	"github.com/tendermint/abci/types"
	cmn "github.com/tendermint/tmlibs/common"
)

func main() {
	runABCIServer(os.Args)
}

func runABCIServer(args []string) {
	address := args[1]

	var app types.Application
	app = did.NewDIDApplication()

	// Start the listener
	srv, err := server.NewServer(address, "socket", app)
	if err != nil {
		fmt.Println(err.Error())
	}
	if err := srv.Start(); err != nil {
		fmt.Println(err.Error())
	}

	// Wait forever
	cmn.TrapSignal(func() {
		srv.Stop()
	})
}
