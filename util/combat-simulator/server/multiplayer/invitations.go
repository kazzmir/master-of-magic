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

func (manager *Manager) closeInvitation(invite *GameInvitation, reason string) {
	delete(manager.Invitations, invite.id)
	message := protocol.Message{Type: "invite-closed", InviteID: invite.id, Text: reason}
	invite.requester.send(message)
	invite.target.send(message)
}

func (manager *Manager) cancelInvitations(clientID uint64, reason string) {
	for _, invite := range manager.Invitations {
		if invite.requesterID == clientID || invite.targetID == clientID {
			manager.closeInvitation(invite, reason)
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
	case "request-game", "respond-game-invite":
	default:
		return false
	}
	fail := func(reason string) {
		user.send(protocol.Message{ID: command.ID, InviteID: command.InviteID, Error: reason})
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
	if invite == nil || invite.targetID != clientID || invite.target != user {
		fail("Game request is no longer available")
		return true
	}
	if !command.Accept {
		manager.closeInvitation(invite, fmt.Sprintf("%s declined the game request", user.name))
		return true
	}
	lobby := manager.Lobbies[invite.room]
	if lobby == nil || invite.requester.room != invite.room || user.room != invite.room || lobby.Users[invite.requesterID] != invite.requester || lobby.Users[clientID] != user {
		manager.closeInvitation(invite, "Game request cancelled: a player left the room")
	} else if invite.requester.gameID != 0 || user.gameID != 0 {
		manager.closeInvitation(invite, "Game request cancelled: a player is already in a game")
	} else {
		manager.closeInvitation(invite, "Game request accepted")
		manager.newGame(invite.requesterID, invite.requester, clientID, user)
	}
	return true
}
