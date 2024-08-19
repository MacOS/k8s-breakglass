package breakglass

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/Nerzal/gocloak/v13"
	"gitlab.devops.telekom.de/schiff/engine/go-breakglass.git/pkg/config"
	"go.uber.org/zap"
)

type CleanupRoutine struct {
	log        *zap.SugaredLogger
	breakglass *BreakglassController
}

func (cr CleanupRoutine) cleanupRoutine() {
	for {
		cr.log.Info("Running cleanup task")
		cr.cleanupTask(context.Background())
		cr.log.Info("Finished cleanup task")
		time.Sleep(time.Minute)
	}
}

func (cr CleanupRoutine) cleanupTask(ctx context.Context) {
	transitions, err := cr.breakglass.getTransitions()
	if err != nil {
		cr.log.Errorf("failed to fetch transitions for cleanup: %v", err)
		return
	}
	groups := getTargetGroups(transitions)

	// Iterate over groups
	for _, gn := range groups {
		cr.log.Infof("Cleaning up group %s", gn)
		group, err := cr.breakglass.keycloak.getGroupByName(ctx, gn)
		if err != nil {
			cr.log.Errorf("failed to fetch group %v during cleanup: %v", gn, err)
			continue
		}
		err = cr.cleanupGroup(context.Background(), *group)
		if err != nil {
			cr.log.Infof("failed to clean up group %v during cleanup: %v", *group.Name, err)
			continue
		}
	}
}

func getTargetGroups(transitions []config.Transition) []string {
	g := map[string]struct{}{}
	for _, t := range transitions {
		g[t.To] = struct{}{}
	}
	groups := make([]string, 0, len(g))
	for g := range g {
		groups = append(groups, g)
	}
	return groups
}

func (cr CleanupRoutine) cleanupGroup(ctx context.Context, group gocloak.Group) error {
	// Get all users in group
	users, err := cr.breakglass.keycloak.GetGroupMembers(ctx, group)
	if err != nil {
		return err
	}

	for _, user := range users {
		cr.log.Infof("Checking user %s of group %s", *user.Username, *group.Name)
		err := cr.cleanupUser(ctx, group, *user)
		if err != nil {
			return fmt.Errorf("error cleaning up user %s in group %s: %v", *user.Username, *group.Name, err)
		}
	}
	return nil
}

func (cr CleanupRoutine) cleanupUser(ctx context.Context, group gocloak.Group, user gocloak.User) error {
	t := time.Now().Unix()
	// Build breakglass key (breakglass-<groupname>) to retrieve attributes (expiry time in unix timestamp format)
	key := breakglassAttributeForGroup(group)

	// If user has no attributes and is in a breakglass group -> remove user, shouldn't be there
	if user.Attributes == nil {
		cr.log.Infof("Removing user %s from group %s because user is in group without tracked expiry", *user.Username, *group.Name)
		return cr.breakglass.keycloak.RemoveUserFromBreakglassGroup(ctx, group, user)
	}

	// Get expiry time and check if expiry is existing
	expiries, expiriesExisting := (*user.Attributes)[key]

	// If expiry is not existing -> remove user, shouldn't be there
	if !expiriesExisting || len(expiries) == 0 {
		cr.log.Infof("Removing user %s from group %s because user is in group without tracked expiry", *user.Username, *group.Name)
		return cr.breakglass.keycloak.RemoveUserFromBreakglassGroup(ctx, group, user)
	}

	// Iterate over all retrieved expiries
	for _, expiry := range expiries {
		expiryTime, err := strconv.ParseInt(expiry, 10, 64)
		if err != nil {
			return err
		}
		// If one (we should really just have one) time is equal or before current time -> remove user, shouldn't be in the group anymore
		if expiryTime <= t {
			cr.log.Infof("Removing user %s from group %s because time expired (%d <= %d)", *user.Username, *group.Name, expiryTime, t)
			return cr.breakglass.keycloak.RemoveUserFromBreakglassGroup(ctx, group, user)
		}
	}
	return nil
}
