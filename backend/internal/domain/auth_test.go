package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRole_Valid_OnlyKnownRoles(t *testing.T) {
	assert.True(t, RoleAdmin.Valid())
	assert.True(t, RoleCajero.Valid())
	assert.False(t, Role("root").Valid())
}

func TestUser_IsLocked_ByTime(t *testing.T) {
	now := time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC)
	future, past := now.Add(time.Second), now.Add(-time.Second)
	assert.False(t, User{}.IsLocked(now))
	assert.True(t, User{LockedUntil: &future}.IsLocked(now))
	assert.False(t, User{LockedUntil: &past}.IsLocked(now))
	assert.False(t, User{LockedUntil: &now}.IsLocked(now))
}

func TestUser_WithFailedLogin_LocksOnFifthFailure(t *testing.T) {
	now := time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC)
	u := User{}
	for range MaxFailedLogins - 1 {
		u = u.WithFailedLogin(now)
		assert.Nil(t, u.LockedUntil)
	}
	u = u.WithFailedLogin(now)
	assert.Equal(t, int32(5), u.FailedLogins)
	require.NotNil(t, u.LockedUntil)
	assert.Equal(t, now.Add(15*time.Minute), *u.LockedUntil)
}

func TestUser_WithFailedLogin_ExpiredLock_RestartsCount(t *testing.T) {
	now := time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC)
	past := now.Add(-time.Minute)
	u := User{FailedLogins: 5, LockedUntil: &past}.WithFailedLogin(now)
	assert.Equal(t, int32(1), u.FailedLogins)
	assert.Nil(t, u.LockedUntil)
}
