package main

import (
	"context"
	"time"

	"github.com/kubecano/cano-collector/config"
	"github.com/kubecano/cano-collector/pkg/alert"
	alert_interfaces "github.com/kubecano/cano-collector/pkg/alert/interfaces"
	"github.com/kubecano/cano-collector/pkg/destination"
	destination_interfaces "github.com/kubecano/cano-collector/pkg/destination/interfaces"
	"github.com/kubecano/cano-collector/pkg/health"
	health_interfaces "github.com/kubecano/cano-collector/pkg/health/interfaces"
	"github.com/kubecano/cano-collector/pkg/logger"
	logger_interfaces "github.com/kubecano/cano-collector/pkg/logger/interfaces"
	"github.com/kubecano/cano-collector/pkg/metric"
	metric_interfaces "github.com/kubecano/cano-collector/pkg/metric/interfaces"
	"github.com/kubecano/cano-collector/pkg/router"
	router_interfaces "github.com/kubecano/cano-collector/pkg/router/interfaces"
	"github.com/kubecano/cano-collector/pkg/tracer"
	tracer_interfaces "github.com/kubecano/cano-collector/pkg/tracer/interfaces"
	"github.com/kubecano/cano-collector/pkg/workflow"
	workflow_interfaces "github.com/kubecano/cano-collector/pkg/workflow/interfaces"

	"github.com/getsentry/sentry-go"

	config_team "github.com/kubecano/cano-collector/config/team"
	"github.com/kubecano/cano-collector/pkg/util"
	"github.com/kubecano/cano-collector/pkg/workflow/actions"
)

type AppDependencies struct {
	LoggerFactory          func(level string, env string) logger_interfaces.LoggerInterface
	HealthCheckerFactory   func(cfg config.Config, log logger_interfaces.LoggerInterface) health_interfaces.HealthInterface
	TracerManagerFactory   func(cfg config.Config, log logger_interfaces.LoggerInterface) tracer_interfaces.TracerInterface
	MetricsFactory         func(log logger_interfaces.LoggerInterface) metric_interfaces.MetricsInterface
	DestinationFactory     func(log logger_interfaces.LoggerInterface) destination_interfaces.DestinationFactoryInterface
	DestinationRegistry    func(factory destination_interfaces.DestinationFactoryInterface, log logger_interfaces.LoggerInterface) destination_interfaces.DestinationRegistryInterface
	TeamResolverFactory    func(teams config_team.TeamsConfig, log logger_interfaces.LoggerInterface, m metric_interfaces.MetricsInterface) alert_interfaces.TeamResolverInterface
	AlertDispatcherFactory func(registry destination_interfaces.DestinationRegistryInterface, log logger_interfaces.LoggerInterface, m metric_interfaces.MetricsInterface) alert_interfaces.AlertDispatcherInterface
	AlertHandlerFactory    func(cfg config.Config, log logger_interfaces.LoggerInterface, m metric_interfaces.MetricsInterface, tr alert_interfaces.TeamResolverInterface, ad alert_interfaces.AlertDispatcherInterface, converter alert_interfaces.ConverterInterface, workflowEngine workflow_interfaces.WorkflowEngineInterface) alert_interfaces.AlertHandlerInterface
	RouterManagerFactory   func(cfg config.Config, log logger_interfaces.LoggerInterface, t tracer_interfaces.TracerInterface, m metric_interfaces.MetricsInterface, h health_interfaces.HealthInterface, a alert_interfaces.AlertHandlerInterface) router_interfaces.RouterInterface
	ConverterFactory       func(log logger_interfaces.LoggerInterface, cfg config.Config) alert_interfaces.ConverterInterface
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic("Error loading config: " + err.Error())
	}

	deps := AppDependencies{
		LoggerFactory: func(level, env string) logger_interfaces.LoggerInterface {
			return logger.NewLogger(level, env)
		},
		HealthCheckerFactory: func(cfg config.Config, log logger_interfaces.LoggerInterface) health_interfaces.HealthInterface {
			return health.NewHealthChecker(cfg, log)
		},
		TracerManagerFactory: func(cfg config.Config, log logger_interfaces.LoggerInterface) tracer_interfaces.TracerInterface {
			return tracer.NewTracerManager(cfg, log)
		},
		MetricsFactory: metric.NewMetricsCollector,
		DestinationFactory: func(log logger_interfaces.LoggerInterface) destination_interfaces.DestinationFactoryInterface {
			return destination.NewDestinationFactory(log, util.GetSharedHTTPClient())
		},
		DestinationRegistry: func(factory destination_interfaces.DestinationFactoryInterface, log logger_interfaces.LoggerInterface) destination_interfaces.DestinationRegistryInterface {
			return destination.NewDestinationRegistry(factory, log)
		},
		TeamResolverFactory: func(teams config_team.TeamsConfig, log logger_interfaces.LoggerInterface, m metric_interfaces.MetricsInterface) alert_interfaces.TeamResolverInterface {
			return alert.NewTeamResolver(teams, log, m)
		},
		AlertDispatcherFactory: func(registry destination_interfaces.DestinationRegistryInterface, log logger_interfaces.LoggerInterface, m metric_interfaces.MetricsInterface) alert_interfaces.AlertDispatcherInterface {
			return alert.NewAlertDispatcher(registry, log, m)
		},
		AlertHandlerFactory: func(cfg config.Config, log logger_interfaces.LoggerInterface, m metric_interfaces.MetricsInterface, tr alert_interfaces.TeamResolverInterface, ad alert_interfaces.AlertDispatcherInterface, converter alert_interfaces.ConverterInterface, workflowEngine workflow_interfaces.WorkflowEngineInterface) alert_interfaces.AlertHandlerInterface {
			return alert.NewAlertHandler(log, m, tr, ad, converter, workflowEngine)
		},
		RouterManagerFactory: func(cfg config.Config, log logger_interfaces.LoggerInterface, t tracer_interfaces.TracerInterface, m metric_interfaces.MetricsInterface, h health_interfaces.HealthInterface, a alert_interfaces.AlertHandlerInterface) router_interfaces.RouterInterface {
			return router.NewRouterManager(cfg, log, t, m, h, a)
		},
		ConverterFactory: func(log logger_interfaces.LoggerInterface, cfg config.Config) alert_interfaces.ConverterInterface {
			return alert.NewConverterWithConfig(log, cfg)
		},
	}

	if err := run(cfg, deps); err != nil {
		panic("Error running app: " + err.Error())
	}
}

