package webhook

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.devops.telekom.de/schiff/engine/go-breakglass.git/pkg/config"
	"go.uber.org/zap"
	"k8s.io/kubernetes/pkg/apis/authorization"
)

type SubjectAccessReviewResponseStatus struct {
	Allowed bool `json:"allowed"`
}
type SubjectAccessReviewResponse struct {
	ApiVersion string                            `json:"apiVersion"`
	Kind       string                            `json:"kind"`
	Status     SubjectAccessReviewResponseStatus `json:"status"`
}

type WebhookController struct {
	log    *zap.SugaredLogger
	config config.Config
}

func (WebhookController) BasePath() string {
	return "breakglass/webhook"
}

func (wc *WebhookController) Register(rg *gin.RouterGroup) error {
	rg.POST("/authorize/:cluster_name", wc.handleAuthorize)
	return nil
}

func (wc *WebhookController) handleAuthorize(c *gin.Context) {
	cluster := c.Param("cluster_name")
	sar := authorization.SubjectAccessReview{}
	err := json.NewDecoder(c.Request.Body).Decode(&sar)
	fmt.Println(cluster)
	if err != nil {
		log.Println("error while decoding body:", err)
		c.Status(http.StatusUnprocessableEntity)
		return
	}

	fmt.Println("--------------------------------------------------------------------------------------------")
	fmt.Printf(`User %q with would like to access cluster %q groups are %q
  Requested operation %q %q version %q for namespace %q group %q`,
		sar.Spec.User,
		cluster,
		sar.Spec.Groups,
		sar.Spec.ResourceAttributes.Verb,
		sar.Spec.ResourceAttributes.Resource,
		sar.Spec.ResourceAttributes.Version,
		sar.Spec.ResourceAttributes.Namespace,
		sar.Spec.ResourceAttributes.Group,
	)
	fmt.Println("\n----------------------------------------------------------------------------------------\n")

	response := SubjectAccessReviewResponse{
		ApiVersion: sar.APIVersion,
		Kind:       sar.Kind,
		Status:     SubjectAccessReviewResponseStatus{Allowed: true},
	}

	c.JSON(http.StatusOK, &response)
}

func NewWebhookController(log *zap.SugaredLogger, cfg config.Config) *WebhookController {
	controller := &WebhookController{
		log:    log,
		config: cfg,
	}

	return controller
}
