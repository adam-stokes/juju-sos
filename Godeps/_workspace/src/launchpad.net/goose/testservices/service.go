package testservices

import (
	"launchpad.net/goose/testservices/hook"
	"launchpad.net/goose/testservices/identityservice"
	"net/http"
)

// An HttpService provides the HTTP API for a service double.
type HttpService interface {
	SetupHTTP(mux *http.ServeMux)
}

// A ServiceInstance is an Openstack module, one of nova, swift, glance.
type ServiceInstance struct {
	identityservice.ServiceProvider
	hook.TestService
	IdentityService identityservice.IdentityService
	Scheme          string
	Hostname        string
	VersionPath     string
	TenantId        string
	Region          string
}

// Internal Openstack errors.

var RateLimitExceededError = NewRateLimitExceededError()

// NoMoreFloatingIPs corresponds to "HTTP 404 Zero floating ips available."
var NoMoreFloatingIPs = NewNoMoreFloatingIpsError()

// IPLimitExceeded corresponds to "HTTP 413 Maximum number of floating ips exceeded"
var IPLimitExceeded = NewIPLimitExceededError()

// AvailabilityZoneIsNotAvailable corresponds to
// "HTTP 400 The requested availability zone is not available"
var AvailabilityZoneIsNotAvailable = NewAvailabilityZoneIsNotAvailableError()