func run(cfg config.Config, deps AppDependencies) error {
	log := deps.LoggerFactory(cfg.LogLevel, cfg.AppEnv)
	log.Debug("Logger initialized")

	healthChecker := deps.HealthCheckerFactory(cfg, log)
	err := healthChecker.RegisterHealthChecks()
	if err != nil {
		log.Panicf("Failed to register health checks: %v", err)
		return err
	}
	log.Debug("Health checks registered")

	tracerManager := deps.TracerManagerFactory(cfg, log)
	metricsCollector := deps.MetricsFactory(log)

	// Initialize destination components
	destinationFactory := deps.DestinationFactory(log)
	destinationRegistry := deps.DestinationRegistry(destinationFactory, log)

	// Load destinations from config
	if err := destinationRegistry.LoadFromConfig(cfg.Destinations); err != nil {
		log.Fatalf("Failed to load destinations from config: %v", err)
		return err
	}
	log.Debug("Destinations loaded from config")

	// Initialize alert processing components
	teamResolver := deps.TeamResolverFactory(cfg.Teams, log, metricsCollector)
	alertDispatcher := deps.AlertDispatcherFactory(destinationRegistry, log, metricsCollector)
	converter := deps.ConverterFactory(log, cfg)

	// Initialize workflow components
	actionRegistry := actions.NewDefaultActionRegistry(log, metricsCollector)

	// Register workflow actions
	if err := registerWorkflowActions(actionRegistry, log, metricsCollector); err != nil {
		log.Fatalf("Failed to register workflow actions: %v", err)
		return err
	}

	actionExecutor := actions.NewDefaultActionExecutor(actionRegistry, log, metricsCollector)
	workflowEngine := workflow.NewWorkflowEngine(&cfg.Workflows, actionExecutor)

	alertHandler := deps.AlertHandlerFactory(cfg, log, metricsCollector, teamResolver, alertDispatcher, converter, workflowEngine)

	// Validate team destinations configuration
	if err := teamResolver.ValidateTeamDestinations(destinationRegistry); err != nil {
		log.Fatalf("Team destinations validation failed: %v", err)
		return err
	}
	log.Debug("Team destinations validation passed")

	routerManager := deps.RouterManagerFactory(cfg, log, tracerManager, metricsCollector, healthChecker, alertHandler)

	if cfg.SentryEnabled {
		if err := initSentry(cfg.SentryDSN); err != nil {
			log.Fatalf("Sentry initialization failed: %v", err)
		}
		log.Debug("Sentry initialized")
	} else {
		log.Debug("Sentry is disabled")
	}

	defer sentry.Flush(2 * time.Second)

	ctx := context.Background()
	err = tracerManager.InitTracer(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize tracing: %v", err)
	}

	defer func(tracerManager tracer_interfaces.TracerInterface, ctx context.Context) {
		err := tracerManager.ShutdownTracer(ctx)
		if err != nil {
			log.Fatalf("Failed to shutdown tracing: %v", err)
		}
	}(tracerManager, ctx)

	r := routerManager.SetupRouter()
	log.Debug("Router setup complete")
	routerManager.StartServer(r)

	return nil
}

// registerWorkflowActions registers all available workflow actions in the action registry
func registerWorkflowActions(actionRegistry *actions.DefaultActionRegistry, log logger_interfaces.LoggerInterface, metrics metric_interfaces.MetricsInterface) error {
	// Create Kubernetes client for pod logs action
	var kubeClient actions.KubernetesClient
	realClient, err := actions.NewRealKubernetesClient(log)
	if err != nil {
		log.Warnf("Failed to create real Kubernetes client, using placeholder client: %v", err)
		kubeClient = actions.NewPlaceholderKubernetesClient(log)
	} else {
		kubeClient = realClient
	}

	// Register Pod Logs Action Factory
	podLogsFactory := actions.NewPodLogsActionFactory(log, metrics, kubeClient)
	if err := actionRegistry.Register("pod_logs", podLogsFactory); err != nil {
		return err
	}

	// Register Label Filter Action Factory
	labelFilterFactory := actions.NewLabelFilterActionFactory(log, metrics)
	if err := actionRegistry.Register("label_filter", labelFilterFactory); err != nil {
		return err
	}

	// Register Severity Router Action Factory
	severityRouterFactory := actions.NewSeverityRouterActionFactory(log, metrics)
	if err := actionRegistry.Register("severity_router", severityRouterFactory); err != nil {
		return err
	}

	log.Info("Workflow actions registered successfully")
	return nil
}

func initSentry(sentryDSN string) error {
	return sentry.Init(sentry.ClientOptions{
		Dsn:              sentryDSN,
		EnableTracing:    true,
		TracesSampleRate: 1.0,
	})
}
