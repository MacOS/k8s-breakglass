package breakglass

import (
	"context"
	"strings"

	"github.com/Nerzal/gocloak/v13"
	"gitlab.devops.telekom.de/schiff/engine/go-breakglass.git/pkg/config"
)

const (
	AppownerSuffix     = "-appowner"
	CollaboratorSuffix = "-collaborator"
	PoweruserSuffix    = "-poweruser"
	DebugSuffix        = "-debug"
)

func (b *BreakglassController) discoverTransitions(ctx context.Context) ([]config.Transition, error) {
	appownerGroups, err := b.keycloak.SearchGroups(ctx, AppownerSuffix)
	if err != nil {
		return nil, err
	}

	collaboratorGroups, err := b.keycloak.SearchGroups(ctx, CollaboratorSuffix)
	if err != nil {
		return nil, err
	}

	poweruserGroups, err := b.keycloak.SearchGroups(ctx, PoweruserSuffix)
	if err != nil {
		return nil, err
	}

	powerGroups, err := b.keycloak.SearchGroups(ctx, DebugSuffix)
	if err != nil {
		return nil, err
	}

	transitions := []config.Transition{}
	for _, g := range appownerGroups {
		if g.Name == nil {
			continue
		}
		gn := *g.Name
		tenant := strings.TrimSuffix(gn, AppownerSuffix)
		gnDebug := tenant + DebugSuffix
		if findGroup(powerGroups, gnDebug) == nil {
			continue
		}

		gnCollaborator := tenant + CollaboratorSuffix
		if findGroup(collaboratorGroups, gnCollaborator) != nil {
			transitions = append(transitions, config.Transition{
				From:           gnCollaborator,
				To:             gnDebug,
				ApprovalGroups: []string{gn},
				Duration:       7200,
				SelfApproval:   true,
			})
		}

		gnPoweruser := tenant + PoweruserSuffix
		if findGroup(poweruserGroups, gnPoweruser) != nil {
			transitions = append(transitions, config.Transition{
				From:           gnPoweruser,
				To:             gnDebug,
				ApprovalGroups: []string{gn},
				Duration:       7200,
				SelfApproval:   true,
				// exclude this one since ther eis no benefit in having two global transitions
				// that escalate to the same target group
				GlobalBreakglassExcluded: true,
			})
		}
	}

	return transitions, nil
}

func findGroup(groups []*gocloak.Group, name string) *gocloak.Group {
	for _, g := range groups {
		if g.Name == nil {
			continue
		}
		if *g.Name == name {
			return g
		}
	}
	return nil
}
