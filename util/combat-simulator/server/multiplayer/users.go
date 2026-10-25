package multiplayer

// Reserve before authentication so unauthenticated connections cannot bypass
// the global limit. Combat sockets belong to an existing user and are excluded.
func (manager *Manager) reserveUserSlot() bool {
	reserved, _ := manager.lock.With2(func() (bool, error) {
		if manager.userSlots >= MaxUsers {
			return false, nil
		}
		manager.userSlots++
		return true, nil
	})
	return reserved
}

func (manager *Manager) releaseUserSlot() {
	manager.lock.With(func() error { manager.userSlots--; return nil })
}
