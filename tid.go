package tid

import (
	"time"

	"codeberg.org/reiver/go-erorr"
	"codeberg.org/reiver/go-field"
)

const (
	// Length is the length, in characters, of a TID.
	Length = 13

	// MaxClockID is the largest clock identifier.
	MaxClockID = 1<<10 - 1 // 1023

	// MaxTimestamp is the largest timestamp (in microseconds since the UNIX epoch) a TID made by this package can hold.
	MaxTimestamp = 1<<53 - 1
)

const alphabet string = "234567abcdefghijklmnopqrstuvwxyz"

// TID is a timestamp identifier. See the package documentation.
//
// TIDs compare (with <, ==, >) the same as their strings do.
type TID uint64

// Parse parses a TID from its 13-character string form.
//
// Example usage:
//
//	t, err := tid.Parse("3jzfcijpj2z2a")
func Parse(s string) (TID, error) {
	if Length != len(s) {
		return 0, erorr.Wrap(ErrBadLength, "bad TID length", field.Int("length", len(s)))
	}

	var value uint64
	for index := 0; index < Length; index++ {
		var digit int = decode(s[index])
		if digit < 0 {
			return 0, erorr.Wrap(ErrBadCharacter, "bad TID character", field.Int("index", index), field.String("tid", s))
		}
		// 13 characters × 5 bits = 65 bits, so the first character may only hold 4 bits.
		if 0 == index && 16 <= digit {
			return 0, erorr.Wrap(ErrTooBig, "TID too big", field.String("tid", s))
		}
		value = value<<5 | uint64(digit)
	}

	return TID(value), nil
}

// Valid returns whether 's' is a syntactically valid TID.
func Valid(s string) bool {
	_, err := Parse(s)
	return nil == err
}

func decode(c byte) int {
	switch {
	case '2' <= c && c <= '7':
		return int(c - '2')
	case 'a' <= c && c <= 'z':
		return int(c-'a') + 6
	default:
		return -1
	}
}

// Make makes a TID from a timestamp (microseconds since the UNIX epoch) and a clock identifier.
func Make(microseconds int64, clockID uint16) (TID, error) {
	if microseconds < 0 || MaxTimestamp < microseconds {
		return 0, erorr.Wrap(ErrTimestampOutOfRange, "TID timestamp out of range", field.Int64("microseconds", microseconds))
	}
	if MaxClockID < clockID {
		return 0, erorr.Wrap(ErrClockIDOutOfRange, "TID clock identifier out of range", field.Uint16("clock-id", clockID))
	}

	return TID(uint64(microseconds)<<10 | uint64(clockID)), nil
}

// FromTime makes a TID from a time (to the microsecond) and a clock identifier.
func FromTime(t time.Time, clockID uint16) (TID, error) {
	return Make(t.UnixMicro(), clockID)
}

// String returns the TID's 13-character string form.
func (receiver TID) String() string {
	var buffer [Length]byte

	var value uint64 = uint64(receiver)
	for index := Length - 1; 0 <= index; index-- {
		buffer[index] = alphabet[value&31]
		value >>= 5
	}

	return string(buffer[:])
}

// Timestamp returns the TID's timestamp — microseconds since the UNIX epoch.
//
// (For a TID with its top bit set — which the syntax allows, but no well-behaved generator makes — the top bit is included.)
func (receiver TID) Timestamp() int64 {
	return int64(uint64(receiver) >> 10)
}

// Time returns the TID's timestamp as a time (in UTC).
func (receiver TID) Time() time.Time {
	return time.UnixMicro(receiver.Timestamp()).UTC()
}

// ClockID returns the TID's clock identifier (0..1023).
func (receiver TID) ClockID() uint16 {
	return uint16(uint64(receiver) & MaxClockID)
}

// MarshalText returns the TID's string form (ex: for JSON).
func (receiver TID) MarshalText() ([]byte, error) {
	return []byte(receiver.String()), nil
}

// UnmarshalText parses a TID from its string form (ex: from JSON).
func (receiver *TID) UnmarshalText(text []byte) error {
	if nil == receiver {
		return ErrNilReceiver
	}

	value, err := Parse(string(text))
	if nil != err {
		return err
	}

	*receiver = value
	return nil
}
