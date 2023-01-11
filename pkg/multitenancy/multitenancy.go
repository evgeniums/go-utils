package multitenancy

import "errors"

type Multitenancy interface {

	// Check if multiple tenancies are enabled
	IsMultiTenancy() bool

	// Find tenancy by ID.
	Tenancy(id string) (Tenancy, error)

	// Add tenancy.
	AddTenancy(id string) error

	// Remove tenance.
	RemoveTenancy(id string) error
}

type MultitenancyBaseConfig struct {
	MULTITENANCY bool
}

func (s *MultitenancyBaseConfig) IsMultiTenancy() bool {
	return s.MULTITENANCY
}

type MultitenancyBase struct {
	tenanciesById   map[string]Tenancy
	tenanciesByPath map[string]Tenancy
}

func (s *MultitenancyBase) Tenancy(id string) (Tenancy, error) {
	tenancy, ok := s.tenanciesById[id]
	if !ok {
		return nil, errors.New("unknown tenancy")
	}
	return tenancy, nil
}

func (s *MultitenancyBase) TenancyByPath(path string) (Tenancy, error) {
	tenancy, ok := s.tenanciesByPath[path]
	if !ok {
		return nil, errors.New("tenancy not found")
	}
	return tenancy, nil
}

func (s *MultitenancyBase) AddTenancy(id string) error {
	return errors.New("not implemented yet")
}

func (s *MultitenancyBase) RemoveTenancy(id string) error {
	return errors.New("not implemented yet")
}
