package authentication

import (
	"github.com/superstackhq/identity/pkg/actor"
)

type AuthenticatedActor struct {
	ActorType      actor.Type
	ActorID        string
	OrganizationID string
	HasFullAccess  bool
}
