package breakglass

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type BreakglassEscalationController struct {
	manager    *EscalationManager
	log        *zap.SugaredLogger
	middleware gin.HandlerFunc
}

func (ec *BreakglassEscalationController) Register(rg *gin.RouterGroup) error {
	rg.GET("/", ec.handleGetEscalations)
	return nil
}

func (ec BreakglassEscalationController) handleGetEscalations(c *gin.Context) {
}

func (BreakglassEscalationController) BasePath() string {
	return "breakglassEscalation/"
}

func NewEscalationController(log *zap.SugaredLogger,
	manager *EscalationManager,
	middleware gin.HandlerFunc,
) *BreakglassEscalationController {
	return &BreakglassEscalationController{
		log:        log,
		manager:    manager,
		middleware: middleware,
	}
}
