package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"

	"github.com/gotomgo/overt-flake/flake"
	"github.com/gotomgo/overt-flake/ofsserver"
)

var usage = `Usage: ofsrvr [options]
Options:
    -ip              specify the network interface/address to listen on                 default=0.0.0.0:4444
    -hidtype         specify the type of the hardware ID provider                       default=mac
    -gentype         specify the type of generator used to generate IDs                 default=default
    -epoch           specify the epoch in milliseconds elapsed since Unix Epoch         default=1483228800000
    -waitfor         specify a time at which id generation may start, but not before    default=0
    -auth            specify the sequence of characters that make up the auth token     default=""
    -config          specify a path to a configuration file                             default=""
    -hid             specify a hardware id to use when -hidype == "fixed"               default=""

Notes:
* arguments specified on the command-line override values specified in -config file
* waitfor *must* be specified on the command line

Hid Types:
    simple           simple MAC hardware ID provider
    mac              standard MAC hardware ID provider (default)
    fixed            specifies that a fixed hardware id is used (see -hid)

Generator Types:
    default          the standard overt-flake ID generator

Common Options:
    -help, --help    Show this message
    -v, --version    Show version
`

func showUsage() {
	fmt.Fprintf(os.Stderr, "%s\n", usage)
	os.Exit(0)
}

func showAppVersion() {
	fmt.Fprintf(os.Stderr, `ofsrvr: %s, Runtime: %s, Compiler: %s, Copyright © 2017 Overtone Studios, Inc.`,
		"0.3.1",
		runtime.Compiler,
		runtime.Version())
	fmt.Fprintln(os.Stderr, "")
}

func showError(fmtstr string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, fmtstr, args...)
	fmt.Fprintln(os.Stderr, "")
	os.Exit(-1)
}

func showErrorWithUsage(fmtstr string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, fmtstr, args...)
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintf(os.Stderr, "%s\n", usage)
	os.Exit(-1)
}

func createHardwareIDProvider(hidType string, fixedID []byte) (flake.HardwareIDProvider, error) {
	var hidProvider flake.HardwareIDProvider

	switch strings.ToLower(hidType) {
	case "mac":
		hidProvider = flake.NewMacHardwareIDProvider()
		break
	case "simple":
		hidProvider = flake.NewSimpleMacHardwareIDProvider()
		break
	case "fixed":
		// A fixed Hardware ID is used. Create a fixed provider
		hidProvider = flake.NewFixedHardwareIDProvider(fixedID)
		break
	default:
		showErrorWithUsage("Unsupported type for Hardware ID provider: %s", hidType)
	}

	return hidProvider, nil
}

func createOvertFlakeIDGenerator(genType string, epoch int64, hardwareID flake.HardwareID, waitForTime int64) (flake.Generator, error) {
	var generator flake.Generator

	switch strings.ToLower(genType) {
	case "default":
		generator = flake.NewGenerator(epoch, hardwareID, os.Getpid(), waitForTime)
		break
	default:
		showErrorWithUsage("Unsupported type for Generator: %s", genType)
	}

	return generator, nil
}

