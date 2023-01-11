package api_server

// Interface of auth part of request.
type AuthRequest interface {
	GetAuthParameter(key string) string
}

// Interface of auth part of response.
type AuthResponse interface {
	SetAuthParameter(key string, value string)
}
