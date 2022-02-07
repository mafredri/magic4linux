package m4p

import "time"

// Protocol constants.
const (
	protocolVersion         = 1
	keepaliveTimeout        = 3 * time.Second
	clientKeepaliveInterval = 2 * time.Second
)

// Magic remote keycodes.
const (
	KeyWheelPressed = 13
	KeyChannelUp    = 33
	KeyChannelDown  = 34
	KeyLeft         = 37
	KeyUp           = 38
	KeyRight        = 39
	KeyDown         = 40
	Key0            = 48
	Key1            = 49
	Key2            = 50
	Key3            = 51
	Key4            = 52
	Key5            = 53
	Key6            = 54
	Key7            = 55
	Key8            = 56
	Key9            = 57
	KeyRed          = 403
	KeyGreen        = 404
	KeyYellow       = 405
	KeyBlue         = 406
	KeyBack         = 461
)

// DefaultFilters used for remote updates.
var DefaultFilters = []string{
	"returnValue",
	"deviceId",
	"coordinate",
	"gyroscope",
	"acceleration",
	"quaternion",
}
