package tid

import (
	"crypto/rand"
	"encoding/binary"
	"sync"
	"time"
)

// Generator makes TIDs from the current time, with one clock identifier.
//
// Each TID it makes is greater than the one before — even if two are made in the same microsecond,
// or the clock goes backwards (then the timestamp just moves on, a microsecond at a time, until the clock catches up).
//
// It is safe to use from more than one go-routine.
type Generator struct {
	mutex   sync.Mutex
	clockID uint16
	last    TID
	now     func() time.Time
}

// NewGenerator returns a generator with a random clock identifier.
func NewGenerator() *Generator {
	var buffer [2]byte
	rand.Read(buffer[:]) // never returns an error

	return &Generator{
		clockID: binary.BigEndian.Uint16(buffer[:]) & MaxClockID,
		now:     time.Now,
	}
}

// NewGeneratorWithClockID returns a generator with the given clock identifier (0..1023).
func NewGeneratorWithClockID(clockID uint16) (*Generator, error) {
	if MaxClockID < clockID {
		return nil, ErrClockIDOutOfRange
	}

	return &Generator{
		clockID: clockID,
		now:     time.Now,
	}, nil
}

// ClockID returns the generator's clock identifier.
func (receiver *Generator) ClockID() uint16 {
	if nil == receiver {
		return 0
	}
	return receiver.clockID
}

// Next returns a new TID — greater than every TID this generator returned before.
//
// It panics if the timestamp would go past [MaxTimestamp] (in the year 2255).
func (receiver *Generator) Next() TID {
	if nil == receiver {
		panic("tid: nil generator")
	}

	receiver.mutex.Lock()
	defer receiver.mutex.Unlock()

	var now func() time.Time = receiver.now
	if nil == now {
		now = time.Now
	}

	var microseconds int64 = now().UnixMicro()
	if 0 != receiver.last && microseconds <= receiver.last.Timestamp() {
		microseconds = receiver.last.Timestamp() + 1
	}

	next, err := Make(microseconds, receiver.clockID)
	if nil != err {
		panic(err)
	}

	receiver.last = next
	return next
}
