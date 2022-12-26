package api_server

import "github.com/evgeniums/go-backend-helpers/pkg/utils"

// Interface of auth part of request and response.
type Auth interface {
	utils.WithParameters
}

// Base interface for types wuth auth section.
type WithAuth interface {
	Auth() Auth
}
