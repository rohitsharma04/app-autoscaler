package main

import (
	helpers "autoscaler/helpers"
	"autoscaler/metricsforwarder/config"
	"autoscaler/metricsforwarder/server"
	"flag"
	"fmt"
	"os"

	"code.cloudfoundry.org/lager"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
	"github.com/tedsuo/ifrit/sigmon"
)

func main() {

	var path string
	flag.StringVar(&path, "c", "", "config file")
	flag.Parse()
	if path == "" {
		fmt.Fprintln(os.Stderr, "missing config file")
		os.Exit(1)
	}
	configFile, err := os.Open(path)
	if err != nil {
		fmt.Fprintf(os.Stdout, "failed to open config file '%s' : %s\n", path, err.Error())
		os.Exit(1)
	}

	var conf *config.Config
	conf, err = config.LoadConfig(configFile)
	if err != nil {
		fmt.Fprintf(os.Stdout, "failed to read config file '%s' : %s\n", path, err.Error())
		os.Exit(1)
	}
	configFile.Close()

	logger := initLoggerFromConfig(conf.LogLevel)

	httpServer, err := server.NewServer(logger.Session("http_server"), conf)
	if err != nil {
		logger.Error("failed to create http server", err)
		os.Exit(1)
	}

	members := grouper.Members{
		{"http_server", httpServer},
	}

	monitor := ifrit.Invoke(sigmon.New(grouper.NewOrdered(os.Interrupt, members)))

	logger.Info("started")

	err = <-monitor.Wait()
	if err != nil {
		logger.Error("exited-with-failure", err)
		os.Exit(1)
	}
	logger.Info("exited")
}

func initLoggerFromConfig(logLevelConfig string) lager.Logger {
	logLevel, err := getLogLevel(logLevelConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize logger: %s\n", err.Error())
		os.Exit(1)
	}
	logger := lager.NewLogger("metricsforwarder")

	keyPatterns := []string{"[Pp]wd", "[Pp]ass", "[Ss]ecret", "[Tt]oken"}

	redactedSink, err := helpers.NewRedactingWriterWithURLCredSink(os.Stdout, logLevel, keyPatterns, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create redacted sink", err.Error())
	}
	logger.RegisterSink(redactedSink)

	return logger
}

func getLogLevel(level string) (lager.LogLevel, error) {
	switch level {
	case "debug":
		return lager.DEBUG, nil
	case "info":
		return lager.INFO, nil
	case "error":
		return lager.ERROR, nil
	case "fatal":
		return lager.FATAL, nil
	default:
		return -1, fmt.Errorf("Error: unsupported log level:%s", level)
	}
}
