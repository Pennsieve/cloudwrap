// Package config models the (environment, service) pair that identifies a set
// of parameters in SSM Parameter Store.
package config

import "fmt"

// Config identifies a group of parameters nested under an environment and a
// service, e.g. the path "/staging/auth-service/".
type Config struct {
	Environment string
	Service     string
}

// New returns a Config for the given environment and service.
func New(environment, service string) Config {
	return Config{Environment: environment, Service: service}
}

// Path returns the SSM path prefix for this config, always with a trailing
// slash: "/{environment}/{service}/".
func (c Config) Path() string {
	return fmt.Sprintf("/%s/%s/", c.Environment, c.Service)
}
