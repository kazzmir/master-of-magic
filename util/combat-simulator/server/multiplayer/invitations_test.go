package multiplayer

import (
	"strings"
	"sync"
	"testing"

	"github.com/kazzmir/master-of-magic/game/network-combat/protocol"
)

func invitationCommand(manager *Manager, user *User, id uint64, command protocol.Message) {
	manager.lock.With(func() error { manager.handleInvitationCommand(user, id, command); return nil })
}

func pendingInvitation(t *testing.T, manager *Manager) *GameInvitation {
	t.Helper()
	if len(manager.Invitations) != 1 {
		t.Fatalf("expected one request, got %d", len(manager.Invitations))
	}
	for _, invite := range manager.Invitations {
		return invite
	}
	return nil
}

func TestGameInvitationAccept(t *testing.T) {
	manager, users := testGameManager()
	invitationCommand(manager, users[1], 1, protocol.Message{Type: "request-game", TargetID: 2})
	invite := pendingInvitation(t, manager)
	if len(manager.Games) != 0 {
		t.Fatal("request created game before acceptance")
	}
	for _, pair := range []struct {
		user *User
		kind string
	}{{users[1], "invite-sent"}, {users[2], "game-invite"}} {
		messages := drainGameMessages(pair.user)
		if len(messages) != 1 || messages[0].Type != pair.kind || messages[0].InviteID != invite.id {
			t.Fatal(messages)
		}
	}
	invitationCommand(manager, users[3], 3, protocol.Message{Type: "respond-game-invite", InviteID: invite.id, Accept: true})
	requireGameError(t, users[3])
	if len(manager.Invitations) != 1 || len(manager.Games) != 0 {
		t.Fatal("third party accepted invitation")
	}
	invitationCommand(manager, users[2], 2, protocol.Message{Type: "respond-game-invite", InviteID: invite.id, Accept: true})
	if len(manager.Invitations) != 0 || len(manager.Games) != 1 || users[1].gameID == 0 || users[1].gameID != users[2].gameID {
		t.Fatal("accepted request did not create shared game")
	}
	game := manager.Games[users[1].gameID]
	if game.hostID != 1 || len(game.players) != 2 || game.ready[1] || game.ready[2] {
		t.Fatal("incorrect invited game setup")
	}
	for _, user := range []*User{users[1], users[2]} {
		closed, state := false, false
		for _, message := range drainGameMessages(user) {
			closed = closed || message.Type == "invite-closed"
			state = state || message.Type == "game-state" && message.Game != nil && message.Game.ID == game.id
		}
		if !closed || !state {
			t.Fatal("participants not notified of new game")
		}
	}
	invitationCommand(manager, users[2], 2, protocol.Message{Type: "respond-game-invite", InviteID: invite.id, Accept: true})
	requireGameError(t, users[2])
	if len(manager.Games) != 1 {
		t.Fatal("replayed acceptance created another game")
	}
}

func TestBusyInvitationRejectedImmediately(t *testing.T) {
	manager, users := testGameManager()
	users[2].name = "Busy player"
	gameCommand(manager, users[2], 2, protocol.Message{Type: "create-game"})
	for _, user := range users {
		drainGameMessages(user)
	}
	invitationCommand(manager, users[1], 1, protocol.Message{Type: "request-game", TargetID: 2})
	messages := drainGameMessages(users[1])
	if len(messages) != 1 || !strings.Contains(messages[0].Error, "Busy player is already in a game") {
		t.Fatal(messages)
	}
	if len(manager.Invitations) != 0 || users[1].gameID != 0 || len(drainGameMessages(users[2])) != 0 {
		t.Fatal("busy target received a request or requester joined a game")
	}
	for _, target := range []uint64{1, 4, 99} {
		invitationCommand(manager, users[1], 1, protocol.Message{Type: "request-game", TargetID: target})
		requireGameError(t, users[1])
	}
	invitationCommand(manager, users[2], 2, protocol.Message{Type: "request-game", TargetID: 1})
	requireGameError(t, users[2])
}

func TestInvitationDeclineAndCancellation(t *testing.T) {
	for _, action := range []string{"decline", "disconnect", "create-game", "join-game"} {
		t.Run(action, func(t *testing.T) {
			manager, users := testGameManager()
			if action == "join-game" {
				gameCommand(manager, users[3], 3, protocol.Message{Type: "create-game"})
			}
			invitationCommand(manager, users[1], 1, protocol.Message{Type: "request-game", TargetID: 2})
			invite := pendingInvitation(t, manager)
			invitationCommand(manager, users[3], 3, protocol.Message{Type: "request-game", TargetID: 2})
			requireGameError(t, users[3])
			switch action {
			case "decline":
				invitationCommand(manager, users[2], 2, protocol.Message{Type: "respond-game-invite", InviteID: invite.id})
			case "disconnect":
				manager.lock.With(func() error { manager.leave(users[1], 1); return nil })
			case "create-game":
				gameCommand(manager, users[2], 2, protocol.Message{Type: "create-game"})
			case "join-game":
				gameCommand(manager, users[2], 2, protocol.Message{Type: "join-game", GameID: users[3].gameID})
			}
			if len(manager.Invitations) != 0 {
				t.Fatal("pending request leaked")
			}
			for _, user := range []*User{users[1], users[2]} {
				closed := false
				for _, message := range drainGameMessages(user) {
					closed = closed || message.Type == "invite-closed" && message.InviteID == invite.id
				}
				if !closed {
					t.Fatal("request cancellation not delivered")
				}
			}
		})
	}
}

func TestInvitationAcceptanceRacesWithGameCreation(t *testing.T) {
	manager, users := testGameManager()
	invitationCommand(manager, users[1], 1, protocol.Message{Type: "request-game", TargetID: 2})
	invite := pendingInvitation(t, manager)
	var wait sync.WaitGroup
	wait.Go(func() {
		invitationCommand(manager, users[2], 2, protocol.Message{Type: "respond-game-invite", InviteID: invite.id, Accept: true})
	})
	wait.Go(func() { gameCommand(manager, users[2], 2, protocol.Message{Type: "create-game"}) })
	wait.Wait()
	if len(manager.Games) != 1 || len(manager.Invitations) != 0 || users[2].gameID == 0 {
		t.Fatal("racing acceptance created inconsistent games")
	}
}
