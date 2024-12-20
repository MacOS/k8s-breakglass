package accessreview

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.devops.telekom.de/schiff/engine/go-breakglass.git/pkg/config"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/types"
)

type BreakglassSessionController struct {
	log        *zap.SugaredLogger
	config     config.Config
	manager    *CRDManager
	middleware gin.HandlerFunc
}

func (BreakglassSessionController) BasePath() string {
	return "breakglass/cluster_access/"
}

func (wc *BreakglassSessionController) Register(rg *gin.RouterGroup) error {
	rg.GET("/breakglassSession", wc.handleGetBreakglassSessions)
	rg.GET("/groups", wc.handleGetGroups)
	rg.GET("/clusters", wc.handleListClusters)
	rg.POST("/groups", wc.handleListClusters)

	rg.GET("/reviews", wc.handleGetReviews)
	rg.POST("/accept/:name", wc.handleAccept)
	rg.POST("/reject/:name", wc.handleReject)
	return nil
}

func (b BreakglassSessionController) Handlers() []gin.HandlerFunc {
	return []gin.HandlerFunc{b.middleware}
}

type ClusterAccessReviewResponse struct {
	// v1alpha1.ClusterAccessReviewSpec
	Name string    `json:"name,omitempty"`
	UID  types.UID `json:"uid,omitempty"`
}

func (wc BreakglassSessionController) handleGetBreakglassSessions(c *gin.Context) {
	accesses, err := wc.manager.GetAllBreakglassSessions(c.Request.Context())
	if err != nil {
		log.Printf("Error getting access reviews %v", err)
		c.JSON(http.StatusInternalServerError, "Failed to extract cluster group access information")
		return
	}

	c.JSON(http.StatusOK, accesses)
}

func (wc BreakglassSessionController) handleListClusters(c *gin.Context) {
	sessions, err := wc.manager.GetAllBreakglassSessions(c.Request.Context())
	if err != nil {
		log.Printf("Error getting access reviews %v", err)
		c.JSON(http.StatusInternalServerError, "Failed to extract cluster group access information")
		return
	}

	clusters := make([]string, 0, len(sessions))
	for _, session := range sessions {
		clusters = append(clusters, session.Spec.Cluster)
	}

	c.JSON(http.StatusOK, clusters)
}

func (wc BreakglassSessionController) handleGetReviews(c *gin.Context) {
	// reviews, err := wc.manager.GetReviews(c.Request.Context())
	// if err != nil {
	// 	log.Printf("Error getting access reviews %v", err)
	// 	c.JSON(http.StatusInternalServerError, "Failed to extract review information")
	// 	return
	// }
	//
	// outReviews := []ClusterAccessReviewResponse{}
	// for _, review := range reviews {
	// 	resp := ClusterAccessReviewResponse{
	// 		ClusterAccessReviewSpec: review.Spec,
	// 		Name:                    review.Name,
	// 		UID:                     review.UID,
	// 	}
	// 	outReviews = append(outReviews, resp)
	// }

	// c.JSON(http.StatusOK, outReviews)
}

// handleGetGroups
func (wc BreakglassSessionController) handleGetGroups(c *gin.Context) {
	// TODO: Should be stored in CRD or in config yaml
	groupList := []string{}
	c.JSON(http.StatusOK, groupList)
}

func (wc BreakglassSessionController) handleAccept(c *gin.Context) {
	// wc.handleStatusChange(c, v1alpha1.StatusAccepted)
}

func (wc BreakglassSessionController) handleReject(c *gin.Context) {
	// wc.handleStatusChange(c, v1alpha1.StatusRejected)
}

func (wc BreakglassSessionController) handleStatusChange(c *gin.Context) {
	name := c.Param("name")
	// err := wc.manager.UpdateReviewStatusByName(c.Request.Context(), name, newStatus)
	var err error
	// err := wc.manager.UpdateReviewStatusByUID(types.UID(name), newStatus)
	if err != nil {
		log.Printf("Error getting access review with id %q %v", name, err)
		c.JSON(http.StatusInternalServerError, "Failed to extract review information")
		return
	}

	c.Status(http.StatusOK)
}

func NewBreakglassSessionController(log *zap.SugaredLogger,
	cfg config.Config,
	manager *CRDManager,
	middleware gin.HandlerFunc,
) *BreakglassSessionController {
	controller := &BreakglassSessionController{
		log:        log,
		config:     cfg,
		manager:    manager,
		middleware: middleware,
	}

	return controller
}
