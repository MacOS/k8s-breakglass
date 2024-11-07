package main

import (
	"flag"
	"fmt"
	stdlog "log"

	"go.uber.org/zap"

	"gitlab.devops.telekom.de/schiff/engine/go-breakglass.git/pkg/api"
	"gitlab.devops.telekom.de/schiff/engine/go-breakglass.git/pkg/breakglass"
	"gitlab.devops.telekom.de/schiff/engine/go-breakglass.git/pkg/config"
	"gitlab.devops.telekom.de/schiff/engine/go-breakglass.git/pkg/system"
	"gitlab.devops.telekom.de/schiff/engine/go-breakglass.git/pkg/webhook"
	accessreview "gitlab.devops.telekom.de/schiff/engine/go-breakglass.git/pkg/webhook/access_review"
	"gitlab.devops.telekom.de/schiff/engine/go-breakglass.git/pkg/webhook/access_review/api/v1alpha1"
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

	if debug {
		log.Infof("%#v", config)
	}

	auth := api.NewAuth(log, config)
	server := api.NewServer(log.Desugar(), config, debug, auth)

	reviewDB, err := accessreview.NewAccessReviewDB(log, config)
	if err != nil {
		log.Fatalf("Error creating access review database manager: %v", err)
	}

	crdManager, err := accessreview.NewCRDManager()
	if err != nil {
		log.Fatalf("Error creating access review CRD manager: %v", err)
		return
	}

	err = crdManager.AddAccessReview(v1alpha1.ClusterAccessReview{
		Spec: v1alpha1.ClusterAccessReviewSpec{
			Status:  v1alpha1.StatusPending,
			Cluster: "kind2",
			Subject: v1alpha1.ClusterAccessReviewSubject{Username: "tester"},
		},
	})
	if err != nil {
		log.Fatalf("Error creating new access review: %v", err)
		return
	}

	ars, err := crdManager.GetClusterUserReviews("kind", "unknown")
	if err != nil {
		log.Fatalf("Error getting reviews from access review CRD manager: %v", err)
		return
	}

	for _, ar := range ars {
		fmt.Println("Current ar", ar.UID, " SPEC:=", ar.Spec)
	}

	err = server.RegisterAll([]api.APIController{
		breakglass.NewBreakglassController(log, config, auth.Middleware()),
		accessreview.NewClusterAccessReviewController(log, config, &reviewDB, auth.Middleware()),
		webhook.NewWebhookController(log, config, &reviewDB),
	})
	if err != nil {
		log.Fatalf("Error registering breakglass controllers: %v", err)
	}

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
