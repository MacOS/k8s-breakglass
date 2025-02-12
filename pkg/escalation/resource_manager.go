package escalation

import (
	"context"
	"sync"

	"github.com/pkg/errors"
	"gitlab.devops.telekom.de/schiff/engine/go-breakglass.git/api/v1alpha1"
	telekomv1alpha1 "gitlab.devops.telekom.de/schiff/engine/go-breakglass.git/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

type ResourceManager struct {
	client.Client
	writeMutex *sync.Mutex
}

var scheme = runtime.NewScheme()

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(telekomv1alpha1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

// Get all stored GetClusterGroupAccess
func (c ResourceManager) GetAllBreakglassEscalations(ctx context.Context) ([]telekomv1alpha1.BreakglassEscalation, error) {
	escal := v1alpha1.BreakglassEscalationList{}
	if err := c.List(ctx, &escal); err != nil {
		return nil, errors.Wrap(err, "failed to get BreakglassSessionList")
	}

	return escal.Items, nil
}

func NewResourceManager() (ResourceManager, error) {
	cfg := config.GetConfigOrDie()
	c, err := client.New(cfg, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return ResourceManager{}, errors.Wrap(err, "failed to create new client")
	}

	return ResourceManager{c, new(sync.Mutex)}, nil
}
