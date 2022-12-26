package api_server

// Interface of auth part of request and response.
type Auth interface {
	WithParameters
}

// Base interface for types wuth auth section.
type WithAuth interface {
	Auth() Auth
}
