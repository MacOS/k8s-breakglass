package breakglass

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/Nerzal/gocloak/v13"
)

type KeycloakConnector struct {
	clientID     string
	clientSecret string

	client      *gocloak.GoCloak
	token       *gocloak.JWT
	loginRealm  string
	userRealm   string
	lastRequest time.Time

	prematureRefreshTokenRefreshThreshold int
	prematureAccessTokenRefreshThreshold  int
}

func NewKeycloakConnector(url, clientID, clientSecret, loginRealm string, userRealm string) (*KeycloakConnector, error) {
	connector := &KeycloakConnector{
		clientID:     clientID,
		clientSecret: clientSecret,

		// ctx:        context.Background(),
		client:     gocloak.NewClient(url),
		loginRealm: loginRealm,
		userRealm:  userRealm,

		prematureRefreshTokenRefreshThreshold: 0,
		prematureAccessTokenRefreshThreshold:  0,
	}

	jwt, err := connector.client.LoginClient(context.Background(), clientID, clientSecret, loginRealm)
	if err != nil {
		return nil, err
	}
	rptResult, err := connector.client.RetrospectToken(context.Background(), jwt.AccessToken, clientID, clientSecret, loginRealm)
	if err != nil {
		panic("Inspection failed:" + err.Error())
	}
	if rptResult != nil && !*rptResult.Active {
		panic("Token is not active")
	}

	connector.token = jwt

	return connector, nil
}

func (k *KeycloakConnector) isAccessTokenValid() bool {
	if k.token == nil {
		return false
	}

	if k.lastRequest.IsZero() {
		return false
	}

	sessionExpiry := k.token.ExpiresIn - k.prematureAccessTokenRefreshThreshold
	if int(time.Since(k.lastRequest).Seconds()) > sessionExpiry {
		return false
	}

	token, _, err := k.client.DecodeAccessToken(context.Background(), k.token.AccessToken, k.loginRealm)
	return err == nil && token.Valid
}

func (k *KeycloakConnector) isRefreshTokenValid() bool {
	if k.token == nil {
		return false
	}

	if k.lastRequest.IsZero() {
		return false
	}

	sessionExpiry := k.token.RefreshExpiresIn - k.prematureRefreshTokenRefreshThreshold
	return int(time.Since(k.lastRequest).Seconds()) > sessionExpiry
}

func (k *KeycloakConnector) authenticate(ctx context.Context) error {
	k.lastRequest = time.Now()

	jwt, err := k.client.LoginClient(ctx, k.clientID, k.clientSecret, k.loginRealm)
	if err != nil {
		return err
	}
	k.token = jwt

	return nil
}

func (k *KeycloakConnector) refresh(ctx context.Context) error {
	k.lastRequest = time.Now()

	jwt, err := k.client.RefreshToken(ctx, k.token.RefreshToken, "admin-cli", "", k.loginRealm)
	if err != nil {
		return err
	}
	k.token = jwt

	return nil
}

func (k *KeycloakConnector) GetKeycloakAuthToken(ctx context.Context) (*gocloak.JWT, error) {
	if k.isAccessTokenValid() {
		return k.token, nil
	}

	if k.isRefreshTokenValid() {
		err := k.refresh(ctx)
		if err == nil {
			return k.token, nil
		}
	}

	err := k.authenticate(ctx)
	if err != nil {
		return nil, err
	}

	return k.token, nil
}

func (k *KeycloakConnector) GetUserGroups(ctx context.Context, userID string) ([]*gocloak.Group, error) {
	token, err := k.GetKeycloakAuthToken(ctx)
	if err != nil {
		return nil, err
	}

	briefRepresentation := false

	// Retrieve groups of a user from Keycloak
	return k.client.GetUserGroups(ctx, token.AccessToken, k.userRealm, userID, gocloak.GetGroupsParams{
		BriefRepresentation: &briefRepresentation,
	})
}

func (k *KeycloakConnector) GetUser(ctx context.Context, userID string) (*gocloak.User, error) {
	token, err := k.GetKeycloakAuthToken(ctx)
	if err != nil {
		return nil, err
	}

	// Retrieve user from Keycloak
	return k.client.GetUserByID(ctx, token.AccessToken, k.userRealm, userID)
}

func breakglassAttributeForGroup(group gocloak.Group) string {
	// Build breakglass user key like "breakglass-<groupname>"
	return fmt.Sprintf("breakglass-%s", *group.Name)
}

func (k *KeycloakConnector) GetActiveBreakglass(ctx context.Context, userID string) ([]BreakglassState, error) {
	states := []BreakglassState{}

	user, err := k.GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	groups, err := k.GetUserGroups(ctx, userID)
	if err != nil {
		return nil, err
	}

	attributes := user.Attributes
	if attributes == nil {
		return states, nil
	}

	for _, group := range groups {
		key := breakglassAttributeForGroup(*group)
		expiries, expiriesExisting := (*user.Attributes)[key]
		if !expiriesExisting || len(expiries) == 0 {
			continue
		}

		expiry := int64(math.MaxInt64)
		for _, expiryEntry := range expiries {
			expiryTime, err := strconv.ParseInt(expiryEntry, 10, 64)
			if err != nil {
				continue
			}
			if expiryTime < expiry {
				expiry = expiryTime
			}
		}

		states = append(states, BreakglassState{
			Group:  *group.Name,
			Expiry: expiry,
		})
	}
	return states, nil
}

