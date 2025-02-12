package breakglass

import (
	"context"
	"sync"

	"github.com/pkg/errors"
	"gitlab.devops.telekom.de/schiff/engine/go-breakglass.git/api/v1alpha1"
	telekomv1alpha1 "gitlab.devops.telekom.de/schiff/engine/go-breakglass.git/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

type EscalationManager struct {
	client.Client
	writeMutex *sync.Mutex
}

// var scheme = runtime.NewScheme()
//
// func init() {
// 	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
//
// 	utilruntime.Must(telekomv1alpha1.AddToScheme(scheme))
// 	// +kubebuilder:scaffold:scheme
// }

// Get all stored GetClusterGroupAccess
func (c EscalationManager) GetAllBreakglassEscalations(ctx context.Context) ([]telekomv1alpha1.BreakglassEscalation, error) {
	escal := v1alpha1.BreakglassEscalationList{}
	if err := c.List(ctx, &escal); err != nil {
		return nil, errors.Wrap(err, "failed to get BreakglassSessionList")
	}

	return escal.Items, nil
}

func NewEscalationManager() (EscalationManager, error) {
	cfg := config.GetConfigOrDie()
	c, err := client.New(cfg, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return EscalationManager{}, errors.Wrap(err, "failed to create new client")
	}

	return EscalationManager{c, new(sync.Mutex)}, nil
}
