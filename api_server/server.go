package api_server

type Server interface {
	ApiVersion() string

	Run()

	AddGroup(group Group)
}
