package multiplayer

import (
	"fmt"

	"github.com/kazzmir/master-of-magic/game/network-combat/protocol"
)

// Invitations and their participants are protected by manager.lock.
type GameInvitation struct {
	id                    uint64
	room                  string
	requesterID, targetID uint64
	requester, target     *User
}

func (manager *Manager) closeInvitation(invite *GameInvitation, reason string, accepted bool) {
	delete(manager.Invitations, invite.id)
	message := protocol.Message{Type: "invite-closed", InviteID: invite.id, Text: reason, Accept: accepted}
	invite.requester.send(message)
	invite.target.send(message)
}

func (manager *Manager) cancelInvitations(clientID uint64, reason string) {
	for _, invite := range manager.Invitations {
		if invite.requesterID == clientID || invite.targetID == clientID {
			manager.closeInvitation(invite, reason, false)
		}
	}
}

func (manager *Manager) hasInvitation(clientID uint64) bool {
	for _, invite := range manager.Invitations {
		if invite.requesterID == clientID || invite.targetID == clientID {
			return true
		}
	}
	return false
}

func (manager *Manager) handleInvitationCommand(user *User, clientID uint64, command protocol.Message) bool {
	switch command.Type {
	case "request-game", "respond-game-invite", "cancel-game-invite":
	default:
		return false
	}
	fail := func(reason string) {
		user.send(protocol.Message{Type: command.Type, ID: command.ID, InviteID: command.InviteID, Error: reason})
	}
	if command.Type == "request-game" {
		lobby := manager.Lobbies[user.room]
		if lobby == nil {
			fail("Join a room first")
			return true
		}
		target := lobby.Users[command.TargetID]
		switch {
		case command.TargetID == clientID:
			fail("You cannot invite yourself")
		case target == nil:
			fail("That player is no longer in your room")
		case user.gameID != 0:
			fail("Exit your current game before requesting another player")
		case target.gameID != 0:
			fail(fmt.Sprintf("%s is already in a game", target.name))
		case manager.hasInvitation(clientID):
			fail("You already have a pending game request")
		case manager.hasInvitation(command.TargetID):
			fail(fmt.Sprintf("%s already has a pending game request", target.name))
		default:
			if manager.Invitations == nil {
				manager.Invitations = make(map[uint64]*GameInvitation)
			}
			manager.nextInviteID++
			invite := &GameInvitation{id: manager.nextInviteID, room: user.room, requesterID: clientID, targetID: command.TargetID, requester: user, target: target}
			manager.Invitations[invite.id] = invite
			user.send(protocol.Message{Type: "invite-sent", ID: command.ID, InviteID: invite.id, TargetID: command.TargetID})
			target.send(protocol.Message{Type: "game-invite", InviteID: invite.id, ClientID: clientID, Name: user.name, Room: user.room})
		}
		return true
	}
	invite := manager.Invitations[command.InviteID]
	if command.Type == "cancel-game-invite" {
		if invite != nil && invite.requesterID == clientID && invite.requester == user {
			manager.closeInvitation(invite, "Game request cancelled by requester", false)
		} else if game := manager.Games[user.gameID]; game != nil && game.hostID == clientID && game.inviteID == command.InviteID && command.InviteID != 0 {
			// Acceptance may have won the race with Exit. Exit still releases
			// the requester, but must never affect a different, newer game.
			manager.leaveGame(user, clientID)
		} else {
			fail("Game request is no longer available")
		}
		return true
	}
	if invite == nil || invite.targetID != clientID || invite.target != user {
		fail("Game request is no longer available")
		return true
	}
	if !command.Accept {
		manager.closeInvitation(invite, fmt.Sprintf("%s declined the game request", user.name), false)
		return true
	}
	lobby := manager.Lobbies[invite.room]
	if lobby == nil || invite.requester.room != invite.room || user.room != invite.room || lobby.Users[invite.requesterID] != invite.requester || lobby.Users[clientID] != user {
		manager.closeInvitation(invite, "Game request cancelled: a player left the room", false)
	} else if invite.requester.gameID != 0 || user.gameID != 0 {
		manager.closeInvitation(invite, "Game request cancelled: a player is already in a game", false)
	} else {
		manager.closeInvitation(invite, "Game request accepted", true)
		game := manager.newGame(invite.requesterID, invite.requester, clientID, user)
		game.inviteID = invite.id
	}
	return true
}
