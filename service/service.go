package service

import (
	"context"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/filter"
	"github.com/ONSdigital/dp-api-clients-go/v2/health"
	"github.com/ONSdigital/dp-api-clients-go/v2/hierarchy"
	"github.com/ONSdigital/dp-api-clients-go/v2/search"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/assets"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/config"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/routes"
	render "github.com/ONSdigital/dp-renderer/v2"
	"github.com/ONSdigital/dp-renderer/v2/middleware/renderror"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"github.com/pkg/errors"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
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
	log.Info(ctx, "running service")

	// Initialise Service struct
	svc = &Service{
		Config:      cfg,
		ServiceList: serviceList,
	}

	// Get health client for api router
	svc.routerHealthClient = serviceList.GetHealthClient("api-router", cfg.APIRouterURL)

	// Initialise clients
	svc.clients = &routes.Clients{
		Render:    render.NewWithDefaultClient(assets.Asset, assets.AssetNames, cfg.PatternLibraryAssetsPath, cfg.SiteDomain),
		Filter:    filter.NewWithHealthClient(svc.routerHealthClient),
		Dataset:   dataset.NewWithHealthClient(svc.routerHealthClient),
		Hierarchy: hierarchy.NewWithHealthClient(svc.routerHealthClient),
		Search:    search.NewWithHealthClient(svc.routerHealthClient),
		Zebedee:   zebedee.NewWithHealthClient(svc.routerHealthClient),
	}

	// Get healthcheck with checkers
	svc.HealthCheck, err = serviceList.GetHealthCheck(cfg, buildTime, gitCommit, version)
	if err != nil {
		log.Fatal(ctx, "failed to create health check", err)
		return nil, err
	}
	if err := svc.registerCheckers(ctx, cfg); err != nil {
		return nil, errors.Wrap(err, "unable to register checkers")
	}
	svc.clients.HealthcheckHandler = svc.HealthCheck.Handler

	// Initialise router
	r := mux.NewRouter()
	r.Use(otelmux.Middleware(cfg.OTServiceName))
	middleware := []alice.Constructor{
		renderror.Handler(svc.clients.Render),
	}
	newAlice := alice.New(middleware...).Then(r)
	routes.Init(ctx, r, cfg, svc.clients)
	svc.Server = serviceList.GetHTTPServer(cfg.BindAddr, newAlice)

	// Start Healthcheck and HTTP Server
	log.Info(ctx, "service listening...", log.Data{
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
	log.Info(ctx, "commencing graceful shutdown", log.Data{"graceful_shutdown_timeout": timeout})
	ctx, cancel := context.WithTimeout(ctx, timeout)
	hasShutdownError := false

	go func() {
		defer cancel()

		// stop healthcheck, as it depends on everything else
		if svc.ServiceList.HealthCheck {
			log.Info(ctx, "stop health checkers")
			svc.HealthCheck.Stop()
		}

		// stop any incoming requests
		if err := svc.Server.Shutdown(ctx); err != nil {
			log.Error(ctx, "failed to shutdown http server", err)
			hasShutdownError = true
		}
	}()

	// wait for shutdown success (via cancel) or failure (timeout)
	<-ctx.Done()

	// timeout expired
	if ctx.Err() == context.DeadlineExceeded {
		log.Error(ctx, "shutdown timed out", ctx.Err())
		return ctx.Err()
	}

	// other error
	if hasShutdownError {
		err := errors.New("failed to shutdown gracefully")
		log.Error(ctx, "failed to shutdown gracefully ", err)
		return err
	}

	log.Info(ctx, "graceful shutdown was successful")
	return nil
}

func (svc *Service) registerCheckers(ctx context.Context, cfg *config.Config) (err error) {
	hasErrors := false

	if err = svc.HealthCheck.AddCheck("API router", svc.routerHealthClient.Checker); err != nil {
		hasErrors = true
		log.Error(ctx, "failed to add API router checker", err)
	}

	if hasErrors {
		return errors.New("Error(s) registering checkers for healthcheck")
	}
	return nil
}
