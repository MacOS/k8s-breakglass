package accessreview

import (
	"errors"

	"github.com/gin-gonic/gin"
)

type IdentityProvider interface {
	GetIdentity(*gin.Context) (string, error)
}

type KeycloakIdentityProvider struct{}

func (kip KeycloakIdentityProvider) GetIdentity(c *gin.Context) (email string, err error) {
	email = c.GetString("email")

	if email == "" {
		err = errors.New("keycloak provider failed to retrieve email identity")
	}
	return
}
