package access_control

type AccessType uint32

const (
	Create AccessType = 1
	Read   AccessType = 2
	Update AccessType = 4
	Delete AccessType = 8
)

type Access interface {
	Grant(accessType AccessType)
	Revoke(accessType AccessType)
	Check(accessType AccessType) bool
	Mask() uint32
}

type AccessBase struct {
	mask uint32
}

func NewAccess(mask uint32) AccessBase {
	return AccessBase{mask: mask}
}

func (a *AccessBase) Grant(accessType AccessType) {
	a.mask = a.mask | uint32(accessType)
}

func (a *AccessBase) Revoke(accessType AccessType) {
	a.mask = a.mask & ^uint32(accessType)
}

func (a *AccessBase) Check(accessType AccessType) bool {
	return (a.mask & uint32(accessType)) != 0
}

func (a *AccessBase) Mask() uint32 {
	return a.mask
}
