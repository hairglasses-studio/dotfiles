//go:build darwin

package output

// DarwinKeyCodes maps Linux KEY_* names (used in TOML profiles) to macOS
// virtual key codes (CGKeyCode). This allows profiles to use the same
// key names across platforms.
//
// macOS key codes are from Events.h (Carbon framework).
var DarwinKeyCodes = map[string]uint16{
	// Letters
	"KEY_A": 0x00, "KEY_S": 0x01, "KEY_D": 0x02, "KEY_F": 0x03,
	"KEY_H": 0x04, "KEY_G": 0x05, "KEY_Z": 0x06, "KEY_X": 0x07,
	"KEY_C": 0x08, "KEY_V": 0x09, "KEY_B": 0x0B, "KEY_Q": 0x0C,
	"KEY_W": 0x0D, "KEY_E": 0x0E, "KEY_R": 0x0F, "KEY_Y": 0x10,
	"KEY_T": 0x11, "KEY_O": 0x1F, "KEY_U": 0x20, "KEY_I": 0x22,
	"KEY_P": 0x23, "KEY_L": 0x25, "KEY_J": 0x26, "KEY_K": 0x28,
	"KEY_N": 0x2D, "KEY_M": 0x2E,

	// Numbers
	"KEY_1": 0x12, "KEY_2": 0x13, "KEY_3": 0x14, "KEY_4": 0x15,
	"KEY_6": 0x16, "KEY_5": 0x17, "KEY_9": 0x19, "KEY_7": 0x1A,
	"KEY_8": 0x1C, "KEY_0": 0x1D,

	// Special keys
	"KEY_ENTER":      0x24, "KEY_TAB": 0x30, "KEY_SPACE": 0x31,
	"KEY_BACKSPACE":  0x33, "KEY_ESC": 0x35, "KEY_DELETE": 0x75,
	"KEY_HOME":       0x73, "KEY_END": 0x77,
	"KEY_PAGEUP":     0x74, "KEY_PAGEDOWN": 0x79,
	"KEY_UP":         0x7E, "KEY_DOWN": 0x7D,
	"KEY_LEFT":       0x7B, "KEY_RIGHT": 0x7C,

	// Modifiers
	"KEY_LEFTSHIFT":  0x38, "KEY_RIGHTSHIFT": 0x3C,
	"KEY_LEFTCTRL":   0x3B, "KEY_RIGHTCTRL": 0x3E,
	"KEY_LEFTALT":    0x3A, "KEY_RIGHTALT": 0x3D,
	"KEY_LEFTMETA":   0x37, "KEY_RIGHTMETA": 0x36, // Command keys

	// Function keys
	"KEY_F1": 0x7A, "KEY_F2": 0x78, "KEY_F3": 0x63, "KEY_F4": 0x76,
	"KEY_F5": 0x60, "KEY_F6": 0x61, "KEY_F7": 0x62, "KEY_F8": 0x64,
	"KEY_F9": 0x65, "KEY_F10": 0x6D, "KEY_F11": 0x67, "KEY_F12": 0x6F,

	// Punctuation
	"KEY_MINUS":      0x1B, "KEY_EQUAL": 0x18,
	"KEY_LEFTBRACE":  0x21, "KEY_RIGHTBRACE": 0x1E,
	"KEY_BACKSLASH":  0x2A, "KEY_SEMICOLON": 0x29,
	"KEY_APOSTROPHE": 0x27, "KEY_GRAVE": 0x32,
	"KEY_COMMA":      0x2B, "KEY_DOT": 0x2F, "KEY_SLASH": 0x2C,

	// Keypad
	"KEY_KPASTERISK": 0x43, "KEY_KPPLUS": 0x45, "KEY_KPMINUS": 0x4E,
	"KEY_KPDOT":      0x41, "KEY_KPSLASH": 0x4B, "KEY_KPENTER": 0x4C,
	"KEY_KP0": 0x52, "KEY_KP1": 0x53, "KEY_KP2": 0x54, "KEY_KP3": 0x55,
	"KEY_KP4": 0x56, "KEY_KP5": 0x57, "KEY_KP6": 0x58,
	"KEY_KP7": 0x59, "KEY_KP8": 0x5B, "KEY_KP9": 0x5C,

	// Media
	"KEY_VOLUMEUP":   0x48, "KEY_VOLUMEDOWN": 0x49, "KEY_MUTE": 0x4A,

	// Capslock
	"KEY_CAPSLOCK": 0x39,
}
