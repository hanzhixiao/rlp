package geth

import (
	"awesomeProject/internal/flags"
	"github.com/urfave/cli/v2"
	"sort"
)

var app = flags.NewApp("the go-ethereum command line interface")

func init() {
	app.Action = geth
	app.HideVersion = true
	app.Copyright = "copyright 2013-2018 The go-ethereum Authors"
	app.Commands = []cli.Command{
		initCommand,
		importCommand,
		exportCommand,
		importPreimagesCommand,
		exportPreimagesCommand,
		removedbCommand,
		dumpCommand,
		dumpGenesisCommand,
		// See accountcmd.go:
		accountCommand,
		walletCommand,
		// See consolecmd.go:
		consoleCommand,
		attachCommand,
		javascriptCommand,
		// See misccmd.go:
		makecacheCommand,
		makedagCommand,
		versionCommand,
		versionCheckCommand,
		licenseCommand,
		// See config.go
		dumpConfigCommand,
		// see dbcmd.go
		dbCommand,
		// See cmd/utils/flags_legacy.go
		utils.ShowDeprecated,
		// See snapshot.go
		snapshotCommand,
		// See verkle.go
		verkleCommand,
	}
	sort.Sort(cli.CommandsByName(app.Commands))
	app.Flags = flags.Merge(
		nodeFlags,
		rpcFlags,
		consoleFlags,
		debug.Flags,
		metricsFlags,
	)
}

func geth(context *cli.Context) error {

}
