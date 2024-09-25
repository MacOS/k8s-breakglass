package breakglass

import (
	"context"
	"crypto/rsa"
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"gitlab.devops.telekom.de/schiff/engine/go-breakglass.git/pkg/config"
	"gitlab.devops.telekom.de/schiff/engine/go-breakglass.git/pkg/mail"
	"go.uber.org/zap"
)

const transitionsMaxAge = time.Hour

type BreakglassController struct {
	log      *zap.SugaredLogger
	config   config.Config
	keycloak *KeycloakConnector
	mail     mail.Sender

	transitions        []config.Transition
	transitionsFetched time.Time
	transitionsMutex   sync.Mutex

	jwtPrivateKey *rsa.PrivateKey
	jwtPublicKey  *rsa.PublicKey
	middleware    gin.HandlerFunc
}

type Requestor struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type BreakglassJWTClaims struct {
	jwt.StandardClaims
	Transition config.Transition `json:"transition"`
	Requestor  Requestor         `json:"requestor"`
}

func NewBreakglassController(log *zap.SugaredLogger,
	cfg config.Config, middleware gin.HandlerFunc,
) *BreakglassController {
	controller := &BreakglassController{
		log:        log,
		config:     cfg,
		mail:       mail.NewSender(cfg),
		middleware: middleware,
	}

	connector, err := NewKeycloakConnector(
		controller.config.Keycloak.Url,
		controller.config.Keycloak.ClientID,
		controller.config.Keycloak.ClientSecret,
		controller.config.Keycloak.LoginRealm,
		controller.config.Keycloak.ManagedRealm,
	)
	if err != nil {
		log.Fatalf("Error building keycloak connector: %v\n", err)
	}

	controller.keycloak = connector

	// Parse Signing / Verifying keys for breakglass request JWTs
	controller.jwtPrivateKey, err = jwt.ParseRSAPrivateKeyFromPEM([]byte(controller.config.BreakglassJWT.JWTPrivateKey))
	if err != nil {
		log.Fatalf("Error parsing JWT private key: %v\n", err)
	}
	controller.jwtPublicKey, err = jwt.ParseRSAPublicKeyFromPEM([]byte(controller.config.BreakglassJWT.JWTPublicKey))
	if err != nil {
		log.Fatalf("Error parsing JWT public key: %v\n", err)
	}

	// Run cleanup Task / Routine
	go CleanupRoutine{log: log, breakglass: controller}.cleanupRoutine()

	return controller
}

func (*BreakglassController) BasePath() string {
	return "/breakglass"
}

func (b *BreakglassController) getTransitions() ([]config.Transition, error) {
	b.transitionsMutex.Lock()
	defer b.transitionsMutex.Unlock()
	if time.Now().Sub(b.transitionsFetched) < transitionsMaxAge {
		return b.transitions, nil
	}

	b.log.Info("discovering transitions on keycloak")

	discoveredTransitions, err := b.discoverTransitions(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to discover transitions error: %v", err)
	}

	transitions := b.config.PossibleTransitions
	for _, t := range discoveredTransitions {
		if !containsTransition(transitions, t.From, t.To) {
			transitions = append(transitions, t)
		}
	}

	b.log.Infof("%v total transitions available", len(transitions))

	b.transitions = transitions
	b.transitionsFetched = time.Now()
	return transitions, nil
}

func containsTransition(transitions []config.Transition, from, to string) bool {
	for _, t := range transitions {
		if t.From == from && t.To == to {
			return true
		}
	}
	return false
}

func (b *BreakglassController) isBreakglassTargetGroup(group string) bool {
	transitions, err := b.getTransitions()
	if err != nil {
		b.log.Errorf("failed to fetch transitions: %v", err)
		return false
	}
	for _, t := range transitions {
		if t.To == group {
			return true
		}
	}
	return false
}

func (b *BreakglassController) getGlobalBreakglassTransitions() ([]config.Transition, error) {
	allTransitions, err := b.getTransitions()
	b.log.Debug("all transitions", allTransitions)
	if err != nil {
		return nil, err
	}
	globalTransitions := []config.Transition{}
	for _, t := range allTransitions {
		if !t.GlobalBreakglassExcluded {
			globalTransitions = append(globalTransitions, t)
		}
	}
	return globalTransitions, nil
}

func (b *BreakglassController) getUserTransitions(ctx context.Context, userId string) ([]config.Transition, error) {
	groups, err := b.keycloak.GetUserGroups(ctx, userId)
	if err != nil {
		return nil, err
	}

	possibleTransitions, err := b.getTransitions()
	if err != nil {
		return nil, err
	}

	userTransitions := []config.Transition{}
	for _, transition := range possibleTransitions {
		for _, group := range groups {
			if *group.Name == transition.From {
				userTransitions = append(userTransitions, transition)
			}
		}
	}

	globalBreakglass := false
	for _, g := range groups {
		if g.Name != nil && contains(b.config.GlobalBreakglassGroups, *g.Name) {
			globalBreakglass = true
			break
		}
	}

	if globalBreakglass {
		globalTransitions, err := b.getGlobalBreakglassTransitions()
		if err != nil {
			return userTransitions, err
		}
		userTransitions = append(userTransitions, globalTransitions...)
	}

	return userTransitions, nil
}

func (b *BreakglassController) Handlers() []gin.HandlerFunc {
	return []gin.HandlerFunc{b.middleware}
}

func contains(arr []string, s string) bool {
	for _, a := range arr {
		if a == s {
			return true
		}
	}
	return false
}
