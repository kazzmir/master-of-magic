package multiplayer

import (
	"sync"
	"testing"

	"github.com/kazzmir/master-of-magic/game/network-combat/protocol"
)

func TestInvitationCancelRacesWithAccept(t *testing.T) {
	manager, users := testGameManager()
	invitationCommand(manager, users[1], 1, protocol.Message{Type: "request-game", TargetID: 2})
	invite := pendingInvitation(t, manager)
	var wait sync.WaitGroup
	wait.Go(func() {
		invitationCommand(manager, users[2], 2, protocol.Message{Type: "respond-game-invite", InviteID: invite.id, Accept: true})
	})
	wait.Go(func() {
		invitationCommand(manager, users[1], 1, protocol.Message{Type: "cancel-game-invite", InviteID: invite.id})
	})
	wait.Wait()
	if len(manager.Games) != 0 || len(manager.Invitations) != 0 || users[1].gameID != 0 || users[2].gameID != 0 {
		t.Fatal("Exit did not release participants when racing with acceptance")
	}
}

func TestStaleInvitationCancelCannotExitNewGame(t *testing.T) {
	manager, users := testGameManager()
	invitationCommand(manager, users[1], 1, protocol.Message{Type: "request-game", TargetID: 2})
	invite := pendingInvitation(t, manager)
	invitationCommand(manager, users[2], 2, protocol.Message{Type: "respond-game-invite", InviteID: invite.id, Accept: true})
	gameCommand(manager, users[1], 1, protocol.Message{Type: "leave-game", GameID: users[1].gameID})
	gameCommand(manager, users[1], 1, protocol.Message{Type: "create-game"})
	newGameID := users[1].gameID
	invitationCommand(manager, users[1], 1, protocol.Message{Type: "cancel-game-invite", InviteID: invite.id})
	requireGameError(t, users[1])
	if users[1].gameID != newGameID || manager.Games[newGameID] == nil {
		t.Fatal("stale cancellation exited a different game")
	}
}
