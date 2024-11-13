package accessreview

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.devops.telekom.de/schiff/engine/go-breakglass.git/pkg/config"
	"gitlab.devops.telekom.de/schiff/engine/go-breakglass.git/pkg/webhook/access_review/api/v1alpha1"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/types"
)

type ClusterAccessReviewController struct {
	log        *zap.SugaredLogger
	config     config.Config
	manager    *CRDManager
	middleware gin.HandlerFunc
}

func (ClusterAccessReviewController) BasePath() string {
	return "breakglass/cluster_access/"
}

func (wc *ClusterAccessReviewController) Register(rg *gin.RouterGroup) error {
	rg.GET("/reviews", wc.handleGetReviews)
	rg.POST("/accept/:name", wc.handleAccept)
	rg.POST("/reject/:name", wc.handleReject)
	return nil
}

func (b ClusterAccessReviewController) Handlers() []gin.HandlerFunc {
	return []gin.HandlerFunc{b.middleware}
}

type ClusterAccessReviewResponse struct {
	v1alpha1.ClusterAccessReviewSpec
	Name string    `json:"name,omitempty"`
	UID  types.UID `json:"uid,omitempty"`
}

func (wc ClusterAccessReviewController) handleGetReviews(c *gin.Context) {
	reviews, err := wc.manager.GetReviews()
	if err != nil {
		log.Printf("Error getting access reviews %v", err)
		c.JSON(http.StatusInternalServerError, "Failed to extract review information")
		return
	}

	outReviews := []ClusterAccessReviewResponse{}
	for _, review := range reviews {
		resp := ClusterAccessReviewResponse{
			ClusterAccessReviewSpec: review.Spec,
			Name:                    review.Name,
			UID:                     review.UID,
		}
		outReviews = append(outReviews, resp)
	}

	c.JSON(http.StatusOK, outReviews)
}

func (wc ClusterAccessReviewController) handleAccept(c *gin.Context) {
	wc.handleStatusChange(c, v1alpha1.StatusAccepted)
}

func (wc ClusterAccessReviewController) handleReject(c *gin.Context) {
	wc.handleStatusChange(c, v1alpha1.StatusRejected)
}

func (wc ClusterAccessReviewController) handleStatusChange(c *gin.Context, newStatus v1alpha1.AccessReviewApplicationStatus) {
	name := c.Param("name")
	err := wc.manager.UpdateReviewStatusByName(name, newStatus)
	// err := wc.manager.UpdateReviewStatusByUID(types.UID(name), newStatus)
	if err != nil {
		log.Printf("Error getting access review with id %q %v", name, err)
		c.JSON(http.StatusInternalServerError, "Failed to extract review information")
		return
	}

	c.Status(http.StatusOK)
}

func NewClusterAccessReviewController(log *zap.SugaredLogger,
	cfg config.Config,
	manager *CRDManager,
	middleware gin.HandlerFunc,
) *ClusterAccessReviewController {
	controller := &ClusterAccessReviewController{
		log:        log,
		config:     cfg,
		manager:    manager,
		middleware: middleware,
	}

	return controller
}
