package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	logrus "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type Opt struct {
	LogDir     string `yaml:"log_dir"`
	LogName    string `yaml:"log_name"`
	LogLevel   string `yaml:"log_level"`
	Addr       string `yaml:"addr"`
	Port       int    `yaml:"port"`
	PiHoleHost string `yaml:"pihole_host"`
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

func main() {
	logger = setupLogger()
	app := fiber.New()

	app.Get("/metrics", func(c *fiber.Ctx) {
		promhttp.Handler()
	})

	app.Listen(Options.Port)
}
