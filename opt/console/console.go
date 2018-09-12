// Copyright 2016 The aquachain Authors
// This file is part of the aquachain library.
//
// The aquachain library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The aquachain library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the aquachain library. If not, see <http://www.gnu.org/licenses/>.

package console

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"syscall"

	"github.com/mattn/go-colorable"
	"github.com/peterh/liner"
	"github.com/robertkrimen/otto"
	"gitlab.com/aquachain/aquachain/internal/jsre"
	"gitlab.com/aquachain/aquachain/internal/web3ext"
	"gitlab.com/aquachain/aquachain/rpc"
)

var (
	passwordRegexp = regexp.MustCompile(`personal.[nus]`)
	onlyWhitespace = regexp.MustCompile(`^\s*$`)
	exit           = regexp.MustCompile(`^\s*exit\s*;*\s*$`)
	help           = regexp.MustCompile(`^\s*help\s*;*\s*$`)
	sendline       = regexp.MustCompile(`^\s*send\s*;*\s*$`)
)

// HistoryFile is the file within the data directory to store input scrollback.
const HistoryFile = "history"

// DefaultPrompt is the default prompt line prefix to use for user input querying.
const DefaultPrompt = "AQUA> "

const helpText = `
Web links:

	Explorer: https://aquachain.github.io/explorer/
	Wiki: http://github.com/aquanetwork/aquachain/wiki/Basics
	Chat: https://t.me/AquaCrypto

Common AQUA commands::

	New address:              personal.newAccount()
	Import private key:       personal.importRawKey('the private key')
	Start solo mining (cpu):  miner.start()
	Get balance:              aqua.balance(aqua.coinbase)
	Get all balances:         balance()
	Send transaction:         send
	List accounts:            aqua.accounts
	Show Transaction:         aqua.getTransaction('the tx hash')
	Show Block #1000:         aqua.getBlock('1000')
	Show Latest:              aqua.getBlock('latest')

In this javascript console, you can define variables and load script.

	loadScript('filename.js')
	block = aqua.getBlock
	myBlock = block('0x92cd50f36edddd9347ec37ef93206135518acd4115941f6287ea00407f186e15')
	tx = aqua.getTransaction('0x23eabf63f8da796e2e68cd2ae602c1b5a9cb8f9946ad9d87a9561924e3d20db8')
	web3.fromWei(tx.value)

Press TAB to autocomplete commands
`

const logo = `                              _           _
  __ _  __ _ _   _  __ _  ___| |__   __ _(_)_ __
 / _ '|/ _' | | | |/ _' |/ __| '_ \ / _' | | '_ \
| (_| | (_| | |_| | (_| | (__| | | | (_| | | | | |
 \__,_|\__, |\__,_|\__,_|\___|_| |_|\__,_|_|_| |_|
          |_|` + "\nUpdate Often! https://gitlab.com/aquachain/aquachain\n\n"

// Config is the collection of configurations to fine tune the behavior of the
// JavaScript console.
type Config struct {
	DataDir  string       // Data directory to store the console history at
	DocRoot  string       // Filesystem path from where to load JavaScript files from
	Client   *rpc.Client  // RPC client to execute AquaChain requests through
	Prompt   string       // Input prompt prefix string (defaults to DefaultPrompt)
	Prompter UserPrompter // Input prompter to allow interactive user feedback (defaults to TerminalPrompter)
	Printer  io.Writer    // Output writer to serialize any display strings to (defaults to os.Stdout)
	Preload  []string     // Absolute paths to JavaScript files to preload
}

// Console is a JavaScript interpreted runtime environment. It is a fully fleged
// JavaScript console attached to a running node via an external or in-process RPC
// client.
type Console struct {
	client   *rpc.Client  // RPC client to execute AquaChain requests through
	jsre     *jsre.JSRE   // JavaScript runtime environment running the interpreter
	prompt   string       // Input prompt prefix string
	prompter UserPrompter // Input prompter to allow interactive user feedback
	histPath string       // Absolute path to the console scrollback history
	history  []string     // Scroll history maintained by the console
	printer  io.Writer    // Output writer to serialize any display strings to
}

