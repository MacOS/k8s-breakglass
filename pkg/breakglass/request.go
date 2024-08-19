package breakglass

import (
	"crypto/rsa"
	"fmt"

	"github.com/Nerzal/gocloak/v13"
	"github.com/golang-jwt/jwt/v4"
	"github.com/pkg/errors"
)

func parseRequestToken(token string, publicKey *rsa.PublicKey) (*BreakglassJWTClaims, error) {
	t, err := jwt.ParseWithClaims(token, &BreakglassJWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodRS256 {
			return nil, fmt.Errorf("wrong signing method")
		}
		return publicKey, nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse request token")
	}

	claim, ok := t.Claims.(*BreakglassJWTClaims)
	// Check if token and claims (expiry, not-before) are valid
	if !ok || !t.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claim, nil
}

func isAllowedToApprove(apprUserID string, apprUserGroups []*gocloak.Group, claim *BreakglassJWTClaims) bool {
	if claim.Subject == apprUserID && !claim.Transition.SelfApproval {
		return false
	}

	// Check if user is allowed approver
	for _, group := range apprUserGroups {
		for _, allowedApprover := range claim.Transition.ApprovalGroups {
			// If approving user is in approval group => allowed approver
			if *group.Name == allowedApprover {
				return true
			}
		}
	}
	return false
}
