package access_control

import "net/http"

type AccessType uint32
type Operation uint32

const (
	Create AccessType = 1

	ReadContent AccessType = 2
	ReadMeta    AccessType = 4
	Read        AccessType = ReadContent | ReadMeta

	UpdateWhole   AccessType = 8
	UpdatePartial AccessType = 10
	Update        AccessType = UpdateWhole | UpdatePartial

	Delete AccessType = 20

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

func (a *AccessBase) Check(accessType AccessType) bool {
	return (a.mask & uint32(accessType)) != 0
}

func (a *AccessBase) Mask() uint32 {
	return a.mask
}

var httpMethods2AccessTypes = map[string]AccessType{
	http.MethodGet:     ReadContent,
	http.MethodPost:    Create,
	http.MethodPut:     UpdateWhole,
	http.MethodPatch:   UpdatePartial,
	http.MethodDelete:  Delete,
	http.MethodOptions: ReadMeta,
}

func HttpMethod2Access(method string) AccessType {
	at, ok := httpMethods2AccessTypes[method]
	if !ok {
		at = 0
	}
	return at
}

var accessTypes2HttpMethods = map[AccessType]string{
	ReadContent:   http.MethodGet,
	Create:        http.MethodPost,
	UpdateWhole:   http.MethodPut,
	UpdatePartial: http.MethodPatch,
	Delete:        http.MethodDelete,
	ReadMeta:      http.MethodOptions,
}

func Access2HttpMethod(access AccessType) string {
	method := accessTypes2HttpMethods[access]
	return method
}