func (k *KeycloakConnector) getGroupByName(ctx context.Context, groupName string) (*gocloak.Group, error) {
	token, err := k.GetKeycloakAuthToken(ctx)
	if err != nil {
		return nil, err
	}
	// Search for group by name
	briefRepresentation := false
	groups, err := k.client.GetGroups(ctx, token.AccessToken, k.userRealm, gocloak.GetGroupsParams{
		Search:              &groupName,
		BriefRepresentation: &briefRepresentation,
	})
	if err != nil {
		return nil, err
	}

	var group *gocloak.Group
	for _, g := range groups {
		if g.Name != nil && *g.Name == groupName {
			if group != nil {
				return nil, fmt.Errorf("found multiple groups called %s", groupName)
			}
			group = g
		}
	}
	if group == nil {
		return nil, fmt.Errorf("group not found")
	}
	return group, nil
}

func (k *KeycloakConnector) PersistBreakglass(ctx context.Context, userID, groupName string, duration int64) error {
	token, err := k.GetKeycloakAuthToken(ctx)
	if err != nil {
		return err
	}

	group, err := k.getGroupByName(ctx, groupName)
	if err != nil {
		return err
	}

	// Get Keycloak user by userID
	user, err := k.GetUser(ctx, userID)
	if err != nil {
		return err
	}
	attributes := user.Attributes
	// Init attributes when nil
	if attributes == nil {
		attributes = &map[string][]string{}
	}

	expiryTime := time.Now().Add(time.Duration(duration) * time.Second).Unix()

	// Set breakglass-<groupname> to expiry time
	key := breakglassAttributeForGroup(*group)
	(*attributes)[key] = []string{strconv.FormatInt(expiryTime, 10)}

	// Only update / set attributes of user
	user.Attributes = attributes

	// First update user
	err = k.client.UpdateUser(ctx, token.AccessToken, k.userRealm, *user)
	if err != nil {
		return err
	}
	// Add user to group
	return k.client.AddUserToGroup(ctx, token.AccessToken, k.userRealm, *user.ID, *group.ID)
}

func (k *KeycloakConnector) GetGroupMembers(ctx context.Context, group gocloak.Group) ([]*gocloak.User, error) {
	token, err := k.GetKeycloakAuthToken(ctx)
	if err != nil {
		return nil, err
	}

	briefRepresentation := false

	// Get all members of a group
	return k.client.GetGroupMembers(ctx, token.AccessToken, k.userRealm, *group.ID, gocloak.GetGroupsParams{
		BriefRepresentation: &briefRepresentation,
	})
}

func (k *KeycloakConnector) RemoveUserFromBreakglassGroup(ctx context.Context, group gocloak.Group, user gocloak.User) error {
	token, err := k.GetKeycloakAuthToken(ctx)
	if err != nil {
		return err
	}
	attributes := user.Attributes

	// Remove specific group attribute from user
	if attributes != nil {
		key := breakglassAttributeForGroup(group)
		delete(*attributes, key)

		// Only update attributes of user
		user.Attributes = attributes

		// Update user
		err := k.client.UpdateUser(ctx, token.AccessToken, k.userRealm, user)
		if err != nil {
			return err
		}

	}
	// Remove User from group
	return k.client.DeleteUserFromGroup(ctx, token.AccessToken, k.userRealm, *user.ID, *group.ID)
}

func (k *KeycloakConnector) DropBreakglass(ctx context.Context, userID, groupName string) error {
	activeBreakglasses, err := k.GetActiveBreakglass(ctx, userID)
	if err != nil {
		return err
	}

	var selectedState *BreakglassState

	for _, state := range activeBreakglasses {
		if state.Group == groupName {
			selectedState = &state
			break
		}
	}
	if selectedState == nil {
		return fmt.Errorf("user has no active breakglass in group %s", groupName)
	}

	group, err := k.getGroupByName(ctx, groupName)
	if err != nil {
		return err
	}
	user, err := k.GetUser(ctx, userID)
	if err != nil {
		return err
	}
	return k.RemoveUserFromBreakglassGroup(ctx, *group, *user)
}

func (k *KeycloakConnector) SearchGroups(ctx context.Context, query string) ([]*gocloak.Group, error) {
	token, err := k.GetKeycloakAuthToken(ctx)
	if err != nil {
		return nil, err
	}

	return k.client.GetGroups(ctx, token.AccessToken, k.userRealm, gocloak.GetGroupsParams{
		Search: &query,
	})
}
