package m4p

import (
	"bytes"
	"encoding/json"
)

type MessageType string

// MessageType enums.
const (
	Magic4PCAdMessage   MessageType = "magic4pc_ad"
	SubSensorMessage    MessageType = "sub_sensor"
	RemoteUpdateMessage MessageType = "remote_update"
	InputMessage        MessageType = "input"
	MouseMessage        MessageType = "mouse"
	WheelMessage        MessageType = "wheel"
	KeepAliveMessage    MessageType = "keepalive"
)

// Message format sent over the wire.
type Message struct {
	Type    MessageType `json:"t"`
	Version int         `json:"version"`
	*DeviceInfo
	*Register
	*RemoteUpdate
	*Input
	Mouse Mouse `json:"mouse"`
	Wheel Wheel `json:"wheel"`
}

// NewMessage initializes a message with the type and protocol version.
func NewMessage(typ MessageType) Message {
	return Message{
		Type:    typ,
		Version: protocolVersion,
	}
}

// DeviceInfo represents a magic4pc server.
type DeviceInfo struct {
	Model  string `json:"model"`
	IPAddr string `json:"-"`
	Port   int    `json:"port"`
	MAC    string `json:"mac"`
}

// Register payload for registering a new client on the server.
type Register struct {
	UpdateFrequency int      `json:"updateFreq"`
	Filter          []string `json:"filter"`
}

// Input event (key pressed).
type Input struct {
	Parameters struct {
		KeyCode int  `json:"keyCode"`
		IsDown  bool `json:"isDown"`
	} `json:"parameters"`
}

// RemoteUpdate event, sensor data from the magic remote.
type RemoteUpdate struct {
	// filters []string
	Payload []byte `json:"payload"`
}

type Coordinates struct {
	X int32 `json:"x"`
	Y int32 `json:"y"`
}

type Mouse struct {
	Type string `json:"type"` // mousedown, mouseup
	Coordinates
}

type Wheel struct {
	Delta int32 `json:"delta"`
	Coordinates
}

// type (
// 	ReturnValue  uint8
// 	DeviceID     uint8
// 	Gyroscope    struct{ X, Y, Z float32 }
// 	Acceleration struct{ X, Y, Z float32 }
// 	Quaternion   struct{ Q0, Q1, Q2, Q3 float32 }
// )

// func (ru RemoteUpdate) Coordinates() Coordinates {
// 	return Coordinates{}
// }

// func (ru RemoteUpdate) Acceleration() Acceleration {
// 	return Acceleration{}
// }

func decode(b []byte) (Message, error) {
	var m Message
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&m); err != nil {
		return Message{}, err
	}

	return m, nil
}
