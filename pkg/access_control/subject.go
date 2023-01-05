package access_control

type Subject interface {
	Roles() []Role
}
