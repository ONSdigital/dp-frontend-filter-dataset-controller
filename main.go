package main

import (
	"context"
	goerrors "errors"
	"os"
	"os/signal"

	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/config"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/service"
	dpotelgo "github.com/ONSdigital/dp-otel-go"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/pkg/errors"

	_ "net/http/pprof"
)

const serviceName = "dp-frontend-filter-dataset-controller"

// App version informaton retrieved on runtime
var (
	// BuildTime represents the time in which the service was built
	BuildTime string
	// GitCommit represents the commit (SHA-1) hash of the service that is running
	GitCommit string
	// Version represents the version of the service that is running
	Version string
)

func main() {
	log.Namespace = serviceName
	ctx := context.Background()

	if err := run(ctx); err != nil {
		log.Fatal(ctx, "application unexpectedly failed", err)
		os.Exit(1)
	}

	os.Exit(0)
}

func run(ctx context.Context) error {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)

	// Create service initialiser and an error channel for fatal errors
	svcErrors := make(chan error, 1)
	svcList := service.NewServiceList(&service.Init{})

	// Read config
	cfg, err := config.Get()
	if err != nil {
		log.Fatal(ctx, "unable to retrieve service configuration", err)
		return err
	}
	log.Info(ctx, "got service configuration", log.Data{"config": cfg})

	// Set up OpenTelemetry
	otelConfig := dpotelgo.Config{
		OtelServiceName:          cfg.OTServiceName,
		OtelExporterOtlpEndpoint: cfg.OTExporterOTLPEndpoint,
		OtelBatchTimeout:         cfg.OTBatchTimeout,
	}

	otelShutdown, err := dpotelgo.SetupOTelSDK(ctx, otelConfig)

	if err != nil {
		log.Error(ctx, "error setting up OpenTelemetry - hint: ensure OTEL_EXPORTER_OTLP_ENDPOINT is set", err)
	}

	// Handle shutdown properly so nothing leaks.
	defer func() {
		err = goerrors.Join(err, otelShutdown(context.Background()))
	}()

	// Start service
	svc, err := service.Run(ctx, cfg, svcList, BuildTime, GitCommit, Version, svcErrors)
	if err != nil {
		return errors.Wrap(err, "running service failed")
	}

	// Blocks until an os interrupt or a fatal error occurs
	select {
	case err := <-svcErrors:
		log.Error(ctx, "service error received", err)
	case sig := <-signals:
		log.Info(ctx, "os signal received", log.Data{"signal": sig})
	}
	return svc.Close(ctx)
}
