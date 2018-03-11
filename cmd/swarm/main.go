// Copyright 2016 The aquachain Authors
// This file is part of aquachain.
//
// aquachain is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// aquachain is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with aquachain. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"crypto/ecdsa"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"syscall"

	"github.com/aquanetwork/aquachain/accounts"
	"github.com/aquanetwork/aquachain/accounts/keystore"
	"github.com/aquanetwork/aquachain/aquaclient"
	"github.com/aquanetwork/aquachain/cmd/utils"
	"github.com/aquanetwork/aquachain/common"
	"github.com/aquanetwork/aquachain/console"
	"github.com/aquanetwork/aquachain/crypto"
	"github.com/aquanetwork/aquachain/internal/debug"
	"github.com/aquanetwork/aquachain/log"
	"github.com/aquanetwork/aquachain/node"
	"github.com/aquanetwork/aquachain/p2p"
	"github.com/aquanetwork/aquachain/p2p/discover"
	"github.com/aquanetwork/aquachain/params"
	"github.com/aquanetwork/aquachain/swarm"
	bzzapi "github.com/aquanetwork/aquachain/swarm/api"
	swarmmetrics "github.com/aquanetwork/aquachain/swarm/metrics"

	"gopkg.in/urfave/cli.v1"
)

const clientIdentifier = "swarm"

var (
	gitCommit        string // Git SHA1 commit hash of the release (set via linker flags)
	testbetBootNodes = []string{}
)

var (
	ChequebookAddrFlag = cli.StringFlag{
		Name:   "chequebook",
		Usage:  "chequebook contract address",
		EnvVar: SWARM_ENV_CHEQUEBOOK_ADDR,
	}
	SwarmAccountFlag = cli.StringFlag{
		Name:   "bzzaccount",
		Usage:  "Swarm account key file",
		EnvVar: SWARM_ENV_ACCOUNT,
	}
	SwarmListenAddrFlag = cli.StringFlag{
		Name:   "httpaddr",
		Usage:  "Swarm HTTP API listening interface",
		EnvVar: SWARM_ENV_LISTEN_ADDR,
	}
	SwarmPortFlag = cli.StringFlag{
		Name:   "bzzport",
		Usage:  "Swarm local http api port",
		EnvVar: SWARM_ENV_PORT,
	}
	SwarmNetworkIdFlag = cli.IntFlag{
		Name:   "bzznetworkid",
		Usage:  "Network identifier (integer, default 3=swarm testnet)",
		EnvVar: SWARM_ENV_NETWORK_ID,
	}
	SwarmConfigPathFlag = cli.StringFlag{
		Name:  "bzzconfig",
		Usage: "DEPRECATED: please use --config path/to/TOML-file",
	}
	SwarmSwapEnabledFlag = cli.BoolFlag{
		Name:   "swap",
		Usage:  "Swarm SWAP enabled (default false)",
		EnvVar: SWARM_ENV_SWAP_ENABLE,
	}
	SwarmSwapAPIFlag = cli.StringFlag{
		Name:   "swap-api",
		Usage:  "URL of the AquaChain API provider to use to settle SWAP payments",
		EnvVar: SWARM_ENV_SWAP_API,
	}
	SwarmSyncEnabledFlag = cli.BoolTFlag{
		Name:   "sync",
		Usage:  "Swarm Syncing enabled (default true)",
		EnvVar: SWARM_ENV_SYNC_ENABLE,
	}
	EnsAPIFlag = cli.StringSliceFlag{
		Name:   "ens-api",
		Usage:  "ENS API endpoint for a TLD and with contract address, can be repeated, format [tld:][contract-addr@]url",
		EnvVar: SWARM_ENV_ENS_API,
	}
	SwarmApiFlag = cli.StringFlag{
		Name:  "bzzapi",
		Usage: "Swarm HTTP endpoint",
		Value: "http://127.0.0.1:8500",
	}
	SwarmRecursiveUploadFlag = cli.BoolFlag{
		Name:  "recursive",
		Usage: "Upload directories recursively",
	}
	SwarmWantManifestFlag = cli.BoolTFlag{
		Name:  "manifest",
		Usage: "Automatic manifest upload",
	}
	SwarmUploadDefaultPath = cli.StringFlag{
		Name:  "defaultpath",
		Usage: "path to file served for empty url path (none)",
	}
	SwarmUpFromStdinFlag = cli.BoolFlag{
		Name:  "stdin",
		Usage: "reads data to be uploaded from stdin",
	}
	SwarmUploadMimeType = cli.StringFlag{
		Name:  "mime",
		Usage: "force mime type",
	}
	CorsStringFlag = cli.StringFlag{
		Name:   "corsdomain",
		Usage:  "Domain on which to send Access-Control-Allow-Origin header (multiple domains can be supplied separated by a ',')",
		EnvVar: SWARM_ENV_CORS,
	}

	// the following flags are deprecated and should be removed in the future
	DeprecatedAquaAPIFlag = cli.StringFlag{
		Name:  "aquaapi",
		Usage: "DEPRECATED: please use --ens-api and --swap-api",
	}
	DeprecatedEnsAddrFlag = cli.StringFlag{
		Name:  "ens-addr",
		Usage: "DEPRECATED: ENS contract address, please use --ens-api with contract address according to its format",
	}
)

