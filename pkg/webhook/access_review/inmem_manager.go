package accessreview

import (
	"reflect"

	"k8s.io/kubernetes/pkg/apis/authorization"
)

type InMemManager struct {
	reviews []AccessReview
}

func NewInMemManger() InMemManager {
	return InMemManager{reviews: []AccessReview{}}
}

func (c *InMemManager) AddAccessReview(ar AccessReview) {
	c.reviews = append(c.reviews, ar)
}

func (c InMemManager) GetUsersReviews(user string) []AccessReview {
	return []AccessReview{}
}

// TODO Add option of passing some parameters
func (c InMemManager) GetAccessReviews() []AccessReview {
	return c.reviews
}

func (c InMemManager) GetSubjectReviews(
	s authorization.SubjectAccessReviewSpec,
) (outReviews []AccessReview) {
	for _, review := range c.reviews {
		if reflect.DeepEqual(review.Subject, s) {
			outReviews = append(outReviews, review)
		}
	}
	return outReviews
}

func (c InMemManager) ShouldAllow(sars authorization.SubjectAccessReviewSpec) bool {
	for _, review := range c.reviews {
		// for now it is deep equal probably will move it as some lib method
		if reflect.DeepEqual(review.Subject, sars) && review.Status == StatusAccepted {
			return true
		}
	}
	return false
}
