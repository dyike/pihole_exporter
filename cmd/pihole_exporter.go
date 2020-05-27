package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"pihole_exporter/exporter"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	logrus "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type Opt struct {
	LogDir         string `yaml:"log_dir"`
	LogName        string `yaml:"log_name"`
	LogLevel       string `yaml:"log_level"`
	Port           string `yaml:"port"`
	PiHoleInterval int    `yaml:"pihole_interval"`
	PiHoleHost     string `yaml:"pihole_host"`
	PiHoleToken    string `yaml:"pihole_token"`
}

var (
	Options *Opt
	logger  *logrus.Logger
)

func setupLogger() *logrus.Logger {
	level, err := logrus.ParseLevel(Options.LogLevel)
	if err != nil {
		log.Fatalf("Log level error %v", err)
	}
	logPath := fmt.Sprintf("%s/%s", Options.LogDir, Options.LogName)
	path, err := filepath.Abs(logPath + ".%Y%m%d")
	if err != nil {
		log.Fatalf("Log level error %v", err)
	}
	rl, err := rotatelogs.New(path,
		rotatelogs.WithLinkName(logPath),
		rotatelogs.WithRotationTime(3600*time.Second),
	)
	if err != nil {
		log.Fatalf("Log level error %v", err)
	}
	out := io.MultiWriter(os.Stdout, rl)
	logger := logrus.Logger{
		Formatter: &logrus.TextFormatter{},
		Level:     level,
		Out:       out,
	}
	logger.Info("Setup log finished.")

	return &logger
}

func init() {
	configPath := flag.String("c", "./config.yaml", "Service config yaml")
	flag.Parse()

	buf, err := ioutil.ReadFile(*configPath)
	if err != nil {
		log.Fatal("cannot open config file, err=", err)
	}

	err = yaml.Unmarshal(buf, &Options)
	if err != nil {
		log.Fatal("cannot parse config file, err=", err)
	}
}

var s *exporter.Server

func main() {
	logger = setupLogger()

	exporter.InitMetrics()

	// init client
	client := exporter.NewClient(Options.PiHoleHost, Options.PiHoleToken, Options.PiHoleInterval)
	go client.Collect()

	// init http server
	s = exporter.NewServer("9510")
	go s.ListenAndServe()

	// handle exit signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop

	s.Stop()
}