//declare a few constant error messages, useful for later error check comparisons in test
var (
	SWARM_ERR_NO_BZZACCOUNT   = "bzzaccount option is required but not set; check your config file, command line or environment variables"
	SWARM_ERR_SWAP_SET_NO_API = "SWAP is enabled but --swap-api is not set"
)

var defaultNodeConfig = node.DefaultConfig

// This init function sets defaults so cmd/swarm can run alongside aquad.
func init() {
	defaultNodeConfig.Name = clientIdentifier
	defaultNodeConfig.Version = params.VersionWithCommit(gitCommit)
	defaultNodeConfig.P2P.ListenAddr = ":30399"
	defaultNodeConfig.IPCPath = "bzzd.ipc"
	// Set flag defaults for --help display.
	utils.ListenPortFlag.Value = 30399
}

var app = utils.NewApp(gitCommit, "AquaChain Swarm")

// This init function creates the cli.App.
func init() {
	app.Action = bzzd
	app.HideVersion = true // we have a command to print the version
	app.Copyright = "Copyright 2013-2016 The aquachain Authors"
	app.Commands = []cli.Command{
		{
			Action:    version,
			Name:      "version",
			Usage:     "Print version numbers",
			ArgsUsage: " ",
			Description: `
The output of this command is supposed to be machine-readable.
`,
		},
		{
			Action:    upload,
			Name:      "up",
			Usage:     "upload a file or directory to swarm using the HTTP API",
			ArgsUsage: " <file>",
			Description: `
"upload a file or directory to swarm using the HTTP API and prints the root hash",
`,
		},
		{
			Action:    list,
			Name:      "ls",
			Usage:     "list files and directories contained in a manifest",
			ArgsUsage: " <manifest> [<prefix>]",
			Description: `
Lists files and directories contained in a manifest.
`,
		},
		{
			Action:    hash,
			Name:      "hash",
			Usage:     "print the swarm hash of a file or directory",
			ArgsUsage: " <file>",
			Description: `
Prints the swarm hash of file or directory.
`,
		},
		{
			Name:      "manifest",
			Usage:     "update a MANIFEST",
			ArgsUsage: "manifest COMMAND",
			Description: `
Updates a MANIFEST by adding/removing/updating the hash of a path.
`,
			Subcommands: []cli.Command{
				{
					Action:    add,
					Name:      "add",
					Usage:     "add a new path to the manifest",
					ArgsUsage: "<MANIFEST> <path> <hash> [<content-type>]",
					Description: `
Adds a new path to the manifest
`,
				},
				{
					Action:    update,
					Name:      "update",
					Usage:     "update the hash for an already existing path in the manifest",
					ArgsUsage: "<MANIFEST> <path> <newhash> [<newcontent-type>]",
					Description: `
Update the hash for an already existing path in the manifest
`,
				},
				{
					Action:    remove,
					Name:      "remove",
					Usage:     "removes a path from the manifest",
					ArgsUsage: "<MANIFEST> <path>",
					Description: `
Removes a path from the manifest
`,
				},
			},
		},
		{
			Name:      "db",
			Usage:     "manage the local chunk database",
			ArgsUsage: "db COMMAND",
			Description: `
Manage the local chunk database.
`,
			Subcommands: []cli.Command{
				{
					Action:    dbExport,
					Name:      "export",
					Usage:     "export a local chunk database as a tar archive (use - to send to stdout)",
					ArgsUsage: "<chunkdb> <file>",
					Description: `
Export a local chunk database as a tar archive (use - to send to stdout).

    swarm db export ~/.aquachain/swarm/bzz-KEY/chunks chunks.tar

The export may be quite large, consider piping the output through the Unix
pv(1) tool to get a progress bar:

    swarm db export ~/.aquachain/swarm/bzz-KEY/chunks - | pv > chunks.tar
`,
				},
				{
					Action:    dbImport,
					Name:      "import",
					Usage:     "import chunks from a tar archive into a local chunk database (use - to read from stdin)",
					ArgsUsage: "<chunkdb> <file>",
					Description: `
Import chunks from a tar archive into a local chunk database (use - to read from stdin).

    swarm db import ~/.aquachain/swarm/bzz-KEY/chunks chunks.tar

The import may be quite large, consider piping the input through the Unix
pv(1) tool to get a progress bar:

    pv chunks.tar | swarm db import ~/.aquachain/swarm/bzz-KEY/chunks -
`,
				},
				{
					Action:    dbClean,
					Name:      "clean",
					Usage:     "remove corrupt entries from a local chunk database",
					ArgsUsage: "<chunkdb>",
					Description: `
Remove corrupt entries from a local chunk database.
`,
				},
			},
		},
		{
			Action: func(ctx *cli.Context) {
				utils.Fatalf("ERROR: 'swarm cleandb' has been removed, please use 'swarm db clean'.")
			},
			Name:      "cleandb",
			Usage:     "DEPRECATED: use 'swarm db clean'",
			ArgsUsage: " ",
			Description: `
DEPRECATED: use 'swarm db clean'.
`,
		},
		// See config.go
		DumpConfigCommand,
	}
	sort.Sort(cli.CommandsByName(app.Commands))

	app.Flags = []cli.Flag{
		utils.IdentityFlag,
		utils.DataDirFlag,
		utils.BootnodesFlag,
		utils.KeyStoreDirFlag,
		utils.ListenPortFlag,
		utils.NoDiscoverFlag,
		utils.DiscoveryV5Flag,
		utils.NetrestrictFlag,
		utils.NodeKeyFileFlag,
		utils.NodeKeyHexFlag,
		utils.MaxPeersFlag,
		utils.NATFlag,
		utils.IPCDisabledFlag,
		utils.IPCPathFlag,
		utils.PasswordFileFlag,
		// bzzd-specific flags
		CorsStringFlag,
		EnsAPIFlag,
		SwarmTomlConfigPathFlag,
		SwarmConfigPathFlag,
		SwarmSwapEnabledFlag,
		SwarmSwapAPIFlag,
		SwarmSyncEnabledFlag,
		SwarmListenAddrFlag,
		SwarmPortFlag,
		SwarmAccountFlag,
		SwarmNetworkIdFlag,
		ChequebookAddrFlag,
		// upload flags
		SwarmApiFlag,
		SwarmRecursiveUploadFlag,
		SwarmWantManifestFlag,
		SwarmUploadDefaultPath,
		SwarmUpFromStdinFlag,
		SwarmUploadMimeType,
		//deprecated flags
		DeprecatedAquaAPIFlag,
		DeprecatedEnsAddrFlag,
	}
	app.Flags = append(app.Flags, debug.Flags...)
	app.Flags = append(app.Flags, swarmmetrics.Flags...)
	app.Before = func(ctx *cli.Context) error {
		runtime.GOMAXPROCS(runtime.NumCPU())
		if err := debug.Setup(ctx); err != nil {
			return err
		}
		swarmmetrics.Setup(ctx)
		return nil
	}
	app.After = func(ctx *cli.Context) error {
		debug.Exit()
		return nil
	}
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func version(ctx *cli.Context) error {
	fmt.Println(strings.Title(clientIdentifier))
	fmt.Println("Version:", params.Version)
	if gitCommit != "" {
		fmt.Println("Git Commit:", gitCommit)
	}
	fmt.Println("Network Id:", ctx.GlobalInt(utils.NetworkIdFlag.Name))
	fmt.Println("Go Version:", runtime.Version())
	fmt.Println("OS:", runtime.GOOS)
	fmt.Printf("GOPATH=%s\n", os.Getenv("GOPATH"))
	fmt.Printf("GOROOT=%s\n", runtime.GOROOT())
	return nil
}

func bzzd(ctx *cli.Context) error {
	//build a valid bzzapi.Config from all available sources:
	//default config, file config, command line and env vars
	bzzconfig, err := buildConfig(ctx)
	if err != nil {
		utils.Fatalf("unable to configure swarm: %v", err)
	}

	cfg := defaultNodeConfig
	//aquad only supports --datadir via command line
	//in order to be consistent within swarm, if we pass --datadir via environment variable
	//or via config file, we get the same directory for aquad and swarm
	if _, err := os.Stat(bzzconfig.Path); err == nil {
		cfg.DataDir = bzzconfig.Path
	}
	//setup the aquachain node
	utils.SetNodeConfig(ctx, &cfg)
	stack, err := node.New(&cfg)
	if err != nil {
		utils.Fatalf("can't create node: %v", err)
	}
	//a few steps need to be done after the config phase is completed,
	//due to overriding behavior
	initSwarmNode(bzzconfig, stack, ctx)
	//register BZZ as node.Service in the aquachain node
	registerBzzService(bzzconfig, ctx, stack)
	//start the node
	utils.StartNode(stack)

	go func() {
		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc, syscall.SIGTERM)
		defer signal.Stop(sigc)
		<-sigc
		log.Info("Got sigterm, shutting swarm down...")
		stack.Stop()
	}()

	// Add bootnodes as initial peers.
	if bzzconfig.BootNodes != "" {
		bootnodes := strings.Split(bzzconfig.BootNodes, ",")
		injectBootnodes(stack.Server(), bootnodes)
	} else {
		if bzzconfig.NetworkId == 3 {
			injectBootnodes(stack.Server(), testbetBootNodes)
		}
	}

	stack.Wait()
	return nil
}

