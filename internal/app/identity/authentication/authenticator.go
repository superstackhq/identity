package authentication

import (
	"context"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/superstackhq/identity/pkg/actor"
)

type Authenticator struct {
	jwtSigningKey []byte
}

func NewAuthenticator(jwtSigningKey string) *Authenticator {
	return &Authenticator{
		jwtSigningKey: []byte(jwtSigningKey),
	}
}

const (
	BearerToken = "Bearer"
	ApiKey      = "ApiKey"
)

func (a *Authenticator) GenerateToken(userID string, organizationID string, admin bool) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":              userID,
		"admin":           admin,
		"organization_id": organizationID,
		"iss":             "superstack",
	})

	tokenString, err := token.SignedString(a.jwtSigningKey)

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (a *Authenticator) ValidateContext(c *gin.Context, ctx context.Context) (*AuthenticatedActor, error) {
	tokenType, token, err := a.extractToken(c)

	if err != nil {
		return nil, err
	}

	switch tokenType {
	case BearerToken:
		return a.validateBearerToken(token)
	case ApiKey:
		return a.validateApiKey(ctx, token)
	default:
		return nil, fmt.Errorf("invalid token type")
	}
}

func (a *Authenticator) validateApiKey(ctx context.Context, accessKey string) (*AuthenticatedActor, error) {
	return &AuthenticatedActor{
		ActorID:       "", // TODO
		ActorType:     actor.TypeApiKey,
		HasFullAccess: false,
	}, nil
}

func (a *Authenticator) validateBearerToken(tokenString string) (*AuthenticatedActor, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return a.jwtSigningKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		id, ok := claims["id"]

		if !ok {
			return nil, fmt.Errorf("invalid access token")
		}

		admin, ok := claims["admin"]

		if !ok {
			return nil, fmt.Errorf("invalid access token")
		}

		userIDString, ok := id.(string)

		if !ok {
			return nil, fmt.Errorf("invalid access token")
		}

		adminBool, ok := admin.(bool)

		if !ok {
			return nil, fmt.Errorf("invalid access token")
		}

		organizationID, ok := claims["organization_id"]

		if !ok {
			return nil, fmt.Errorf("invalid access token")
		}

		organizationIDString, ok := organizationID.(string)

		if !ok {
			return nil, fmt.Errorf("invalid access token")
		}

		return &AuthenticatedActor{
			ActorID:        userIDString,
			ActorType:      actor.TypeUser,
			OrganizationID: organizationIDString,
			HasFullAccess:  adminBool,
		}, nil
	} else {
		return nil, fmt.Errorf("invalid access token")
	}
}

func (a *Authenticator) extractToken(c *gin.Context) (string, string, error) {
	authorizationHeader := c.GetHeader("Authorization")

	if len(authorizationHeader) == 0 {
		return "", "", fmt.Errorf("authorization header is not set")
	}

	components := strings.Split(authorizationHeader, " ")

	if len(components) != 2 {
		return "", "", fmt.Errorf("invalid access token")
	}

	tokenType := components[0]

	if tokenType != BearerToken && tokenType != ApiKey {
		return "", "", fmt.Errorf("invalid bearer token")
	}

	return tokenType, components[1], nil
}