func main() {
	showAppVersion()

	// args that can override configuration
	var argIPAddr string
	var argHidType string
	var argGenType string
	var argEpoch int64
	var argAuthToken string
	var argHardwareID string

	// other args
	var waitForTime int64
	var configPath string
	var showVersion bool

	flag.StringVar(&argIPAddr, "ip", "", "the interface/address to listen on")
	flag.Int64Var(&waitForTime, "waitfor", 0, "the time to wait for prior to generating ids")
	flag.StringVar(&argHidType, "hidtype", "", "the hardware id provider")
	flag.StringVar(&argGenType, "gentype", "", "the type of the id generator")
	flag.StringVar(&argAuthToken, "auth", "", "the auth token used to authenticate clients")
	flag.Int64Var(&argEpoch, "epoch", -1, "the epoch used for id generation")
	flag.BoolVar(&showVersion, "version", false, "print ofsrvr version information")
	flag.BoolVar(&showVersion, "v", false, "print ofsrvr version information")
	flag.StringVar(&configPath, "config", "", "the path to a ofs server configuration file")
	flag.StringVar(&argHardwareID, "hid", "", "the fixed hardware id")

	flag.Usage = showUsage
	flag.Parse()

	if showVersion {
		os.Exit(0)
	}

	//	---------------------------------------------------------
	//	Create a server configuration with default values
	//	---------------------------------------------------------

	var config = &serverConfig{
		IPAddr:     "0.0.0.0:4444",
		HidType:    "mac",
		GenType:    "default",
		Epoch:      flake.OvertoneEpochMs,
		AuthToken:  "",
		HardwareID: []byte{},
	}

	//	---------------------------------------------------------
	//	Load a server configuration from a file if appropriate
	//	(replacing the defaults)
	//	---------------------------------------------------------

	var err error
	if len(configPath) > 0 {
		config, err = loadServerConfig(configPath)
		if err != nil {
			showError("Error loading configuration from '%s': %s", configPath, err)
		}
	}

	//	---------------------------------------------------------
	//	Override configuration where appropriate
	//	---------------------------------------------------------

	if len(argHidType) > 0 {
		config.HidType = argHidType
	}

	if len(argGenType) > 0 {
		config.GenType = argGenType
	}

	if argEpoch > -1 {
		config.Epoch = argEpoch
	}

	if len(argIPAddr) > 0 {
		config.IPAddr = argIPAddr
	}

	if len(argAuthToken) > 0 {
		config.AuthToken = argAuthToken
	}

	if len(argHardwareID) > 0 {
		if config.HidType != "fixed" {
			showError("Use of fixed hardware ID (-hid) requires '-hidType fixed'")
		}
		config.HardwareID = []byte(argHardwareID)
	}

	//	---------------------------------------------------------
	//	Create the components needed to run the server
	//
	//	- Hardware ID (via HardwareIDProvider)
	//	- overt-flake ID Generator
	//	- overt-flake ID Server
	//	---------------------------------------------------------

	// create the hardware id provider. Note we always pass config.HardwareID as
	// we don't know if it will be used or not (it is used when hidType = "fixed")
	hidProvider, err := createHardwareIDProvider(config.HidType, config.HardwareID)
	if err != nil {
		showError("Error creating HardwareIDProvider: %s", err)
	}

	// generate the hardware id
	hid, err := hidProvider.GetHardwareID(6)
	if err != nil {
		showError("Error generating Hardware ID: %s", err)
	}

	// create an ID generator
	generator, err := createOvertFlakeIDGenerator(config.GenType, config.Epoch, hid, waitForTime)
	if err != nil {
		showError("Error creating Overt-Flake generator: %s", err)
	}

	// create an OvertFlakeServer
	server, err := ofsserver.NewOvertFlakeServer(generator, config.IPAddr, config.AuthToken)
	if err != nil {
		showError("Error creating Overt-Flake server: %s", err)
	}

	// Wait for SIGINT and SIGTERM (HIT CTRL-C)
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for _ = range ch {
			fmt.Fprintln(os.Stderr, "\nExiting ofsrvr...")
			os.Exit(0)
		}
	}()

	//	---------------------------------------------------------
	//	Echo the server configuration
	//	---------------------------------------------------------

	fmt.Fprintf(os.Stderr, "Starting overt-flake ID server on %s\n", config.IPAddr)
	fmt.Fprintf(os.Stderr, "  with epoch = %d\n", config.Epoch)
	fmt.Fprintf(os.Stderr, "  with hardware id = %v\n", hid)
	fmt.Fprintf(os.Stderr, "  with generator type = %s\n", config.GenType)

	if waitForTime != 0 {
		fmt.Fprintf(os.Stderr, "  with waitForTime = %d\n", waitForTime)
	}

	if len(config.AuthToken) == 0 {
		fmt.Fprintln(os.Stderr, "  with server AUTH ****DISABLED****")
	} else {
		fmt.Fprintf(os.Stderr, "  with server AUTH enabled (%s)\n", config.AuthToken)
	}

	//	---------------------------------------------------------
	//	Run the server
	//	---------------------------------------------------------

	err = server.Serve()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Exiting ofsrvr: %s", err)
	}
}
