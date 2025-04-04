package breakglass

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gitlab.devops.telekom.de/schiff/engine/go-breakglass.git/api/v1alpha1"
	"gitlab.devops.telekom.de/schiff/engine/go-breakglass.git/pkg/config"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

type FakeMailSender struct {
	LastRecivers          []string
	LastSubject, LastBody string
	OnSendError           error
}

func (s *FakeMailSender) Send(receivers []string, subject, body string) error {
	s.LastRecivers = receivers
	s.LastSubject = subject
	s.LastBody = body
	return s.OnSendError
}

var sessionIndexFunctions = map[string]client.IndexerFunc{
	"status.expired": func(o client.Object) []string {
		return []string{strconv.FormatBool(o.(*v1alpha1.BreakglassSession).Status.Expired)}
	},
	"status.approved": func(o client.Object) []string {
		return []string{strconv.FormatBool(o.(*v1alpha1.BreakglassSession).Status.Approved)}
	},
	"status.idleTimeoutReached": func(o client.Object) []string {
		return []string{strconv.FormatBool(o.(*v1alpha1.BreakglassSession).Status.IdleTimeoutReached)}
	},
	"spec.username": func(o client.Object) []string {
		return []string{o.(*v1alpha1.BreakglassSession).Spec.Username}
	},
	"spec.cluster": func(o client.Object) []string {
		return []string{o.(*v1alpha1.BreakglassSession).Spec.Cluster}
	},
	"spec.group": func(o client.Object) []string {
		return []string{o.(*v1alpha1.BreakglassSession).Spec.Group}
	},
}

var escalationIndexFunctions = map[string]client.IndexerFunc{
	"spec.cluster": func(o client.Object) []string {
		return []string{o.(*v1alpha1.BreakglassEscalation).Spec.Cluster}
	},

	"spec.username": func(o client.Object) []string {
		return []string{o.(*v1alpha1.BreakglassEscalation).Spec.Username}
	},
	"spec.escalatedGroup": func(o client.Object) []string {
		return []string{o.(*v1alpha1.BreakglassEscalation).Spec.EscalatedGroup}
	},
}

func TestRequestApproveRejectGetSession(t *testing.T) {
	builder := fake.NewClientBuilder().WithScheme(Scheme)
	for index, fn := range sessionIndexFunctions {
		builder.WithIndex(&v1alpha1.BreakglassSession{}, index, fn)
	}
	for index, fn := range escalationIndexFunctions {
		builder.WithIndex(&v1alpha1.BreakglassEscalation{}, index, fn)
	}
	builder.WithObjects(&v1alpha1.BreakglassEscalation{
		ObjectMeta: metav1.ObjectMeta{
			Name: "tester-allow-create-all",
		},
		Spec: v1alpha1.BreakglassEscalationSpec{
			Cluster:        "test",
			Username:       "tester@telekom.de",
			AllowedGroups:  []string{"system:authenticated"},
			EscalatedGroup: "breakglass-create-all",
			Approvers: v1alpha1.BreakglassEscalationApprovers{
				Users: []string{"approver@telekom.de"},
			},
		},
	})
	cli := builder.WithStatusSubresource(&v1alpha1.BreakglassSession{}).Build()
	sesmanager := SessionManager{
		Client: cli,
	}
	escmanager := EscalationManager{
		Client: cli,
	}

	logger, _ := zap.NewDevelopment()
	ctrl := NewBreakglassSessionController(logger.Sugar(), config.Config{},
		&sesmanager, &escmanager,
		func(c *gin.Context) {
			if c.Request.Method == http.MethodOptions {
				c.Next()
				return
			}

			if c.Request.Method == http.MethodPost {
				c.Set("email", "tester@telekom.de")
				c.Set("username", "Tester")
			}
			if c.Request.Method == http.MethodGet {
				c.Set("email", "approver@telekom.de")
				c.Set("username", "Approver")
			}

			c.Next()
		})

	ctrl.getUserGroupsFn = func(ctx context.Context, cug ClusterUserGroup) ([]string, error) {
		return []string{"system:authenticated", "breakglass-standard-user"}, nil
	}

	ctrl.mail = &FakeMailSender{}

	engine := gin.New()
	_ = ctrl.Register(engine.Group("", ctrl.Handlers()...))

	// create request
	reqData := BreakglassSessionRequest{
		Clustername: "test",
		Username:    "tester@telekom.de",
		Groupname:   "breakglass-create-all",
	}
	b, _ := json.Marshal(reqData)
	req, _ := http.NewRequest("POST", "/request", bytes.NewReader(b))
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	response := w.Result()
	if response.StatusCode != http.StatusCreated {
		t.Errorf("Expected status created (201) got '%d' instead", response.StatusCode)
	}

	// get created request and check if proper fields are set
	req, _ = http.NewRequest("GET", "/status", nil)
	w = httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	response = w.Result()
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK (200) got '%d' instead", response.StatusCode)
	}

	respSessions := []v1alpha1.BreakglassSession{}
	err := json.NewDecoder(response.Body).Decode(&respSessions)
	if err != nil {
		t.Fatalf("Failed to decode response body %v", err)
	}
	if l := len(respSessions); l != 1 {
		t.Fatalf("Expected one breakglass session to be created go %d instead. (%#v)", l, respSessions)
	}
	ses := respSessions[0]
	if stat := ses.Status.CreatedAt; stat.Day() != time.Now().Day() {
		t.Fatalf("Incorrect session creation date day status %#v", stat)
	}
	if stat := ses.Status.StoreUntil; stat.Day() != time.Now().Add(MonthDuration).Day() {
		t.Fatalf("Incorrect session store until date day status %#v", stat)
	}

	// approve session TODO:
	// check if session status is approved  TODO:
	// reject session TODO:
	// check if session status is back to rejected  TODO:
}
