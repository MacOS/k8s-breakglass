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
	ID       uint
	Cluster  string
	Subject  authorization.SubjectAccessReviewSpec
	Status   AccessReviewStatus
	Until    time.Time
	Duration time.Duration
}

func NewAccessReview(cluster string, subject authorization.SubjectAccessReviewSpec,
	duration time.Duration,
) AccessReview {
	until := time.Now().Add(duration)
	ar := AccessReview{
		Cluster:  cluster,
		Subject:  subject,
		Status:   StatusPending,
		Until:    until,
		Duration: duration,
	}

	return ar
}

func (ar AccessReview) IsValid() bool {
	timeNow := time.Now()
	return timeNow.Before(ar.Until)
}
