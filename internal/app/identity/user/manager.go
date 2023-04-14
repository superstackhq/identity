package user

import (
	"context"
	"fmt"

	"github.com/kamva/mgm/v3"
	"github.com/kamva/mgm/v3/field"
	"github.com/sethvargo/go-password/password"
	"github.com/superstackhq/identity/internal/app/identity/authentication"
	"github.com/superstackhq/identity/internal/app/identity/organization"
	"github.com/superstackhq/identity/pkg/user"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

type Manager struct {
	organizationManager *organization.Manager
	authenticator       *authentication.Authenticator
}

func NewManager(organizationManager *organization.Manager, authenticator *authentication.Authenticator) *Manager {
	return &Manager{
		organizationManager: organizationManager,
		authenticator:       authenticator,
	}
}

func (m *Manager) SignUp(ctx context.Context, signUpRequest *user.SignUpRequest) (*User, error) {

	organizationExists, err := m.organizationManager.NameExists(ctx, signUpRequest.OrganizationName)

	if err != nil {
		return nil, err
	}

	if organizationExists {
		return nil, fmt.Errorf("organization %s already exists", signUpRequest.OrganizationName)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(signUpRequest.Password), bcrypt.DefaultCost)

	if err != nil {
		return nil, err
	}

	user := &User{
		Username:       signUpRequest.Username,
		Password:       string(hashedPassword),
		OrganizationID: "",
		Admin:          true,
		CreatorType:    "",
		CreatorID:      "",
	}

	err = mgm.Coll(user).CreateWithCtx(ctx, user)

	if err != nil {
		return nil, err
	}

	org, err := m.organizationManager.Save(ctx, signUpRequest.OrganizationName, user.ID.Hex())

	if err != nil {
		return nil, err
	}

	user.OrganizationID = org.ID.Hex()

	err = mgm.Coll(user).UpdateWithCtx(ctx, user)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (m *Manager) Authenticate(ctx context.Context, authenticationRequest *user.AuthenticationRequest) (*user.AuthenticationResponse, error) {
	org, err := m.organizationManager.GetByName(ctx, authenticationRequest.OrganizationName)

	if err != nil {
		return nil, err
	}

	u := &User{}

	err = mgm.Coll(u).FirstWithCtx(ctx, bson.M{
		"organization_id": org.ID.Hex(),
		"username":        authenticationRequest.Username,
		"deleted":         false,
	}, u)

	if err == mongo.ErrNoDocuments {
		return nil, fmt.Errorf("invalid username and password combination")
	}

	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(authenticationRequest.Password))

	if err != nil {
		return nil, fmt.Errorf("invalid username and password combination")
	}

	token, err := m.authenticator.GenerateToken(u.ID.Hex(), u.OrganizationID, u.Admin)

	if err != nil {
		return nil, err
	}

	return &user.AuthenticationResponse{
		Token: token,
	}, nil
}

func (m *Manager) Get(ctx context.Context, userID string) (*User, error) {
	id, err := primitive.ObjectIDFromHex(userID)

	if err != nil {
		return nil, err
	}

	user := &User{}

	err = mgm.Coll(user).FirstWithCtx(ctx, bson.M{
		field.ID:  id,
		"deleted": false,
	}, user)

	if err == mongo.ErrNoDocuments {
		return nil, fmt.Errorf("user not found")
	}

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (m *Manager) GetByOrganization(ctx context.Context, userID string, organizationID string) (*User, error) {
	id, err := primitive.ObjectIDFromHex(userID)

	if err != nil {
		return nil, err
	}

	user := &User{}

	err = mgm.Coll(user).FirstWithCtx(ctx, bson.M{
		field.ID:          id,
		"organization_id": organizationID,
		"deleted":         false,
	}, user)

	if err == mongo.ErrNoDocuments {
		return nil, fmt.Errorf("user not found")
	}

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (m *Manager) ChangePassword(ctx context.Context, userID string, passwordChangeRequest *user.PasswordChangeRequest) (*User, error) {
	user, err := m.Get(ctx, userID)

	if err != nil {
		return nil, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(passwordChangeRequest.Password), bcrypt.DefaultCost)

	if err != nil {
		return nil, err
	}

	user.Password = string(hashedPassword)

	err = mgm.Coll(user).UpdateWithCtx(ctx, user)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (m *Manager) Add(ctx context.Context, userAdditionRequest *user.AdditionRequest, actor *authentication.AuthenticatedActor) (*user.PasswordResponse, error) {
	usernameExists, err := m.usernameExists(ctx, userAdditionRequest.Username, actor.OrganizationID)

	if err != nil {
		return nil, err
	}

	if usernameExists {
		return nil, fmt.Errorf("username %s is already taken", userAdditionRequest.Username)
	}

	pass, err := password.Generate(16, 4, 2, false, false)

	if err != nil {
		return nil, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)

	if err != nil {
		return nil, err
	}

	u := &User{
		Username:       userAdditionRequest.Username,
		Password:       string(hashedPassword),
		Admin:          userAdditionRequest.Admin,
		CreatorType:    actor.ActorType,
		CreatorID:      actor.ActorID,
		OrganizationID: actor.OrganizationID,
		Deleted:        false,
	}

	err = mgm.Coll(u).CreateWithCtx(ctx, u)

	if err != nil {
		return nil, err
	}

	return &user.PasswordResponse{Password: pass}, nil
}

func (m *Manager) Delete(ctx context.Context, userID string, organizationID string) (*User, error) {
	u, err := m.GetByOrganization(ctx, userID, organizationID)

	if err != nil {
		return nil, err
	}

	u.Deleted = true

	err = mgm.Coll(u).UpdateWithCtx(ctx, u)

	if err != nil {
		return nil, err
	}

	return u, nil
}

func (m *Manager) List(ctx context.Context, organizationID string, page int64, size int64) ([]*User, error) {
	var users []*User

	err := mgm.Coll(&User{}).SimpleFindWithCtx(ctx, &users, bson.M{
		"organization_id": organizationID,
		"deleted":         false,
	}, options.Find().SetSkip(page*size).SetLimit(size))

	if err != nil {
		return nil, err
	}

	return users, nil
}

func (m *Manager) ResetPassword(ctx context.Context, userID string, organizationID string) (*user.PasswordResponse, error) {
	u, err := m.GetByOrganization(ctx, userID, organizationID)

	if err != nil {
		return nil, err
	}

	pass, err := password.Generate(16, 4, 2, false, false)

	if err != nil {
		return nil, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)

	if err != nil {
		return nil, err
	}

	u.Password = string(hashedPassword)

	err = mgm.Coll(u).UpdateWithCtx(ctx, u)

	if err != nil {
		return nil, err
	}

	return &user.PasswordResponse{Password: pass}, nil
}

func (m *Manager) ChangeAdmin(ctx context.Context, userID string, changeAdminRequest user.AdminChangeRequest, organizationID string) (*User, error) {
	u, err := m.GetByOrganization(ctx, userID, organizationID)

	if err != nil {
		return nil, err
	}

	u.Admin = changeAdminRequest.Admin

	err = mgm.Coll(u).UpdateWithCtx(ctx, u)

	if err != nil {
		return nil, err
	}

	return u, nil
}

func (m *Manager) usernameExists(ctx context.Context, username string, organizationID string) (bool, error) {
	count, err := mgm.Coll(&User{}).CountDocuments(ctx, bson.M{
		"username":        username,
		"organization_id": organizationID,
		"deleted":         false,
	})

	if err != nil {
		return false, err
	}

	return count != 0, nil
}
