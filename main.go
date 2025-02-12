package main

import (
	"flag"
	stdlog "log"

	"go.uber.org/zap"

	"gitlab.devops.telekom.de/schiff/engine/go-breakglass.git/pkg/api"
	"gitlab.devops.telekom.de/schiff/engine/go-breakglass.git/pkg/config"
	"gitlab.devops.telekom.de/schiff/engine/go-breakglass.git/pkg/session"
	"gitlab.devops.telekom.de/schiff/engine/go-breakglass.git/pkg/system"
	"gitlab.devops.telekom.de/schiff/engine/go-breakglass.git/pkg/webhook"
)

func main() {
	debug := true
	flag.BoolVar(&debug, "debug", false, "enables debug mode")
	flag.Parse()

	log := setupLogger(debug)
	log.With("version", system.Version).Info("Starting breakglass api")

	config, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading config for breakglass controller: %v", err)
	}
	config.Defaults()

	if err := config.Validate(); err != nil {
		log.Fatalf("Error validating config for breakglass controller: %v", err)
	}

	if debug {
		log.Infof("%#v", config)
	}

	auth := api.NewAuth(log, config)
	server := api.NewServer(log.Desugar(), config, debug, auth)

	// TODO: I might want to make manager an interface that is passed to webhook controller
	// so that we can switch between CRDManager and DBManager that uses sqlite

	sessionManager, err := session.NewResourceManager()
	if err != nil {
		log.Fatalf("Error creating access review CRD manager: %v", err)
		return
	}

	err = server.RegisterAll([]api.APIController{
		session.NewBreakglassSessionController(log, config, &sessionManager, auth.Middleware()),
		webhook.NewWebhookController(log, config, &sessionManager),
	})
	if err != nil {
		log.Fatalf("Error registering breakglass controllers: %v", err)
	}

	go session.CleanupRoutine{Log: log, Manager: &sessionManager}.CleanupRoutine()

	server.Listen()
}

func setupLogger(debug bool) *zap.SugaredLogger {
	var zlog *zap.Logger
	var err error
	if debug {
		zlog, err = zap.NewDevelopment()
	} else {
		zlog, err = zap.NewProduction()
	}
	if err != nil {
		stdlog.Fatalf("failed to set up logger: %v", err)
	}
	return zlog.Sugar()
}
