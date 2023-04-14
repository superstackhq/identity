package organization

import (
	"context"
	"fmt"

	"github.com/kamva/mgm/v3"
	"github.com/kamva/mgm/v3/field"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Manager struct {
}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) Save(ctx context.Context, name string, creatorID string) (*Organization, error) {
	organization := &Organization{
		Name:      name,
		CreatorID: creatorID,
		Deleted:   false,
	}

	err := mgm.Coll(organization).CreateWithCtx(ctx, organization)

	if err != nil {
		return nil, err
	}

	return organization, nil
}

func (m *Manager) GetByName(ctx context.Context, name string) (*Organization, error) {
	organization := &Organization{}

	err := mgm.Coll(organization).FirstWithCtx(ctx, bson.M{
		"name":    name,
		"deleted": false,
	}, organization)

	if err == mongo.ErrNoDocuments {
		return nil, fmt.Errorf("organization %s not found", name)
	}

	if err != nil {
		return nil, err
	}

	return organization, nil
}

func (m *Manager) Get(ctx context.Context, organizationID string) (*Organization, error) {
	id, err := primitive.ObjectIDFromHex(organizationID)

	if err != nil {
		return nil, err
	}

	organization := &Organization{}

	err = mgm.Coll(organization).FirstWithCtx(ctx, bson.M{
		field.ID:  id,
		"deleted": false,
	}, organization)

	if err == mongo.ErrNoDocuments {
		return nil, fmt.Errorf("organization not found")
	}

	if err != nil {
		return nil, err
	}

	return organization, nil
}

func (m *Manager) NameExists(ctx context.Context, name string) (bool, error) {
	count, err := mgm.Coll(&Organization{}).CountDocuments(ctx, bson.M{
		"name":    name,
		"deleted": false,
	})

	if err != nil {
		return false, err
	}

	return count != 0, nil
}
