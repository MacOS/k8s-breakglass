package webhook

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gitlab.devops.telekom.de/schiff/engine/go-breakglass.git/pkg/config"
	accessreview "gitlab.devops.telekom.de/schiff/engine/go-breakglass.git/pkg/webhook/access_review"
	"go.uber.org/zap"
	"k8s.io/kubernetes/pkg/apis/authorization"
)

type SubjectAccessReviewResponseStatus struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason"`
}

type SubjectAccessReviewResponse struct {
	ApiVersion string                            `json:"apiVersion"`
	Kind       string                            `json:"kind"`
	Status     SubjectAccessReviewResponseStatus `json:"status"`
}

type WebhookController struct {
	log     *zap.SugaredLogger
	config  config.Config
	manager *accessreview.InMemManager
}

func (WebhookController) BasePath() string {
	return "breakglass/webhook"
}

func (wc *WebhookController) Register(rg *gin.RouterGroup) error {
	rg.POST("/authorize/:cluster_name", wc.handleAuthorize)
	return nil
}

func (b WebhookController) Handlers() []gin.HandlerFunc {
	return []gin.HandlerFunc{}
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

	fmt.Println("----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------")
	fmt.Printf(`User %q (uid=%q) would like to access cluster %q groups are %q
    Requested operation %q %q version %q for namespace %q and group %q. Extra info: %#v
    NonResource: %#v
    `,
		sar.Spec.User,
		sar.Spec.UID,
		cluster,
		sar.Spec.Groups,
		sar.Spec.ResourceAttributes.Verb,
		sar.Spec.ResourceAttributes.Resource,
		sar.Spec.ResourceAttributes.Version,
		sar.Spec.ResourceAttributes.Namespace,
		sar.Spec.ResourceAttributes.Group,
		sar.Spec.Extra,
		sar.Spec.NonResourceAttributes,
	)
	fmt.Println("\n--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------\n")

	ar := accessreview.NewAccessReview(sar.Spec, time.Minute*5)
	allowed := false
	reason := ""
	reviews := wc.manager.GetSubjectReviews(sar.Spec)

	fmt.Println("Subject reviews:=", reviews)
	fmt.Println("Subject reviews:=", len(reviews))

	if len(reviews) == 0 {
		reason = "Access added to be reviewed by administrator."
		wc.manager.AddAccessReview(ar)
	} else {
		for _, review := range reviews {
			if !ar.IsValid() {
				// here we should probably remove the review
				// or it might make sense to create some cleanup function loop, but over
				// here should be sufficient enough
				fmt.Println("Should remove review")
				wc.manager.AddAccessReview(ar)
				continue
			}
			switch review.Status {
			case accessreview.StatusAccepted:
				allowed = true
			case accessreview.StatusPending:
				allowed = false
				reason = "Access pending to be reviewed by administrator."
			case accessreview.StatusRejected:
				reason = "Access already once rejected. New request will be created."
				// TODO: Here we should probably edit the existing one to change status from rejected to pending
			}
		}
	}

	response := SubjectAccessReviewResponse{
		ApiVersion: sar.APIVersion,
		Kind:       sar.Kind,
		Status: SubjectAccessReviewResponseStatus{
			Allowed: allowed,
			Reason:  reason,
		},
	}

	c.JSON(http.StatusOK, &response)
}

func NewWebhookController(log *zap.SugaredLogger, cfg config.Config, manager *accessreview.InMemManager) *WebhookController {
	controller := &WebhookController{
		log:     log,
		config:  cfg,
		manager: manager,
	}

	return controller
}
