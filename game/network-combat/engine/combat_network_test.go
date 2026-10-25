package engine

import (
	"io"
	"testing"
	"time"

	"github.com/kazzmir/master-of-magic/game/magic/combat"
	"github.com/kazzmir/master-of-magic/game/magic/data"
	"github.com/kazzmir/master-of-magic/game/magic/player"
	"github.com/kazzmir/master-of-magic/game/magic/setup"
	"github.com/kazzmir/master-of-magic/game/magic/units"
)

type armyPackets struct {
	packets  [][]byte
	deadline time.Time
}

func (peer *armyPackets) Read() ([]byte, error) {
	if len(peer.packets) == 0 {
		return nil, io.EOF
	}
	data := peer.packets[0]
	peer.packets = peer.packets[1:]
	return data, nil
}
func (peer *armyPackets) Send(data []byte) error {
	peer.packets = append(peer.packets, append([]byte(nil), data...))
	return nil
}
func (peer *armyPackets) SetReadDeadline(deadline time.Time) error {
	peer.deadline = deadline
	return nil
}

func TestNetworkArmyExchange(t *testing.T) {
	for _, config := range []struct {
		team combat.Team
		unit units.Unit
	}{
		{combat.TeamAttacker, units.OrcSpearmen}, {combat.TeamDefender, units.LizardSpearmen},
	} {
		owner := player.MakePlayer(setup.WizardCustom{Name: "Player", Banner: data.BannerGreen}, true, 0, 0, nil, &noGlobalEnchantments{})
		peer := &armyPackets{}
		remote := combat.MakeRemote(false, false, peer)
		if err := sendArmy(remote, config.team, createArmyN(owner, config.unit, 2)); err != nil {
			t.Fatal(err)
		}
		peer.packets = append(peer.packets, []byte("first combat event"))
		army, err := receiveArmy(remote, owner)
		if err != nil {
			t.Fatal(err)
		}
		if len(army.GetUnits()) != 2 {
			t.Fatal("wrong army size")
		}
		for _, unit := range army.GetUnits() {
			if unit.Unit.GetRawUnit().GetId() != config.unit.GetId() {
				t.Fatal("wrong spearmen type")
			}
		}
		if len(peer.packets) != 1 || string(peer.packets[0]) != "first combat event" || !peer.deadline.IsZero() {
			t.Fatal("army exchange consumed combat event or left a deadline")
		}
	}
}
