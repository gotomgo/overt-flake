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

Hid Types:
    simple           simple MAC hardware ID provider
    mac              standard MAC hardward ID provider (default)

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
		"0.2.1",
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

func createHardwareIDProvider(hidType string) (flake.HardwareIDProvider, error) {
	var hidProvider flake.HardwareIDProvider

	switch strings.ToLower(hidType) {
	case "mac":
		hidProvider = flake.NewMacHardwareIDProvider()
		break
	case "simple":
		hidProvider = flake.NewSimpleMacHardwareIDProvider()
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

	var showVersion bool
	var ipAddr string
	var waitForTime int64
	var hidType string
	var genType string
	var epoch int64
	var authToken string

	flag.StringVar(&ipAddr, "ip", "0.0.0.0:4444", "the interface/address to listen on")
	flag.Int64Var(&waitForTime, "waitfor", 0, "the time to wait for prior to generating ids")
	flag.StringVar(&hidType, "hidtype", "mac", "the hardware id provider")
	flag.StringVar(&genType, "gentype", "default", "the type of the id generator")
	flag.StringVar(&authToken, "auth", "", "the auth token used to authenticate clients")
	flag.Int64Var(&epoch, "epoch", flake.OvertoneEpochMs, "the epoch used for id generation")
	flag.BoolVar(&showVersion, "version", false, "print ofsrvr version information")
	flag.BoolVar(&showVersion, "v", false, "print ofsrvr version information")

	flag.Usage = showUsage
	flag.Parse()

	if showVersion {
		os.Exit(0)
	}

	// create the hardward id provider
	hidProvider, err := createHardwareIDProvider(hidType)
	if err != nil {
		showError("Error creating HardwareIDProvider: %s", err)
	}

	// generate the hardware id
	hid, err := hidProvider.GetHardwareID(6)
	if err != nil {
		showError("Error generating Hardware ID: %s", err)
	}

	// create an ID generator
	generator, err := createOvertFlakeIDGenerator(genType, epoch, hid, waitForTime)
	if err != nil {
		showError("Error creating Overt-Flake generator: %s", err)
	}

	// create an OvertFlakeServer
	server := ofsserver.NewOvertFlakeServer(generator, ipAddr, authToken)

	// Wait for SIGINT and SIGTERM (HIT CTRL-C)
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for _ = range ch {
			fmt.Fprintln(os.Stderr, "\nExiting ofsrvr...")
			os.Exit(0)
		}
	}()

	err = server.Serve()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Exiting ofsrvr: %s", err)
	}
}