func New(config Config) (*Console, error) {
	// Handle unset config values gracefully
	if config.Prompter == nil {
		config.Prompter = Stdin
	}
	if config.Prompt == "" {
		config.Prompt = DefaultPrompt
	}
	if config.Printer == nil {
		config.Printer = colorable.NewColorableStdout()
	}
	// Initialize the console and return
	console := &Console{
		client:   config.Client,
		jsre:     jsre.New(config.DocRoot, config.Printer),
		prompt:   config.Prompt,
		prompter: config.Prompter,
		printer:  config.Printer,
		histPath: filepath.Join(config.DataDir, HistoryFile),
	}
	if err := os.MkdirAll(config.DataDir, 0700); err != nil {
		return nil, err
	}
	if err := console.init(config.Preload); err != nil {
		return nil, err
	}
	return console, nil
}

// init retrieves the available APIs from the remote RPC provider and initializes
// the console's JavaScript namespaces based on the exposed modules.
func (c *Console) init(preload []string) error {
	// Initialize the JavaScript <-> Go RPC bridge
	bridge := newBridge(c.client, c.prompter, c.printer)
	c.jsre.Set("jeth", struct{}{})

	jethObj, _ := c.jsre.Get("jeth")
	jethObj.Object().Set("send", bridge.Send)
	jethObj.Object().Set("sendAsync", bridge.Send)

	consoleObj, _ := c.jsre.Get("console")
	consoleObj.Object().Set("log", c.consoleOutput)
	consoleObj.Object().Set("error", c.consoleOutput)

	// Load all the internal utility JavaScript libraries
	if err := c.jsre.Compile("bignumber.js", jsre.BigNumber_JS); err != nil {
		return fmt.Errorf("bignumber.js: %v", err)
	}
	if err := c.jsre.Compile("web3.js", jsre.Web3_JS); err != nil {
		return fmt.Errorf("web3.js: %v", err)
	}
	if _, err := c.jsre.Run("var Web3 = require('web3');"); err != nil {
		return fmt.Errorf("web3 require: %v", err)
	}
	if _, err := c.jsre.Run("var web3 = new Web3(jeth);"); err != nil {
		return fmt.Errorf("web3 provider: %v", err)
	}
	// Load the supported APIs into the JavaScript runtime environment
	apis, err := c.client.SupportedModules()
	if err != nil {
		return fmt.Errorf("api modules: %v", err)
	}
	flatten := "var aqua = web3.aqua; var personal = web3.personal; "
	for api := range apis {
		if api == "web3" {
			continue // manually mapped or ignore
		}
		if file, ok := web3ext.Modules[api]; ok {
			// Load our extension for the module.
			if err = c.jsre.Compile(fmt.Sprintf("%s.js", api), file); err != nil {
				return fmt.Errorf("%s.js: %v", api, err)
			}
			flatten += fmt.Sprintf("var %s = web3.%s; ", api, api)
		} else if obj, err := c.jsre.Run("web3." + api); err == nil && obj.IsObject() {
			// Enable web3.js built-in extension if available.
			flatten += fmt.Sprintf("var %s = web3.%s; ", api, api)
		}
	}
	if _, err = c.jsre.Run(flatten); err != nil {
		return fmt.Errorf("namespace flattening: %v", err)
	}
	// Initialize the global name register (disabled for now)
	//c.jsre.Run(`var GlobalRegistrar = aqua.contract(` + registrar.GlobalRegistrarAbi + `);   registrar = GlobalRegistrar.at("` + registrar.GlobalRegistrarAddr + `");`)

	// If the console is in interactive mode, instrument password related methods to query the user
	if c.prompter != nil {
		// Retrieve the account management object to instrument
		personal, err := c.jsre.Get("personal")
		if err != nil {
			return err
		}
		// Override the openWallet, unlockAccount, newAccount and sign methods since
		// these require user interaction. Assign these method in the Console the
		// original web3 callbacks. These will be called by the jeth.* methods after
		// they got the password from the user and send the original web3 request to
		// the backend.
		if obj := personal.Object(); obj != nil { // make sure the personal api is enabled over the interface
			if _, err = c.jsre.Run(`jeth.openWallet = personal.openWallet;`); err != nil {
				return fmt.Errorf("personal.openWallet: %v", err)
			}
			if _, err = c.jsre.Run(`jeth.unlockAccount = personal.unlockAccount;`); err != nil {
				return fmt.Errorf("personal.unlockAccount: %v", err)
			}
			if _, err = c.jsre.Run(`jeth.newAccount = personal.newAccount;`); err != nil {
				return fmt.Errorf("personal.newAccount: %v", err)
			}
			if _, err = c.jsre.Run(`jeth.sign = personal.sign;`); err != nil {
				return fmt.Errorf("personal.sign: %v", err)
			}
			obj.Set("openWallet", bridge.OpenWallet)
			obj.Set("unlockAccount", bridge.UnlockAccount)
			obj.Set("newAccount", bridge.NewAccount)
			obj.Set("sign", bridge.Sign)
		}
	}
	// The admin.sleep and admin.sleepBlocks are offered by the console and not by the RPC layer.
	admin, err := c.jsre.Get("admin")
	if err != nil {
		return err
	}
	if obj := admin.Object(); obj != nil { // make sure the admin api is enabled over the interface
		obj.Set("sleepBlocks", bridge.SleepBlocks)
		obj.Set("sleep", bridge.Sleep)
		obj.Set("clearHistory", c.clearHistory)
	}
	// Preload any JavaScript files before starting the console
	for _, path := range preload {
		if err := c.jsre.Exec(path); err != nil {
			failure := err.Error()
			if ottoErr, ok := err.(*otto.Error); ok {
				failure = ottoErr.String()
			}
			return fmt.Errorf("%s: %v", path, failure)
		}
	}
	// Configure the console's input prompter for scrollback and tab completion
	if c.prompter != nil {
		if content, err := ioutil.ReadFile(c.histPath); err != nil {
			c.prompter.SetHistory(nil)
		} else {
			c.history = strings.Split(string(content), "\n")
			c.prompter.SetHistory(c.history)
		}
		c.prompter.SetWordCompleter(c.AutoCompleteInput)
	}
	return nil
}

