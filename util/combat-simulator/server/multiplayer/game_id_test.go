package multiplayer

import "testing"

func TestRandomGameIDs(t *testing.T) {
	manager := &Manager{Games: make(map[uint64]*Game)}
	manager.lock.With(func() error {
		for range 2000 {
			id := manager.newGameID()
			if id <= 1_000_000_000 {
				t.Fatalf("game ID %d is not above one billion", id)
			}
			if manager.Games[id] != nil {
				t.Fatalf("duplicate game ID %d", id)
			}
			manager.Games[id] = &Game{id: id}
		}
		return nil
	})
}
