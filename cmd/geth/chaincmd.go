package geth

import (
	"awesomeProject/cmd/utils"
	"awesomeProject/core"
	"encoding/json"
	"github.com/urfave/cli/v2"
	"os"
)

var (
	initCommand = &cli.Command{
		Action:    initGenesis,
		Name:      "init",
		Usage:     "Bootstrap and initialize a new genesis block",
		ArgsUsage: "<genesisPath>",
		Flags:     flags.Merge([]cli.Flag{utils.CachePreimagesFlag}, utils.DatabasePathFlags),
		Description: `
The init command initializes a new genesis block and definition for the network.
This is a destructive action and changes the network in which you will be
participating.

It expects the genesis file as argument.`,
	}
	importCommand = &cli.Command{
		Action:    importChain,
		Name:      "import",
		Usage:     "Import a blockchain file",
		ArgsUsage: "<filename> (<filename 2> ... <filename N>) ",
		Flags: flags.Merge([]cli.Flag{
			utils.CacheFlag,
			utils.SyncModeFlag,
			utils.GCModeFlag,
			utils.SnapshotFlag,
			utils.CacheDatabaseFlag,
			utils.CacheGCFlag,
			utils.MetricsEnabledFlag,
			utils.MetricsEnabledExpensiveFlag,
			utils.MetricsHTTPFlag,
			utils.MetricsPortFlag,
			utils.MetricsEnableInfluxDBFlag,
			utils.MetricsEnableInfluxDBV2Flag,
			utils.MetricsInfluxDBEndpointFlag,
			utils.MetricsInfluxDBDatabaseFlag,
			utils.MetricsInfluxDBUsernameFlag,
			utils.MetricsInfluxDBPasswordFlag,
			utils.MetricsInfluxDBTagsFlag,
			utils.MetricsInfluxDBTokenFlag,
			utils.MetricsInfluxDBBucketFlag,
			utils.MetricsInfluxDBOrganizationFlag,
			utils.TxLookupLimitFlag,
		}, utils.DatabasePathFlags),
		Description: `
The import command imports blocks from an RLP-encoded form. The form can be one file
with several RLP-encoded blocks, or several files can be used.

If only one file is used, import error will result in failure. If several files are used,
processing will proceed even if an individual RLP-file import failure occurs.`,
	}
	exportCommand = &cli.Command{
		Action:    exportChain,
		Name:      "export",
		Usage:     "Export blockchain into file",
		ArgsUsage: "<filename> [<blockNumFirst> <blockNumLast>]",
		Flags: flags.Merge([]cli.Flag{
			utils.CacheFlag,
			utils.SyncModeFlag,
		}, utils.DatabasePathFlags),
		Description: `
Requires a first argument of the file to write to.
Optional second and third arguments control the first and
last block to write. In this mode, the file will be appended
if already existing. If the file ends with .gz, the output will
be gzipped.`,
	}
	importPreimagesCommand = &cli.Command{
		Action:    importPreimages,
		Name:      "import-preimages",
		Usage:     "Import the preimage database from an RLP stream",
		ArgsUsage: "<datafile>",
		Flags: flags.Merge([]cli.Flag{
			utils.CacheFlag,
			utils.SyncModeFlag,
		}, utils.DatabasePathFlags),
		Description: `
The import-preimages command imports hash preimages from an RLP encoded stream.
It's deprecated, please use "geth db import" instead.
`,
	}
	exportPreimagesCommand = &cli.Command{
		Action:    exportPreimages,
		Name:      "export-preimages",
		Usage:     "Export the preimage database into an RLP stream",
		ArgsUsage: "<dumpfile>",
		Flags: flags.Merge([]cli.Flag{
			utils.CacheFlag,
			utils.SyncModeFlag,
		}, utils.DatabasePathFlags),
		Description: `
The export-preimages command exports hash preimages to an RLP encoded stream.
It's deprecated, please use "geth db export" instead.
`,
	}
	dumpCommand = &cli.Command{
		Action:    dump,
		Name:      "dump",
		Usage:     "Dump a specific block from storage",
		ArgsUsage: "[? <blockHash> | <blockNum>]",
		Flags: flags.Merge([]cli.Flag{
			utils.CacheFlag,
			utils.IterativeOutputFlag,
			utils.ExcludeCodeFlag,
			utils.ExcludeStorageFlag,
			utils.IncludeIncompletesFlag,
			utils.StartKeyFlag,
			utils.DumpLimitFlag,
		}, utils.DatabasePathFlags),
		Description: `
This command dumps out the state for a given block (or latest, if none provided).
`,
	}
	dumpGenesisCommand = &cli.Command{
		Action:    dumpGenesis,
		Name:      "dumpgenesis",
		Usage:     "Dumps genesis block JSON configuration to stdout",
		ArgsUsage: "",
		Flags:     append([]cli.Flag{utils.DataDirFlag}, utils.NetworkFlags...),
		Description: `
The dumpgenesis command prints the genesis configuration of the network preset
if one is set.  Otherwise it prints the genesis from the datadir.`,
	}
)

func importChain(context *cli.Context) error {

}

func exportChain(context *cli.Context) error {

}

func importPreimages(context *cli.Context) error {

}

func exportPreimages(context *cli.Context) error {

}

func dump(context *cli.Context) error {

}

func dumpGenesis(context *cli.Context) error {

}

func initGenesis(context *cli.Context) error {
	if context.Args().Len() != 1 {
		utils.Fatalf("need generate.json file as the only argument")
	}
	genesisPath := context.Args().First()
	if len(genesisPath) == 0 {
		utils.Fatalf("invalid genesis path")
	}
	file, err := os.Open(genesisPath)
	if err != nil {
		utils.Fatalf("open genesis file error")
	}
	defer file.Close()
	genesis := new(core.Genesis)
	if err = json.NewDecoder(file).Decode(genesis); err != nil {
		utils.Fatalf("decode genesis file error")
	}
	stake, _ := makeConfigNode(context)
	
}
