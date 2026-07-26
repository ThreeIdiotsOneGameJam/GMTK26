package net

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/google/uuid"
	"github.com/threeidiotsonegamejam/gmtk26/src/game"
	"github.com/threeidiotsonegamejam/gmtk26/src/net/packets"
	"github.com/threeidiotsonegamejam/gmtk26/src/util"
)

func TestLocalGameUsesSerializedServerFlow(t *testing.T) {
	originalPlayer := *game.PlayerData
	*game.PlayerData = game.Player{
		ClientID:   game.ClientID(uuid.NewString()),
		PlayerName: "local-test",
		Color:      util.RGB{10, 20, 30},
	}
	t.Cleanup(func() {
		StopLocalGame()
		*game.PlayerData = originalPlayer
	})

	const seed int64 = 8675309
	if err := StartLocalGame(seed); err != nil {
		t.Fatalf("StartLocalGame() error = %v", err)
	}
	if !LocalGameActive() {
		t.Fatal("local game is not active after startup")
	}
	if !client.suspended.Load() {
		t.Fatal("remote client was not suspended for local play")
	}

	var joined *packets.S2CGameJoinedPacket
	var started *packets.S2CGameStartPacket
	var order []packets.PacketType
	deadline := time.Now().Add(5 * time.Second)
	for started == nil && time.Now().Before(deadline) {
		DrainEvents(func(packet packets.S2CPacket) {
			order = append(order, packet.PacketType())
			switch packet := packet.(type) {
			case *packets.S2CConnectAcceptedPacket:
				t.Error("local connect-accepted packet leaked to the UI")
			case *packets.S2CGameJoinedPacket:
				joined = packet
			case *packets.S2CGameStartPacket:
				started = packet
			}
		})
		if started == nil {
			time.Sleep(10 * time.Millisecond)
		}
	}

	if joined == nil {
		t.Fatal("did not receive local game joined packet")
	}
	if started == nil {
		t.Fatal("did not receive local game start packet")
	}
	joinedIndex := -1
	startedIndex := -1
	for i, packetType := range order {
		switch packetType {
		case packets.S2CGameJoinedPacketType:
			joinedIndex = i
		case packets.S2CGameStartPacketType:
			startedIndex = i
		}
	}
	if joinedIndex < 0 || startedIndex <= joinedIndex {
		t.Fatalf("local startup packet order = %v, want joined before start", order)
	}

	state := joined.Game
	if state.Multiplayer {
		t.Fatal("local game was marked multiplayer")
	}
	if state.Public {
		t.Fatal("local game was marked public")
	}
	if state.MaxPlayers != 1 {
		t.Fatalf("local MaxPlayers = %d, want 1", state.MaxPlayers)
	}
	if state.Map.Seed != seed || started.Map.Seed != seed {
		t.Fatalf(
			"local seeds = joined %d, started %d; want %d",
			state.Map.Seed,
			started.Map.Seed,
			seed,
		)
	}
	if state.Factions[0].Player == nil ||
		state.Factions[0].Player.PlayerName != "local-test" {
		t.Fatalf("local player faction = %#v", state.Factions[0])
	}
	for i := 1; i < len(state.Factions); i++ {
		if !state.Factions[i].AI {
			t.Fatalf("faction %d is not AI", i)
		}
	}

	StopLocalGame()
	if LocalGameActive() {
		t.Fatal("local game remains active after stop")
	}
	if client.suspended.Load() {
		t.Fatal("remote client remains suspended after local stop")
	}
}

func TestStartingSecondLocalGameIsRejected(t *testing.T) {
	originalPlayer := *game.PlayerData
	*game.PlayerData = game.Player{
		ClientID:   game.ClientID(uuid.NewString()),
		PlayerName: "local-test",
		Color:      util.RGB{10, 20, 30},
	}
	t.Cleanup(func() {
		StopLocalGame()
		*game.PlayerData = originalPlayer
	})

	if err := StartLocalGame(1); err != nil {
		t.Fatalf("first StartLocalGame() error = %v", err)
	}
	if err := StartLocalGame(2); err != ErrLocalGameActive {
		t.Fatalf("second StartLocalGame() error = %v, want %v", err, ErrLocalGameActive)
	}
}

func TestRemoteClientDisconnectsWhileSuspendedAndReconnects(t *testing.T) {
	originalPlayer := *game.PlayerData
	*game.PlayerData = game.Player{
		ClientID:   game.ClientID(uuid.NewString()),
		PlayerName: "remote-test",
		Color:      util.RGB{10, 20, 30},
	}
	t.Cleanup(func() {
		*game.PlayerData = originalPlayer
	})

	var acceptedConnections atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		socket, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		defer socket.CloseNow()
		acceptedConnections.Add(1)

		_, message, err := socket.Read(context.Background())
		if err != nil {
			return
		}
		packet, err := packets.Deserialize(message)
		if err != nil {
			t.Errorf("deserialize remote connect packet: %v", err)
			return
		}
		if _, ok := packet.(*packets.C2SConnectPacket); !ok {
			t.Errorf("first remote packet = %T, want connect", packet)
			return
		}

		response, err := packets.Serialize(&packets.S2CConnectAcceptedPacket{
			ClientID:   game.PlayerData.ClientID,
			Persistent: true,
		})
		if err != nil {
			t.Errorf("serialize remote connect response: %v", err)
			return
		}
		if err := socket.Write(context.Background(), websocket.MessageText, response); err != nil {
			return
		}

		for {
			if _, _, err := socket.Read(context.Background()); err != nil {
				return
			}
		}
	}))
	t.Cleanup(server.Close)

	remote := newWSClient()
	runDone := make(chan struct{})
	go func() {
		remote.run(strings.TrimPrefix(server.URL, "http://"))
		close(runDone)
	}()
	t.Cleanup(func() {
		remote.close()
		<-runDone
	})

	waitForTest(t, 3*time.Second, func() bool {
		return remote.state() == ConnectionConnected &&
			acceptedConnections.Load() == 1
	}, "initial remote connection")

	remote.suspend()
	waitForTest(t, 3*time.Second, func() bool {
		return remote.state() == ConnectionDisconnected
	}, "remote disconnect after suspension")

	remote.resume()
	waitForTest(t, 3*time.Second, func() bool {
		return remote.state() == ConnectionConnected &&
			acceptedConnections.Load() >= 2
	}, "remote reconnect after resume")
}

func waitForTest(t *testing.T, timeout time.Duration, condition func() bool, description string) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", description)
}