func (c *Console) clearHistory() {
	c.history = nil
	c.prompter.ClearHistory()
	if err := os.Remove(c.histPath); err != nil {
		fmt.Fprintln(c.printer, "can't delete history file:", err)
	} else {
		fmt.Fprintln(c.printer, "history file deleted")
	}
}

// consoleOutput is an override for the console.log and console.error methods to
// stream the output into the configured output stream instead of stdout.
func (c *Console) consoleOutput(call otto.FunctionCall) otto.Value {
	output := []string{}
	for _, argument := range call.ArgumentList {
		output = append(output, fmt.Sprintf("%v", argument))
	}
	fmt.Fprintln(c.printer, strings.Join(output, " "))
	return otto.Value{}
}

// AutoCompleteInput is a pre-assembled word completer to be used by the user
// input prompter to provide hints to the user about the methods available.
func (c *Console) AutoCompleteInput(line string, pos int) (string, []string, string) {
	// No completions can be provided for empty inputs
	if len(line) == 0 || pos == 0 {
		return "", c.jsre.CompleteKeywords(""), ""
	}
	// Chunck data to relevant part for autocompletion
	// E.g. in case of nested lines aqua.getBalance(aqua.coinb<tab><tab>
	start := pos - 1
	for ; start > 0; start-- {
		// Skip all methods and namespaces (i.e. including the dot)
		if line[start] == '.' || (line[start] >= 'a' && line[start] <= 'z') || (line[start] >= 'A' && line[start] <= 'Z') {
			continue
		}
		// Handle web3 in a special way (i.e. other numbers aren't auto completed)
		if start >= 3 && line[start-3:start] == "web3" {
			start -= 3
			continue
		}
		// We've hit an unexpected character, autocomplete form here
		start++
		break
	}
	return line[:start], c.jsre.CompleteKeywords(line[start:pos]), line[pos:]
}

