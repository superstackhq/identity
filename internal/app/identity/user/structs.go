package user

import (
	"github.com/kamva/mgm/v3"
	"github.com/superstackhq/identity/pkg/actor"
)

type User struct {
	mgm.DefaultModel `bson:",inline"`
	Username         string     `json:"username" bson:"username"`
	Password         string     `json:"-" bson:"password"`
	OrganizationID   string     `json:"organization_id" bson:"organization_id"`
	Admin            bool       `json:"admin" bson:"admin"`
	CreatorType      actor.Type `json:"creator_type" bson:"creator_type"`
	CreatorID        string     `json:"creator_id" bson:"creator_id"`
	Deleted          bool       `json:"deleted" bson:"deleted"`
}
