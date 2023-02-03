package common

type WithDescription interface {
	Description() string
	SetDescription(value string)
}

type WithDescriptionBase struct {
	DESCRIPTION string `json:"description"`
}

func (d *WithDescriptionBase) Description() string {
	return d.DESCRIPTION
}

func (d *WithDescriptionBase) SetDescription(value string) {
	d.DESCRIPTION = value
}

type WithActive interface {
	IsActive() bool
	SetActive(value bool)
}

type WithActiveBase struct {
	ACTIVE bool `gorm:"index" json:"active"`
}

func (d *WithActiveBase) IsActive() bool {
	return d.ACTIVE
}

func (d *WithActiveBase) SetActive(value bool) {
	d.ACTIVE = value
}

type WithType interface {
	Type() string
	SetType(value string)
}

type WithTypeBase struct {
	TYPE string `gorm:"index" json:"type"`
}

func (t *WithTypeBase) Type() string {
	return t.TYPE
}

func (t *WithTypeBase) SetType(value string) {
	t.TYPE = value
}

type WithRefId interface {
	RefId() string
	SetRefId(value string)
}

type WithRefIdBase struct {
	REFID string `gorm:"index" json:"refid"`
}

func (t *WithRefIdBase) RefId() string {
	return t.REFID
}

func (t *WithRefIdBase) SetRefId(value string) {
	t.REFID = value
}
