package accessreview

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gitlab.devops.telekom.de/schiff/engine/go-breakglass.git/pkg/config"
	"gitlab.devops.telekom.de/schiff/engine/go-breakglass.git/pkg/webhook/access_review/api/v1alpha1"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type BreakglassSessionController struct {
	log        *zap.SugaredLogger
	config     config.Config
	manager    *CRDManager
	middleware gin.HandlerFunc
}

func (BreakglassSessionController) BasePath() string {
	return "breakglassSession/"
}

func (wc *BreakglassSessionController) Register(rg *gin.RouterGroup) error {
	rg.GET("/", wc.handleGetBreakglassSessions)
	rg.POST("/request", wc.handleRequestBreakglassSession)

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

func (wc BreakglassSessionController) handleRequestBreakglassSession(c *gin.Context) {
	type BreakglassSessionRequest struct {
		Clustername  string `json:"clustername,omitempty"`
		Username     string `json:"username,omitempty"`
		Clustergroup string `json:"clustergroup,omitempty"`
	}
	request := BreakglassSessionRequest{}
	err := json.NewDecoder(c.Request.Body).Decode(&request)
	if err != nil {
		log.Println("error while decoding body:", err)
		c.Status(http.StatusUnprocessableEntity)
		return
	}
	fmt.Println("REQUESTED:=", request)

	// TODO: Approvers should be taken from config or some resource
	bs := v1alpha1.NewBreakglassSession(
		request.Clustername,
		request.Username,
		request.Clustergroup,
		[]string{"approver1@telekom.de"})

	bs.Name = fmt.Sprintf("%s-%s-%s", request.Clustername, request.Username, request.Clustergroup)
	if err := wc.manager.AddBreakglassSession(c.Request.Context(), bs); err != nil {
		log.Println("error while adding breakglass session", err)
		c.Status(http.StatusInternalServerError)
		return
	}

	bs, err = wc.manager.GetBreakglassSessionByName(c.Request.Context(), bs.Name)
	if err != nil {
		log.Println("error while getting bs session", err)
		c.Status(http.StatusInternalServerError)
		return
	}

	const MonthDuration = time.Hour * 24 * 30
	bs.Status = v1alpha1.BreakglassSessionStatus{
		Expired:            false,
		Approved:           true,
		IdleTimeoutReached: false,
		CreatedAt:          metav1.Now(),
		StoreUntil:         metav1.NewTime(time.Now().Add(MonthDuration)),
	}
	if err := wc.manager.UpdateBreakglassSessionStatus(c.Request.Context(), bs); err != nil {
		log.Println("error while updating breakglass session", err)
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, request)
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

// handleGetGroups
func (wc BreakglassSessionController) handleGetGroups(c *gin.Context) {
	// TODO: Should be stored in CRD or in config yaml
	groupList := []string{}
	c.JSON(http.StatusOK, groupList)
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
