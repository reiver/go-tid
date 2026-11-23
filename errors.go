package tid

import (
	"codeberg.org/reiver/go-erorr"
)

const (
	ErrBadCharacter         = erorr.Error("tid: bad character — must be one of 234567abcdefghijklmnopqrstuvwxyz")
	ErrBadLength            = erorr.Error("tid: bad length — must be 13 characters")
	ErrClockIDOutOfRange    = erorr.Error("tid: clock identifier out of range — must be 0..1023")
	ErrNilReceiver          = erorr.Error("tid: nil receiver")
	ErrTimestampOutOfRange  = erorr.Error("tid: timestamp out of range — must be 0..2⁵³-1 microseconds since the UNIX epoch")
	ErrTooBig               = erorr.Error("tid: too big — the first character must be one of 234567abcdefghij")
)
