package main

import (
	"github.com/superstackhq/common/env"
	"github.com/superstackhq/identity/internal/app/identity"
)

func main() {
	identity.NewServer(&identity.Config{
		Host:          env.GetOrDefault("HOST", "0.0.0.0"),
		Port:          env.GetOrDefault("PORT", "8000"),
		MongoEndpoint: env.GetOrDefault("MONGO_ENDPOINT", "mongodb://localhost:27017"),
		MongoDatabase: env.GetOrDefault("MONGO_DATABASE", "identity"),
		JwtSecretKey:  env.GetOrDefault("JWT_SECRET_KEY", "secret"),
	}).Start()
}
