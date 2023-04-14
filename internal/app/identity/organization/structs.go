package organization

import "github.com/kamva/mgm/v3"

type Organization struct {
	mgm.DefaultModel `bson:",inline"`
	Name             string `json:"name" bson:"name"`
	CreatorID        string `json:"creator_id" bson:"creator_id"`
	Deleted          bool   `json:"deleted" bson:"deleted"`
}
