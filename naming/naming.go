package naming

import (
	"errors"

	"jinv/kim"
)

var (
	ErrNotFound = errors.New("service no found")
)

type Naming interface {
	Find(serviceName string, tags ...string) ([]kim.ServiceRegistration, error)
	Subscribe(serviceName string, callback func(services []kim.ServiceRegistration)) error
	Unsubscribe(serviceName string) error
	Register(service kim.ServiceRegistration) error
	Deregister(serviceID string) error
}