func registerBzzService(bzzconfig *bzzapi.Config, ctx *cli.Context, stack *node.Node) {

	//define the swarm service boot function
	boot := func(ctx *node.ServiceContext) (node.Service, error) {
		var swapClient *aquaclient.Client
		var err error
		if bzzconfig.SwapApi != "" {
			log.Info("connecting to SWAP API", "url", bzzconfig.SwapApi)
			swapClient, err = aquaclient.Dial(bzzconfig.SwapApi)
			if err != nil {
				return nil, fmt.Errorf("error connecting to SWAP API %s: %s", bzzconfig.SwapApi, err)
			}
		}

		return swarm.NewSwarm(ctx, swapClient, bzzconfig)
	}
	//register within the aquachain node
	if err := stack.Register(boot); err != nil {
		utils.Fatalf("Failed to register the Swarm service: %v", err)
	}
}

func getAccount(bzzaccount string, ctx *cli.Context, stack *node.Node) *ecdsa.PrivateKey {
	//an account is mandatory
	if bzzaccount == "" {
		utils.Fatalf(SWARM_ERR_NO_BZZACCOUNT)
	}
	// Try to load the arg as a hex key file.
	if key, err := crypto.LoadECDSA(bzzaccount); err == nil {
		log.Info("Swarm account key loaded", "address", crypto.PubkeyToAddress(key.PublicKey))
		return key
	}
	// Otherwise try getting it from the keystore.
	am := stack.AccountManager()
	ks := am.Backends(keystore.KeyStoreType)[0].(*keystore.KeyStore)

	return decryptStoreAccount(ks, bzzaccount, utils.MakePasswordList(ctx))
}

