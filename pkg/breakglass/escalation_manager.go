package breakglass

import (
	"context"
	"fmt"
	"sync"

	"github.com/pkg/errors"
	"gitlab.devops.telekom.de/schiff/engine/go-breakglass.git/api/v1alpha1"
	telekomv1alpha1 "gitlab.devops.telekom.de/schiff/engine/go-breakglass.git/api/v1alpha1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

type EscalationManager struct {
	client.Client
	writeMutex *sync.Mutex
}

// Get all stored GetClusterGroupAccess
func (em EscalationManager) GetAllBreakglassEscalations(ctx context.Context) ([]telekomv1alpha1.BreakglassEscalation, error) {
	escal := v1alpha1.BreakglassEscalationList{}
	if err := em.List(ctx, &escal); err != nil {
		return nil, errors.Wrap(err, "failed to get BreakglassSessionList")
	}

	return escal.Items, nil
}

// Get all stored BreakglassEscalations
func (c EscalationManager) GetEscalationsWith(ctx context.Context) ([]telekomv1alpha1.BreakglassEscalation, error) {
	ess := v1alpha1.BreakglassEscalationList{}
	if err := c.List(ctx, &ess); err != nil {
		return nil, errors.Wrap(err, "failed to get BreakglassSessionList")
	}

	return ess.Items, nil
}

// GetBreakglassEscalationsWithSelector with custom field selector.
func (em EscalationManager) GetBreakglassEscalationsWithSelector(ctx context.Context,
	fs fields.Selector,
) ([]telekomv1alpha1.BreakglassSession, error) {
	ess := v1alpha1.BreakglassSessionList{}

	if err := em.List(ctx, &ess, &client.ListOptions{FieldSelector: fs}); err != nil {
		return nil, errors.Wrapf(err, "failed to list BreakglassSessionList with selector")
	}

	return ess.Items, nil
}

func NewEscalationManager(contextName string) (em EscalationManager, err error) {
	var cfg *rest.Config

	if contextName != "" {
		cfg, err = config.GetConfigWithContext(contextName)
		if err != nil {
			return em, errors.Wrap(err, fmt.Sprintf("failed to get config with context %q", contextName))
		}
	} else {
		cfg, err = config.GetConfig()
		if err != nil {
			return em, errors.Wrap(err, "failed to get default config")
		}
	}

	c, err := client.New(cfg, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return EscalationManager{}, errors.Wrap(err, "failed to create new client")
	}

	return EscalationManager{c, new(sync.Mutex)}, nil
}
