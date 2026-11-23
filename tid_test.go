package tid_test

import (
	"bufio"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"codeberg.org/reiver/go-erorr"

	"github.com/reiver/go-tid"
)

// interopLines returns the test lines of an ATProto interop test file (skipping blank lines and comments).
func interopLines(t *testing.T, filename string) []string {
	t.Helper()

	file, err := os.Open(filename)
	if nil != err {
		t.Fatalf("Did not expect an error but actually got one: (%T) %s", err, err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if "" == strings.TrimSpace(line) || strings.HasPrefix(line, "#") {
			continue
		}
		lines = append(lines, line)
	}
	if err := scanner.Err(); nil != err {
		t.Fatalf("Did not expect an error but actually got one: (%T) %s", err, err)
	}
	if 0 == len(lines) {
		t.Fatalf("Expected some test lines in %q.", filename)
	}
	return lines
}

func TestParse_interopValid(t *testing.T) {
	for _, line := range interopLines(t, "testdata/tid_syntax_valid.txt") {
		value, err := tid.Parse(line)
		if nil != err {
			t.Errorf("For %q, did not expect an error but actually got one: (%T) %s", line, err, err)
			continue
		}
		if actual := value.String(); line != actual {
			t.Errorf("For %q, expected it to round-trip but actually got %q.", line, actual)
		}
	}
}

func TestParse_interopInvalid(t *testing.T) {
	for _, line := range interopLines(t, "testdata/tid_syntax_invalid.txt") {
		if tid.Valid(line) {
			t.Errorf("Expected %q to be invalid, but it was accepted.", line)
		}
	}
}

func TestParse_errors(t *testing.T) {
	tests := []struct {
		Value    string
		Expected error
	}{
		{Value: "", Expected: tid.ErrBadLength},
		{Value: "222", Expected: tid.ErrBadLength},
		{Value: "3jzfcijpj2z2aa", Expected: tid.ErrBadLength},
		{Value: "3jzf-cij-pj2z-2a", Expected: tid.ErrBadLength},
		{Value: "3jzfcijpj2z21", Expected: tid.ErrBadCharacter},
		{Value: "0000000000000", Expected: tid.ErrBadCharacter},
		{Value: "3JZFCIJPJ2Z2A", Expected: tid.ErrBadCharacter},
		{Value: "3jzfcijpj2z2\x00", Expected: tid.ErrBadCharacter},
		{Value: "kjzfcijpj2z2a", Expected: tid.ErrTooBig},
		{Value: "zzzzzzzzzzzzz", Expected: tid.ErrTooBig},
	}

	for testNumber, test := range tests {
		_, err := tid.Parse(test.Value)
		if !erorr.Is(err, test.Expected) {
			t.Errorf("For test #%d (%q), expected error %q but actually got: %v", testNumber, test.Value, test.Expected, err)
		}
	}
}

// Expected values computed independently (with arbitrary-precision integers).
func TestParse_values(t *testing.T) {
	tests := []struct {
		Value             string
		ExpectedInteger   uint64
		ExpectedTimestamp int64
		ExpectedClockID   uint16
	}{
		{Value: "2222222222222", ExpectedInteger: 0, ExpectedTimestamp: 0, ExpectedClockID: 0},
		{Value: "3jzfcijpj2z2a", ExpectedInteger: 1728652679052295174, ExpectedTimestamp: 1688137381887007, ExpectedClockID: 6},
		{Value: "7777777777777", ExpectedInteger: 5950562604422436005, ExpectedTimestamp: 5811096293381285, ExpectedClockID: 165},
		{Value: "3zzzzzzzzzzzz", ExpectedInteger: 2305843009213693951, ExpectedTimestamp: 2251799813685247, ExpectedClockID: 1023},
		{Value: "bzzzzzzzzzzzz", ExpectedInteger: 9223372036854775807, ExpectedTimestamp: tid.MaxTimestamp, ExpectedClockID: 1023},
		// The top bit set: allowed by the syntax.
		{Value: "jzzzzzzzzzzzz", ExpectedInteger: 18446744073709551615, ExpectedTimestamp: 18014398509481983, ExpectedClockID: 1023},
	}

	for testNumber, test := range tests {
		value, err := tid.Parse(test.Value)
		if nil != err {
			t.Errorf("For test #%d (%q), did not expect an error but actually got one: (%T) %s", testNumber, test.Value, err, err)
			continue
		}

		if expected, actual := test.ExpectedInteger, uint64(value); expected != actual {
			t.Errorf("For test #%d (%q), expected integer %d but actually got %d.", testNumber, test.Value, expected, actual)
		}
		if expected, actual := test.ExpectedTimestamp, value.Timestamp(); expected != actual {
			t.Errorf("For test #%d (%q), expected timestamp %d but actually got %d.", testNumber, test.Value, expected, actual)
		}
		if expected, actual := test.ExpectedClockID, value.ClockID(); expected != actual {
			t.Errorf("For test #%d (%q), expected clock-id %d but actually got %d.", testNumber, test.Value, expected, actual)
		}
		if actual := value.String(); test.Value != actual {
			t.Errorf("For test #%d, expected it to round-trip to %q but actually got %q.", testNumber, test.Value, actual)
		}
	}
}

func TestTID_time(t *testing.T) {
	value, err := tid.Parse("3jzfcijpj2z2a")
	if nil != err {
		t.Fatalf("Did not expect an error but actually got one: (%T) %s", err, err)
	}

	expected := time.Date(2023, time.June, 30, 15, 3, 1, 887007000, time.UTC)
	if actual := value.Time(); !expected.Equal(actual) || time.UTC != actual.Location() {
		t.Errorf("Expected time %s but actually got %s.", expected, actual)
	}
}

func TestMake(t *testing.T) {
	tests := []struct {
		Microseconds int64
		ClockID      uint16
		Expected     string
	}{
		{Microseconds: 0, ClockID: 0, Expected: "2222222222222"},
		{Microseconds: 1700000000000000, ClockID: 5, Expected: "3ke6kg3wk2227"},
		{Microseconds: 1795435200123456, ClockID: 1023, Expected: "3n2y3ep66m2zz"},
		{Microseconds: tid.MaxTimestamp, ClockID: tid.MaxClockID, Expected: "bzzzzzzzzzzzz"},
	}

	for testNumber, test := range tests {
		value, err := tid.Make(test.Microseconds, test.ClockID)
		if nil != err {
			t.Errorf("For test #%d, did not expect an error but actually got one: (%T) %s", testNumber, err, err)
			continue
		}
		if actual := value.String(); test.Expected != actual {
			t.Errorf("For test #%d, expected %q but actually got %q.", testNumber, test.Expected, actual)
		}
		if test.Microseconds != value.Timestamp() || test.ClockID != value.ClockID() {
			t.Errorf("For test #%d, expected the parts to round-trip; got timestamp %d, clock-id %d.", testNumber, value.Timestamp(), value.ClockID())
		}
	}
}

func TestMake_errors(t *testing.T) {
	tests := []struct {
		Microseconds int64
		ClockID      uint16
		Expected     error
	}{
		{Microseconds: -1, ClockID: 0, Expected: tid.ErrTimestampOutOfRange},
		{Microseconds: tid.MaxTimestamp + 1, ClockID: 0, Expected: tid.ErrTimestampOutOfRange},
		{Microseconds: 0, ClockID: tid.MaxClockID + 1, Expected: tid.ErrClockIDOutOfRange},
		{Microseconds: 0, ClockID: 65535, Expected: tid.ErrClockIDOutOfRange},
	}

	for testNumber, test := range tests {
		_, err := tid.Make(test.Microseconds, test.ClockID)
		if !erorr.Is(err, test.Expected) {
			t.Errorf("For test #%d, expected error %q but actually got: %v", testNumber, test.Expected, err)
		}
	}
}

func TestFromTime(t *testing.T) {
	value, err := tid.FromTime(time.Date(2023, time.November, 14, 22, 13, 20, 0, time.UTC), 5)
	if nil != err {
		t.Fatalf("Did not expect an error but actually got one: (%T) %s", err, err)
	}
	if expected, actual := "3ke6kg3wk2227", value.String(); expected != actual {
		t.Errorf("Expected %q but actually got %q.", expected, actual)
	}
}

// TIDs compare the same as integers and as strings.
func TestTID_order(t *testing.T) {
	var values = []string{"2222222222222", "2222222222223", "222222222222z", "2222222222232", "3jzfcijpj2z2a", "3ke6kg3wk2227", "bzzzzzzzzzzzz", "c222222222222", "jzzzzzzzzzzzz"}

	for i := 0; i+1 < len(values); i++ {
		a, errA := tid.Parse(values[i])
		b, errB := tid.Parse(values[i+1])
		if nil != errA || nil != errB {
			t.Fatalf("Did not expect an error but actually got: %v, %v", errA, errB)
		}
		if !(values[i] < values[i+1]) {
			t.Fatalf("Test data out of order: %q, %q", values[i], values[i+1])
		}
		if !(a < b) {
			t.Errorf("Expected %q < %q as TIDs too.", values[i], values[i+1])
		}
	}
}

func TestTID_json(t *testing.T) {
	type internalDocument struct {
		Rev tid.TID `json:"rev"`
	}

	var document internalDocument
	if err := json.Unmarshal([]byte(`{"rev":"3jzfcijpj2z2a"}`), &document); nil != err {
		t.Fatalf("Did not expect an error but actually got one: (%T) %s", err, err)
	}
	if expected, actual := uint64(1728652679052295174), uint64(document.Rev); expected != actual {
		t.Errorf("Expected %d but actually got %d.", expected, actual)
	}

	p, err := json.Marshal(document)
	if nil != err {
		t.Fatalf("Did not expect an error but actually got one: (%T) %s", err, err)
	}
	if expected, actual := `{"rev":"3jzfcijpj2z2a"}`, string(p); expected != actual {
		t.Errorf("Expected %s but actually got %s.", expected, actual)
	}

	if err := json.Unmarshal([]byte(`{"rev":"3JZFCIJPJ2Z2A"}`), &document); !erorr.Is(err, tid.ErrBadCharacter) {
		t.Errorf("Expected ErrBadCharacter but actually got: %v", err)
	}
}