func decryptStoreAccount(ks *keystore.KeyStore, account string, passwords []string) *ecdsa.PrivateKey {
	var a accounts.Account
	var err error
	if common.IsHexAddress(account) {
		a, err = ks.Find(accounts.Account{Address: common.HexToAddress(account)})
	} else if ix, ixerr := strconv.Atoi(account); ixerr == nil && ix > 0 {
		if accounts := ks.Accounts(); len(accounts) > ix {
			a = accounts[ix]
		} else {
			err = fmt.Errorf("index %d higher than number of accounts %d", ix, len(accounts))
		}
	} else {
		utils.Fatalf("Can't find swarm account key %s", account)
	}
	if err != nil {
		utils.Fatalf("Can't find swarm account key: %v - Is the provided bzzaccount(%s) from the right datadir/Path?", err, account)
	}
	keyjson, err := ioutil.ReadFile(a.URL.Path)
	if err != nil {
		utils.Fatalf("Can't load swarm account key: %v", err)
	}
	for i := 0; i < 3; i++ {
		password := getPassPhrase(fmt.Sprintf("Unlocking swarm account %s [%d/3]", a.Address.Hex(), i+1), i, passwords)
		key, err := keystore.DecryptKey(keyjson, password)
		if err == nil {
			return key.PrivateKey
		}
	}
	utils.Fatalf("Can't decrypt swarm account key")
	return nil
}

// getPassPhrase retrieves the password associated with bzz account, either by fetching
// from a list of pre-loaded passwords, or by requesting it interactively from user.
func getPassPhrase(prompt string, i int, passwords []string) string {
	// non-interactive
	if len(passwords) > 0 {
		if i < len(passwords) {
			return passwords[i]
		}
		return passwords[len(passwords)-1]
	}

	// fallback to interactive mode
	if prompt != "" {
		fmt.Println(prompt)
	}
	password, err := console.Stdin.PromptPassword("Passphrase: ")
	if err != nil {
		utils.Fatalf("Failed to read passphrase: %v", err)
	}
	return password
}

func injectBootnodes(srv *p2p.Server, nodes []string) {
	for _, url := range nodes {
		n, err := discover.ParseNode(url)
		if err != nil {
			log.Error("Invalid swarm bootnode", "err", err)
			continue
		}
		srv.AddPeer(n)
	}
}
