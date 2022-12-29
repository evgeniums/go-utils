package api_server

import "github.com/evgeniums/go-backend-helpers/pkg/common"

// Interface of auth part of request and response.
type Auth interface {
	common.WithParameters
}

// Base interface for types wuth auth section.
type WithAuth interface {
	Auth() Auth
}