// Welcome show summary of current AquaChain instance and some metadata about the
// console's available modules.
func (c *Console) Welcome() {
	// friendly balance
	c.jsre.Run(`
function pending() {
			var totalBal = 0;
			for (var acctNum in aqua.accounts) {
								var acct = aqua.accounts[acctNum];
								var acctBal = aqua.balance(acct, 'pending');
								totalBal += parseFloat(acctBal);
								console.log("  aqua.accounts[" + acctNum + "]: \t" + acct + " \tbalance: " + acctBal + " AQUA");
						}
			console.log("Pending balance: " + totalBal + " AQUA");
			return totalBal;
};
function balance() {
			var totalBal = 0;
			for (var acctNum in aqua.accounts) {
								var acct = aqua.accounts[acctNum];
								var acctBal = aqua.balance(acct, 'latest');
								totalBal += parseFloat(acctBal);
								console.log("  aqua.accounts[" + acctNum + "]: \t" + acct + " \tbalance: " + acctBal + " AQUA");
						}
			console.log("  Total balance: " + totalBal + " AQUA");
			return totalBal;
};
	`)

	// Print some generic AquaChain metadata
	fmt.Fprintf(c.printer, "\nWelcome to the AquaChain JavaScript console!\n")
	fmt.Fprintf(c.printer, logo)

	c.jsre.Run(`
		console.log("instance: " + web3.version.node);
		console.log("coinbase: " + aqua.coinbase);
		console.log("at block: " + aqua.blockNumber + " (" + new Date(1000 * aqua.getBlock(aqua.blockNumber).timestamp) + ")");
		console.log("    algo: " + aqua.getBlock(aqua.blockNumber).version);
		console.log(" datadir: " + admin.datadir);
	`)

	// List all the supported modules for the user to call
	if apis, err := c.client.SupportedModules(); err == nil {
		modules := make([]string, 0, len(apis))
		for api, version := range apis {
			if api == "eth" {
				continue
			}
			modules = append(modules, fmt.Sprintf("%s:%s", api, version))
		}
		sort.Strings(modules)
		fmt.Fprintln(c.printer, " modules:", strings.Join(modules, " "))
	}
	fmt.Fprintln(c.printer)
}

// Evaluate executes code and pretty prints the result to the specified output
// stream.
func (c *Console) Evaluate(statement string) error {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(c.printer, "[native] error: %v\n", r)
		}
	}()
	return c.jsre.Evaluate(statement, c.printer)
}

// Interactive starts an interactive user session, where input is propted from
// the configured user prompter.
func (c *Console) Interactive() {
	var (
		prompt    = c.prompt          // Current prompt line (used for multi-line inputs)
		indents   = 0                 // Current number of input indents (used for multi-line inputs)
		input     = ""                // Current user input
		scheduler = make(chan string) // Channel to send the next prompt on and receive the input
	)
	// Start a goroutine to listen for promt requests and send back inputs
	go func() {
		for {
			// Read the next user input
			line, err := c.prompter.PromptInput(<-scheduler)
			if err != nil {
				// In case of an error, either clear the prompt or fail
				if err == liner.ErrPromptAborted { // ctrl-C
					prompt, indents, input = c.prompt, 0, ""
					scheduler <- ""
					continue
				}
				close(scheduler)
				return
			}
			// User input retrieved, send for interpretation and loop
			scheduler <- line
		}
	}()
	// Monitor Ctrl-C too in case the input is empty and we need to bail
	abort := make(chan os.Signal, 1)
	signal.Notify(abort, syscall.SIGINT, syscall.SIGTERM)

	// Start sending prompts to the user and reading back inputs
	for {
		// Send the next prompt, triggering an input read and process the result
		scheduler <- prompt
		select {
		case <-abort:
			// User forcefully quite the console
			fmt.Fprintln(c.printer, "caught interrupt, exiting")
			return

		case line, ok := <-scheduler:
			// User input was returned by the prompter, handle special cases
			if !ok || (indents <= 0 && exit.MatchString(line)) {
				return
			}
			if onlyWhitespace.MatchString(line) {
				continue
			}
			if !ok || (indents <= 0 && help.MatchString(line)) {
				fmt.Fprintln(c.printer, helpText)
				continue
			}

			// command: 'send'
			if sendline.MatchString(line) {
				err := handleSend(c)
				if err != nil {
					fmt.Fprintln(c.printer, "Error:", err)
					continue
				}
				fmt.Fprintln(c.printer, "TX Sent!")
				continue
			}
			// Append the line to the input and check for multi-line interpretation
			input += line + "\n"

			indents = countIndents(input)
			if indents <= 0 {
				prompt = c.prompt
			} else {
				prompt = strings.Repeat(".", indents*3) + " "
			}
			// If all the needed lines are present, save the command and run
			if indents <= 0 {
				if len(input) > 0 && input[0] != ' ' && !passwordRegexp.MatchString(input) {
					if command := strings.TrimSpace(input); len(c.history) == 0 || command != c.history[len(c.history)-1] {
						c.history = append(c.history, command)
						if c.prompter != nil {
							c.prompter.AppendHistory(command)
						}
					}
				}
				c.Evaluate(input)
				input = ""
			}
		}
	}
}

