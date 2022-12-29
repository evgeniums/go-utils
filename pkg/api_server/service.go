package api_server

import "github.com/evgeniums/go-backend-helpers/pkg/common"

// Interface of service that implements one or more groups of API endpoints.
type Service interface {
	common.WithNameAndPath

	// Add endpoint to default service group.
	AddEndpoint(endpoint Endpoint) error

	IsMultiTenance() bool
}
