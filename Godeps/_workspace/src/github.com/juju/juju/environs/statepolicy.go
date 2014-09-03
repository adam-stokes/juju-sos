// Copyright 2014 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package environs

import (
	"github.com/juju/errors"

	"github.com/juju/juju/constraints"
	"github.com/juju/juju/environs/config"
	"github.com/juju/juju/state"
)

// environStatePolicy implements state.Policy in
// terms of environs.Environ and related types.
type environStatePolicy struct{}

var _ state.Policy = environStatePolicy{}

// NewStatePolicy returns a state.Policy that is
// implemented in terms of Environ and related
// types.
func NewStatePolicy() state.Policy {
	return environStatePolicy{}
}

func (environStatePolicy) Prechecker(cfg *config.Config) (state.Prechecker, error) {
	// Environ implements state.Prechecker.
	return New(cfg)
}

func (environStatePolicy) ConfigValidator(providerType string) (state.ConfigValidator, error) {
	// EnvironProvider implements state.ConfigValidator.
	return Provider(providerType)
}

func (environStatePolicy) EnvironCapability(cfg *config.Config) (state.EnvironCapability, error) {
	// Environ implements state.EnvironCapability.
	return New(cfg)
}

func (environStatePolicy) ConstraintsValidator(cfg *config.Config) (constraints.Validator, error) {
	env, err := New(cfg)
	if err != nil {
		return nil, err
	}
	return env.ConstraintsValidator()
}

func (environStatePolicy) InstanceDistributor(cfg *config.Config) (state.InstanceDistributor, error) {
	env, err := New(cfg)
	if err != nil {
		return nil, err
	}
	if p, ok := env.(state.InstanceDistributor); ok {
		return p, nil
	}
	return nil, errors.NotImplementedf("InstanceDistributor")
}
