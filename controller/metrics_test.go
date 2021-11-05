package controller

import (
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
)

var (
	listen   = ":9879"
	host     = os.Getenv("VSPHERE_HOST")
	username = os.Getenv("VSPHERE_USERNAME")
	password = os.Getenv("VSPHERE_PASSWORD")
	logLevel = "info"
)

func initLogger() (*log.Logger, error) {
	logger := log.New()
	logrusLogLevel, err := log.ParseLevel(logLevel)
	if err != nil {
		return logger, err
	}
	logger.SetLevel(logrusLogLevel)
	logger.Formatter = &log.TextFormatter{DisableTimestamp: false, FullTimestamp: true}
	return logger, nil
}

func BenchmarkNewVmwareHostMetrics(b *testing.B) {
	logger, err := initLogger()
	if err != nil {
		panic("Could not set logger for testing")
	}

	for i := 0; i < b.N; i++ {
		NewVmwareHostMetrics(host, username, password, logger)
	}
}

func BenchmarkNewVmwareDsMetrics(b *testing.B) {
	logger, err := initLogger()
	if err != nil {
		panic("Could not set logger for testing")
	}

	for i := 0; i < b.N; i++ {
		NewVmwareDsMetrics(host, username, password, logger)
	}
}

func BenchmarkNewVmwareVmMetrics(b *testing.B) {
	logger, err := initLogger()
	if err != nil {
		panic("Could not set logger for testing")
	}

	for i := 0; i < b.N; i++ {
		NewVmwareVmMetrics(host, username, password, logger)
	}
}

func BenchmarkNewVmwareVmPerfMetrics(b *testing.B) {
	logger, err := initLogger()
	if err != nil {
		panic("Could not set logger for testing")
	}

	for i := 0; i < b.N; i++ {
		NewVmwareVmPerfMetrics(host, username, password, logger)
	}
}
