package service

import (
	"context"

	"github.com/ONSdigital/dp-api-clients-go/dataset"
	"github.com/ONSdigital/dp-api-clients-go/filter"
	"github.com/ONSdigital/dp-api-clients-go/health"
	"github.com/ONSdigital/dp-api-clients-go/hierarchy"
	"github.com/ONSdigital/dp-api-clients-go/renderer"
	"github.com/ONSdigital/dp-api-clients-go/search"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/config"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/routes"
	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

// Service contains all the configs, server and clients to run the frontend filter dataset controller
type Service struct {
	Config             *config.Config
	routerHealthClient *health.Client
	HealthCheck        HealthChecker
	Server             HTTPServer
	clients            *routes.Clients
	ServiceList        *ExternalServiceList
}

// Run the service
func Run(ctx context.Context, cfg *config.Config, serviceList *ExternalServiceList, buildTime, gitCommit, version string, svcErrors chan error) (svc *Service, err error) {
	log.Event(ctx, "running service", log.INFO)

	// Initialise Service struct
	svc = &Service{
		Config:      cfg,
		ServiceList: serviceList,
	}

	// Get health client for api router
	svc.routerHealthClient = serviceList.GetHealthClient("api-router", cfg.APIRouterURL)

	// Initialise clients
	svc.clients = &routes.Clients{
		Renderer:  renderer.New(cfg.RendererURL),
		Filter:    filter.NewWithHealthClient(svc.routerHealthClient),
		Dataset:   dataset.NewWithHealthClient(svc.routerHealthClient),
		Hierarchy: hierarchy.NewWithHealthClient(svc.routerHealthClient),
		Search:    search.NewWithHealthClient(svc.routerHealthClient),
	}

	// Get healthcheck with checkers
	svc.HealthCheck, err = serviceList.GetHealthCheck(cfg, buildTime, gitCommit, version)
	if err != nil {
		log.Event(ctx, "failed to create health check", log.FATAL, log.Error(err))
		return nil, err
	}
	if err := svc.registerCheckers(ctx, cfg); err != nil {
		return nil, errors.Wrap(err, "unable to register checkers")
	}
	svc.clients.HealthcheckHandler = svc.HealthCheck.Handler

	// Initialise router
	r := mux.NewRouter()
	routes.Init(ctx, r, cfg, svc.clients)
	svc.Server = serviceList.GetHTTPServer(cfg.BindAddr, r)

	// Start Healthcheck and HTTP Server
	log.Event(ctx, "service listening...", log.INFO, log.Data{
		"bind_address": cfg.BindAddr,
	})
	svc.HealthCheck.Start(ctx)
	go func() {
		if err := svc.Server.ListenAndServe(); err != nil {
			svcErrors <- errors.Wrap(err, "failure in http listen and serve")
		}
	}()

	return svc, nil
}

// Close gracefully shuts the service down in the required order, with timeout
func (svc *Service) Close(ctx context.Context) error {
	timeout := svc.Config.GracefulShutdownTimeout
	log.Event(ctx, "commencing graceful shutdown", log.INFO, log.Data{"graceful_shutdown_timeout": timeout})
	ctx, cancel := context.WithTimeout(ctx, timeout)
	hasShutdownError := false

	go func() {
		defer cancel()

		// stop healthcheck, as it depends on everything else
		if svc.ServiceList.HealthCheck {
			log.Event(ctx, "stop health checkers", log.INFO)
			svc.HealthCheck.Stop()
		}

		// stop any incoming requests
		if err := svc.Server.Shutdown(ctx); err != nil {
			log.Event(ctx, "failed to shutdown http server", log.Error(err), log.ERROR)
			hasShutdownError = true
		}
	}()

	// wait for shutdown success (via cancel) or failure (timeout)
	<-ctx.Done()

	// timeout expired
	if ctx.Err() == context.DeadlineExceeded {
		log.Event(ctx, "shutdown timed out", log.ERROR, log.Error(ctx.Err()))
		return ctx.Err()
	}

	// other error
	if hasShutdownError {
		err := errors.New("failed to shutdown gracefully")
		log.Event(ctx, "failed to shutdown gracefully ", log.ERROR, log.Error(err))
		return err
	}

	log.Event(ctx, "graceful shutdown was successful", log.INFO)
	return nil
}

func (svc *Service) registerCheckers(ctx context.Context, cfg *config.Config) (err error) {

	hasErrors := false

	if err = svc.HealthCheck.AddCheck("frontend renderer", svc.clients.Renderer.Checker); err != nil {
		hasErrors = true
		log.Event(ctx, "failed to add frontend renderer checker", log.ERROR, log.Error(err))
	}

	if err = svc.HealthCheck.AddCheck("API router", svc.routerHealthClient.Checker); err != nil {
		hasErrors = true
		log.Event(ctx, "failed to add API router checker", log.ERROR, log.Error(err))
	}

	if hasErrors {
		return errors.New("Error(s) registering checkers for healthcheck")
	}
	return nil
}
