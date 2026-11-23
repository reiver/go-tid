package tid

import (
	"sync"
	"testing"
	"time"
)

func TestGenerator_monotonic(t *testing.T) {
	var clock = []time.Time{
		time.UnixMicro(1795435200000000),
		time.UnixMicro(1795435200000000), // the same microsecond
		time.UnixMicro(1795435200000000),
		time.UnixMicro(1795435100000000), // the clock goes backwards
		time.UnixMicro(1795435300000000), // and forwards again
	}
	var expectedTimestamps = []int64{
		1795435200000000,
		1795435200000001,
		1795435200000002,
		1795435200000003,
		1795435300000000,
	}

	generator, err := NewGeneratorWithClockID(7)
	if nil != err {
		t.Fatalf("Did not expect an error but actually got one: (%T) %s", err, err)
	}
	var index int
	generator.now = func() time.Time {
		defer func() { index++ }()
		return clock[index]
	}

	var previous TID
	for i, expected := range expectedTimestamps {
		next := generator.Next()
		if actual := next.Timestamp(); expected != actual {
			t.Errorf("For TID #%d, expected timestamp %d but actually got %d.", i, expected, actual)
		}
		if 7 != next.ClockID() {
			t.Errorf("For TID #%d, expected clock-id 7 but actually got %d.", i, next.ClockID())
		}
		if 0 < i && !(previous < next && previous.String() < next.String()) {
			t.Errorf("For TID #%d, expected %s > %s.", i, next, previous)
		}
		previous = next
	}
}

func TestGenerator_concurrent(t *testing.T) {
	var generator = NewGenerator()

	if MaxClockID < generator.ClockID() {
		t.Fatalf("Expected a clock-id in range but actually got %d.", generator.ClockID())
	}

	const workers = 8
	const each = 1000

	var mutex sync.Mutex
	var seen = map[TID]struct{}{}

	var wait sync.WaitGroup
	for w := 0; w < workers; w++ {
		wait.Add(1)
		go func() {
			defer wait.Done()

			var previous TID
			for i := 0; i < each; i++ {
				next := generator.Next()
				if !(previous < next) {
					t.Errorf("Expected each TID (in one go-routine) to be greater than the one before.")
				}
				previous = next

				mutex.Lock()
				seen[next] = struct{}{}
				mutex.Unlock()
			}
		}()
	}
	wait.Wait()

	if expected, actual := workers*each, len(seen); expected != actual {
		t.Errorf("Expected %d distinct TIDs but actually got %d.", expected, actual)
	}
}

func TestNewGeneratorWithClockID_outOfRange(t *testing.T) {
	if _, err := NewGeneratorWithClockID(MaxClockID + 1); ErrClockIDOutOfRange != err {
		t.Errorf("Expected ErrClockIDOutOfRange but actually got: %v", err)
	}
}

func TestGenerator_now(t *testing.T) {
	var generator = NewGenerator()

	before := time.Now().UnixMicro()
	next := generator.Next()
	after := time.Now().UnixMicro()

	if next.Timestamp() < before || after < next.Timestamp() {
		t.Errorf("Expected the TID's timestamp (%d) to be between %d and %d.", next.Timestamp(), before, after)
	}
	if !Valid(next.String()) {
		t.Errorf("Expected a valid TID but actually got %q.", next.String())
	}
}
