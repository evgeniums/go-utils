package access_control

import "net/http"

type AccessType uint32
type Operation uint32

const (
	Read          AccessType = 0x1
	Create        AccessType = 0x2
	UpdateReplace AccessType = 0x4
	UpdatePartial AccessType = 0x8
	Update        AccessType = UpdateReplace | UpdatePartial

	Delete AccessType = 0x10

	Post  = Create
	Get   = Read
	Put   = UpdateReplace
	Patch = UpdatePartial

	All AccessType = 0xFFFFFFFF
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

func Check(accessMask AccessType, accessType AccessType) bool {
	return (uint32(accessMask) & uint32(accessType)) != 0
}

func (a *AccessBase) Check(accessType AccessType) bool {
	return (a.mask & uint32(accessType)) != 0
}

func (a *AccessBase) Mask() uint32 {
	return a.mask
}

var httpMethods2AccessTypes = map[string]AccessType{
	http.MethodGet:    Read,
	http.MethodPost:   Create,
	http.MethodPut:    UpdateReplace,
	http.MethodPatch:  UpdatePartial,
	http.MethodDelete: Delete,
}

func HttpMethod2Access(method string) AccessType {
	at, ok := httpMethods2AccessTypes[method]
	if !ok {
		at = 0
	}
	return at
}

func Access2HttpMethod(access AccessType) string {

	a := NewAccess(uint32(access))

	if a.Check(Read) {
		return http.MethodGet
	}

	if a.Check(Create) {
		return http.MethodPost
	}

	if a.Check(UpdateReplace) {
		return http.MethodPut
	}

	if a.Check(UpdatePartial) {
		return http.MethodPatch
	}

	if a.Check(Delete) {
		return http.MethodDelete
	}

	return ""
}

func HttpContentInQuery(accessType AccessType) bool {
	a := NewAccess(uint32(accessType))
	return a.Check(Get) || a.Check(Delete)
}
