/*
Package tid implements TIDs — ATProto's "timestamp identifiers" — for the Go programming language.

A TID is a 64-bit integer, written as 13 characters of "base32-sortable" (the characters 234567abcdefghijklmnopqrstuvwxyz), most significant first.
For example:

	3jzfcijpj2z2a

Its layout is:

	• the top bit — always 0 (for TIDs made by this package),
	• the next 53 bits — a timestamp: microseconds since the UNIX epoch,
	• the last 10 bits — a "clock identifier", chosen at random by whoever made it.

Because of the fixed length and the alphabet's order, TIDs sort the same as strings and as integers.
In ATProto, TIDs are used as repo revisions ("rev") and as record keys.

The TID spec is here:

https://atproto.com/specs/tid

# Parsing

[Parse] accepts exactly what the spec's syntax allows:

	^[234567abcdefghij][234567abcdefghijklmnopqrstuvwxyz]{12}$

(That allows every 64-bit value — including those with the top bit set, which no well-behaved generator makes.
It rejects anything that does not fit in 64 bits, upper-case, and the old dashed syntax.)

# Making

[Make] and [FromTime] make a TID from a timestamp and a clock identifier.
A [Generator] makes TIDs from the current time that always increase — even within one microsecond, or if the clock goes backwards.

# Example

	t, err := tid.Parse("3jzfcijpj2z2a")
	if nil != err {
		return err
	}

	fmt.Println(t.Time())    // 2023-06-30 15:03:01.887007 +0000 UTC
	fmt.Println(t.ClockID()) // 6

	var generator = tid.NewGenerator()
	next := generator.Next()
	fmt.Println(next.String())
*/
package tid
