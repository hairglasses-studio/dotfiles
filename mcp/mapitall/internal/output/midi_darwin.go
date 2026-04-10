//go:build darwin

package output

/*
#cgo LDFLAGS: -framework CoreMIDI -framework CoreFoundation
#include <CoreMIDI/CoreMIDI.h>
#include <CoreFoundation/CoreFoundation.h>

static MIDIClientRef gClient = 0;
static MIDIPortRef gPort = 0;

int initCoreMIDI() {
    if (gClient != 0) return 0;

    CFStringRef name = CFStringCreateWithCString(NULL, "mapitall", kCFStringEncodingUTF8);
    OSStatus err = MIDIClientCreate(name, NULL, NULL, &gClient);
    CFRelease(name);
    if (err != noErr) return (int)err;

    CFStringRef portName = CFStringCreateWithCString(NULL, "mapitall-out", kCFStringEncodingUTF8);
    err = MIDIOutputPortCreate(gClient, portName, &gPort);
    CFRelease(portName);
    return (int)err;
}

int sendMIDIBytes(int dest, unsigned char b0, unsigned char b1, unsigned char b2) {
    if (gPort == 0) return -1;

    ItemCount numDests = MIDIGetNumberOfDestinations();
    if (dest < 0 || (ItemCount)dest >= numDests) return -2;

    MIDIEndpointRef endpoint = MIDIGetDestination(dest);

    // Build a single 3-byte MIDI packet.
    Byte buffer[128];
    MIDIPacketList *pktList = (MIDIPacketList *)buffer;
    MIDIPacket *pkt = MIDIPacketListInit(pktList);
    Byte data[3] = {b0, b1, b2};
    pkt = MIDIPacketListAdd(pktList, sizeof(buffer), pkt, 0, 3, data);
    if (pkt == NULL) return -3;

    return (int)MIDISend(gPort, endpoint, pktList);
}

int getMIDIDestCount() {
    return (int)MIDIGetNumberOfDestinations();
}
*/
import "C"

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/hairglasses-studio/mapitall/internal/mapping"
)

// MIDITarget sends MIDI messages via macOS CoreMIDI.
type MIDITarget struct {
	once sync.Once
	err  error
}

func NewMIDITarget() *MIDITarget { return &MIDITarget{} }

func (t *MIDITarget) Type() mapping.OutputType { return mapping.OutputMIDI }

func (t *MIDITarget) init() error {
	t.once.Do(func() {
		if ret := C.initCoreMIDI(); ret != 0 {
			t.err = fmt.Errorf("CoreMIDI init failed: %d", ret)
		}
	})
	return t.err
}

func (t *MIDITarget) Execute(action mapping.OutputAction, value float64) error {
	if err := t.init(); err != nil {
		return err
	}

	channel := action.Port // reuse Port field as MIDI channel
	if channel < 0 || channel > 15 {
		channel = 0
	}
	dest := 0 // default to first destination

	switch {
	case action.Address != "":
		// CC message: address = controller number
		cc, err := strconv.Atoi(action.Address)
		if err != nil {
			return fmt.Errorf("invalid CC number: %s", action.Address)
		}
		ccVal := int(value * 127)
		if ccVal > 127 {
			ccVal = 127
		}
		return t.sendCC(dest, channel, cc, ccVal)

	case len(action.Keys) > 0:
		// Note message: Keys[0] = note number
		note, err := strconv.Atoi(action.Keys[0])
		if err != nil {
			return fmt.Errorf("invalid note number: %s", action.Keys[0])
		}
		if value > 0.5 {
			vel := int(value * 127)
			if vel > 127 {
				vel = 127
			}
			return t.sendNoteOn(dest, channel, note, vel)
		}
		return t.sendNoteOff(dest, channel, note)
	}

	return nil
}

func (t *MIDITarget) sendCC(dest, channel, controller, value int) error {
	ret := C.sendMIDIBytes(C.int(dest),
		C.uchar(0xB0|byte(channel)),
		C.uchar(controller),
		C.uchar(value))
	if ret != 0 {
		return fmt.Errorf("MIDI send CC failed: %d", ret)
	}
	return nil
}

func (t *MIDITarget) sendNoteOn(dest, channel, note, velocity int) error {
	ret := C.sendMIDIBytes(C.int(dest),
		C.uchar(0x90|byte(channel)),
		C.uchar(note),
		C.uchar(velocity))
	if ret != 0 {
		return fmt.Errorf("MIDI send NoteOn failed: %d", ret)
	}
	return nil
}

func (t *MIDITarget) sendNoteOff(dest, channel, note int) error {
	ret := C.sendMIDIBytes(C.int(dest),
		C.uchar(0x80|byte(channel)),
		C.uchar(note),
		C.uchar(0))
	if ret != 0 {
		return fmt.Errorf("MIDI send NoteOff failed: %d", ret)
	}
	return nil
}

func (t *MIDITarget) Close() error { return nil }
