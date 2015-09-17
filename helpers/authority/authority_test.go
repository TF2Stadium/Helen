// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package authority

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	ActionOne   AuthAction = iota
	ActionTwo   AuthAction = iota
	ActionThree AuthAction = iota
	ActionFour  AuthAction = iota
)

const (
	RoleNormal AuthRole = iota
	RoleAdmin  AuthRole = iota
)

func TestReset(t *testing.T) {
	RoleNormal.Allow(ActionOne)
	Reset()
	assert.False(t, RoleNormal.Can(ActionOne))
}

func TestSimpleCanCannot(t *testing.T) {
	assert.False(t, RoleNormal.Can(ActionOne))

	RoleNormal.Allow(ActionOne)
	assert.True(t, RoleNormal.Can(ActionOne))
	assert.True(t, Can(int(RoleNormal), ActionOne))

	RoleNormal.Disallow(ActionOne)
	assert.False(t, RoleNormal.Can(ActionOne))
}

func TestInherit(t *testing.T) {
	RoleNormal.Allow(ActionOne).Allow(ActionTwo)

	RoleAdmin.Inherit(RoleNormal).Allow(ActionThree)

	assert.True(t, RoleAdmin.Can(ActionOne))
	assert.True(t, RoleAdmin.Can(ActionTwo))
	assert.True(t, RoleAdmin.Can(ActionThree))
	assert.False(t, RoleAdmin.Can(ActionFour))

	RoleAdmin.Disallow(ActionTwo)
	assert.False(t, RoleAdmin.Can(ActionTwo))
}
