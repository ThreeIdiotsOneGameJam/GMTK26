package packets

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync"
)

type PacketType int

const (
	UnknownPacketType PacketType = iota
	C2SPingPacketType
	S2CPongPacketType
)

type Packet interface {
	PacketType() PacketType
	Handle()
}

type S2CPacket interface {
	Packet
	isS2C()
}

type C2SPacket interface {
	Packet
	isC2S()
}

type packetRegistration struct {
	newPacket func() Packet
	matches   func(Packet) bool
}

var packetRegistry = struct {
	sync.RWMutex
	registrations map[PacketType]packetRegistration
}{
	registrations: make(map[PacketType]packetRegistration),
}

// RegisterPacket associates a packet type ID with a constructor and a concrete
// type matcher. Registration normally happens from init functions before the
// package is used.
func RegisterPacket(
	packetType PacketType,
	newPacket func() Packet,
	matches func(Packet) bool,
) error {
	if packetType == UnknownPacketType {
		return fmt.Errorf("register packet: cannot register unknown packet type")
	}
	if newPacket == nil {
		return fmt.Errorf("register packet type %d: nil constructor", packetType)
	}
	if matches == nil {
		return fmt.Errorf("register packet type %d: nil matcher", packetType)
	}

	sample := newPacket()
	if sample == nil || !matches(sample) {
		return fmt.Errorf(
			"register packet type %d: constructor and matcher disagree",
			packetType,
		)
	}
	if actualType := sample.PacketType(); actualType != packetType {
		return fmt.Errorf(
			"register packet type %d: constructor returned packet type %d",
			packetType,
			actualType,
		)
	}

	packetRegistry.Lock()
	defer packetRegistry.Unlock()

	if _, exists := packetRegistry.registrations[packetType]; exists {
		return fmt.Errorf(
			"register packet type %d: already registered",
			packetType,
		)
	}

	packetRegistry.registrations[packetType] = packetRegistration{
		newPacket: newPacket,
		matches:   matches,
	}

	return nil
}

func mustRegisterPacket(
	packetType PacketType,
	newPacket func() Packet,
	matches func(Packet) bool,
) {
	if err := RegisterPacket(packetType, newPacket, matches); err != nil {
		panic(err)
	}
}

func lookupPacket(packetType PacketType) (packetRegistration, bool) {
	packetRegistry.RLock()
	defer packetRegistry.RUnlock()

	registration, ok := packetRegistry.registrations[packetType]
	return registration, ok
}

type serializedPacket struct {
	Type PacketType      `json:"type"`
	Data json.RawMessage `json:"data"`
}

func Serialize(packet Packet) ([]byte, error) {
	if packet == nil {
		return nil, fmt.Errorf("serialize packet: nil packet")
	}

	packetType := packet.PacketType()
	registration, ok := lookupPacket(packetType)
	if !ok {
		return nil, fmt.Errorf(
			"serialize packet: unregistered packet type %d",
			packetType,
		)
	}

	if !registration.matches(packet) {
		return nil, fmt.Errorf(
			"serialize packet: type %T reports incorrect packet type %d",
			packet,
			packetType,
		)
	}

	payload, err := json.Marshal(packet)
	if err != nil {
		return nil, fmt.Errorf(
			"serialize packet payload %T: %w",
			packet,
			err,
		)
	}

	data, err := json.Marshal(serializedPacket{
		Type: packetType,
		Data: payload,
	})
	if err != nil {
		return nil, fmt.Errorf("serialize packet envelope: %w", err)
	}

	return data, nil
}

func Deserialize(data []byte) (Packet, error) {
	var envelope serializedPacket
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, fmt.Errorf("deserialize packet envelope: %w", err)
	}

	registration, ok := lookupPacket(envelope.Type)
	if !ok {
		return nil, fmt.Errorf(
			"deserialize packet: unknown packet type %d",
			envelope.Type,
		)
	}

	payload := bytes.TrimSpace(envelope.Data)
	if len(payload) == 0 || bytes.Equal(payload, []byte("null")) {
		return nil, fmt.Errorf(
			"deserialize packet type %d: missing or null data",
			envelope.Type,
		)
	}

	packet := registration.newPacket()
	if packet == nil || !registration.matches(packet) {
		return nil, fmt.Errorf(
			"deserialize packet type %d: invalid constructor result",
			envelope.Type,
		)
	}

	if err := json.Unmarshal(payload, packet); err != nil {
		return nil, fmt.Errorf(
			"deserialize packet type %d: %w",
			envelope.Type,
			err,
		)
	}

	if actualType := packet.PacketType(); actualType != envelope.Type {
		return nil, fmt.Errorf(
			"deserialize packet: constructor returned packet type %d for envelope type %d",
			actualType,
			envelope.Type,
		)
	}

	return packet, nil
}