func handleSend(c *Console) error {

	cont, err := c.prompter.PromptConfirm("You are about to create a transaction from your Aquabase. Right?")
	if err != nil {
		return fmt.Errorf("input error: %v", err)
	}
	if !cont {
		return fmt.Errorf("transaction canceled")
	}
	_, err = c.jsre.Run(`personal.unlockAccount(aqua.coinbase);`)
	if err != nil {
		return fmt.Errorf("error: %v", err)
	}

	amount, err := c.prompter.PromptInput("How much AQUA to send? For example: 0.1: ")
	if err != nil {
		return fmt.Errorf("input error: %v", err)
	}

	fmt.Fprintf(c.printer, "Send %q to whom?", amount)

	destination, err := c.prompter.PromptInput("Where to send? With 0x prefix: ")
	if err != nil {
		return fmt.Errorf("input error: %v", err)
	}

	cont, err = c.prompter.PromptConfirm(fmt.Sprintf("Send %s to %s?", amount, destination))
	if err != nil {
		return fmt.Errorf("input error: %v", err)
	}
	if !cont {
		return fmt.Errorf("transaction canceled")
	}

	fmt.Fprintln(c.printer, "Running:\n"+`aqua.sendTransaction({from: aqua.coinbase, to: '`+destination+`', value: web3.toWei(`+amount+`,'aqua')});`)
	cont, err = c.prompter.PromptConfirm(fmt.Sprintf("REALLY Send %s to %s?", amount, destination))
	if err != nil {
		return fmt.Errorf("input error: %v", err)
	}
	if !cont {
		return fmt.Errorf("transaction canceled")
	}
	if !strings.HasPrefix(destination, "0x") && !strings.HasPrefix(destination, "aqua.accounts[") {
		return fmt.Errorf("does not have 0x prefix")
	}
	_, err = c.jsre.Run(`aqua.sendTransaction({from: aqua.coinbase, to: '` + destination + `', value: web3.toWei(` + amount + `,'aqua')});`)
	if err != nil {
		return fmt.Errorf("error: %v", err)
	}

	return nil
}

// countIndents returns the number of identations for the given input.
// In case of invalid input such as var a = } the result can be negative.
func countIndents(input string) int {
	var (
		indents     = 0
		inString    = false
		strOpenChar = ' '   // keep track of the string open char to allow var str = "I'm ....";
		charEscaped = false // keep track if the previous char was the '\' char, allow var str = "abc\"def";
	)

	for _, c := range input {
		switch c {
		case '\\':
			// indicate next char as escaped when in string and previous char isn't escaping this backslash
			if !charEscaped && inString {
				charEscaped = true
			}
		case '\'', '"':
			if inString && !charEscaped && strOpenChar == c { // end string
				inString = false
			} else if !inString && !charEscaped { // begin string
				inString = true
				strOpenChar = c
			}
			charEscaped = false
		case '{', '(':
			if !inString { // ignore brackets when in string, allow var str = "a{"; without indenting
				indents++
			}
			charEscaped = false
		case '}', ')':
			if !inString {
				indents--
			}
			charEscaped = false
		default:
			charEscaped = false
		}
	}

	return indents
}

// Execute runs the JavaScript file specified as the argument.
func (c *Console) Execute(path string) error {
	return c.jsre.Exec(path)
}

// Stop cleans up the console and terminates the runtime environment.
func (c *Console) Stop(graceful bool) error {
	if err := ioutil.WriteFile(c.histPath, []byte(strings.Join(c.history, "\n")), 0600); err != nil {
		return err
	}
	if err := os.Chmod(c.histPath, 0600); err != nil { // Force 0600, even if it was different previously
		return err
	}
	c.jsre.Stop(graceful)
	return nil
}
