package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/coosir/middleman/config"
	"github.com/coosir/middleman/dispatcher"
	"github.com/coosir/middleman/jobqueue/logger"
	logWriter "github.com/coosir/middleman/log"
	repository "github.com/coosir/middleman/repository/factory"
	"github.com/coosir/middleman/service"
	"github.com/coosir/middleman/web"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	out := os.Stderr

	initDefaultConfig()

	args, err := parseCmdArgs(os.Args[1:])
	if err != nil {
		os.Exit(1)
	}

	if args.showVersion {
		_, _ = fmt.Fprintln(out, versionString(" "))
		os.Exit(0)
	}

	for _, k := range config.Keys() {
		config.Set(k, *args.settings[k])
	}

	accessLog := initLogging(syscall.SIGUSR1)
	initProcess()
	dispatcher.Init()
	web.Init()

	startServer(accessLog)
}

type cmdArgs struct {
	showVersion bool
	settings    map[string]*string
}

func parseCmdArgs(args []string) (*cmdArgs, error) {
	out := os.Stderr

	parsed := &cmdArgs{
		settings: make(map[string]*string),
	}

	flags := flag.NewFlagSet(Name, flag.ContinueOnError)
	flags.SetOutput(out)
	flags.Usage = func() {
		_, _ = fmt.Fprint(out, helpText)
		for _, item := range config.Descriptions() {
			fmt.Println("")
			_, _ = fmt.Fprintf(out, item.Describe(2, 80-4))
		}
	}
	flags.BoolVar(&parsed.showVersion, "v", false, "")
	flags.BoolVar(&parsed.showVersion, "version", false, "")

	for _, k := range config.Keys() {
		p := new(string)
		parsed.settings[k] = p
		name := strings.Replace(k, "_", "-", -1)
		flags.StringVar(p, name, config.Get(k), "")
	}

	if err := flags.Parse(args); err != nil {
		return nil, err
	}

	return parsed, nil
}

func initDefaultConfig() {
	config.SetDefault("dispatch_user_agent", versionString("/"))
	config.SetDefault("dispatch_keep_alive", config.Get("keep_alive"))
	if len(os.Getenv("DEBUG")) > 0 {
		config.SetDefault("error_log_level", "debug")
		config.SetDefault("queue_log_level", "debug")
	} else {
		config.SetDefault("error_log_level", "info")
		config.SetDefault("queue_log_level", "info")
	}
}

func initProcess() {
	pid := os.Getpid()
	log.Info().Msgf("PID: %d", pid)

	name := config.Get("pid")
	if name == "" {
		return
	}

	if err := os.MkdirAll(filepath.Dir(name), 0755); err != nil {
		log.Error().Msg(err.Error())
		return
	}

	file, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Panic().Msg(err.Error())
	}
	defer func() {
		err = file.Close()
		log.Error().Msg(err.Error())
	}()

	if _, err := fmt.Fprintf(file, "%d\n", pid); err != nil {
		log.Error().Msg(err.Error())
	}
}

func initLogging(sig syscall.Signal) (accessLog logWriter.Writer) {
	// Access log

	accessLog = logWriter.New(os.Stdout)

	accessLogFile := config.Get("access_log")
	if len(accessLogFile) > 0 {
		var err error
		accessLog, err = logWriter.OpenFile(accessLogFile)
		if err != nil {
			log.Error().Msg(err.Error())
		}
	}

	// Error log

	errorLog := logWriter.New(zerolog.ConsoleWriter{Out: os.Stderr})

	errorLevel := zerolog.InfoLevel
	errorLevel = logWriter.ParseLevel(config.Get("error_log_level"), errorLevel)
	zerolog.SetGlobalLevel(errorLevel)

	errorLogFile := config.Get("error_log")
	if len(errorLogFile) > 0 {
		var err error
		errorLog, err = logWriter.OpenFile(errorLogFile)
		if err != nil {
			log.Error().Msg(err.Error())
		}
	}
	log.Logger = log.Output(errorLog)

	// Queue log
	logger.Init()

	// Reopening log files (for logrotate)

	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, sig)
	go func() {
		for {
			s := <-sigC
			log.Info().Msgf("Received signal %q; reopen log files", s)
			if err := accessLog.Reopen(); err != nil {
				log.Error().Msg(err.Error())
			}
			if err := logger.Writer.Reopen(); err != nil {
				log.Error().Msg(err.Error())
			}
			if err := errorLog.Reopen(); err != nil {
				log.Error().Msg(err.Error())
			}
		}
	}()

	return
}

func startServer(accessLogWriter io.Writer) {
	log.Info().Msg("Starting a job dispatcher...")

	repos := repository.NewRepositories()
	dService := service.NewService(repos)

	app := &web.Application{
		AccessLogWriter:   accessLogWriter,
		Version:           versionString(" "),
		Service:           dService,
		QueueRepository:   repos.Queue,
		RoutingRepository: repos.Routing,
	}
	app.Serve()
}

func versionString(sep string) string {
	var prerelease string
	if Prerelease != "" {
		prerelease = "-" + Prerelease
	}

	var build string
	if Build != "" {
		build = "+" + Build
	}

	return strings.Join([]string{Name, sep, Version, prerelease, build}, "")
}

var (
	helpText = `Usage: middleman [options]

  A lightweight, high-performance, stand-alone job queue system.

Options:

  --version, -v  Show the version string.
  --help, -h     Show the help message.
`
)
