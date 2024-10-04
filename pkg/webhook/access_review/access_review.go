package accessreview

import (
	"time"

	"k8s.io/kubernetes/pkg/apis/authorization"
)

type AccessReviewStatus string

const (
	StatusPending  AccessReviewStatus = "Pending"
	StatusAccepted AccessReviewStatus = "Accepted"
	StatusRejected AccessReviewStatus = "Rejected"
)

type AccessReview struct {
	Subject  authorization.SubjectAccessReviewSpec
	Status   AccessReviewStatus
	Until    time.Time
	TimeNow  func() time.Time `json:"-"`
	Duration time.Duration
}

func NewAccessReview(subject authorization.SubjectAccessReviewSpec,
	duration time.Duration,
) AccessReview {
	until := time.Now().Add(duration)
	ar := AccessReview{
		Subject:  subject,
		Status:   StatusPending,
		Until:    until,
		TimeNow:  time.Now,
		Duration: duration,
	}

	return ar
}

func (ar AccessReview) IsValid() bool {
	return ar.TimeNow().Before(ar.Until)
}
