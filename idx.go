package idx

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/oklog/ulid/v2"
)

type ID [16]byte

var NilID ID

var NotNullNilID = ID([16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1})

func NewID() ID {
	return ID(ulid.Make())
}

func FromString(val string) (ID, error) {
	ulidVal, err := ulid.ParseStrict(val)
	if err != nil {
		return NilID, err
	}
	return ID(ulidVal), nil
}

func IsValidID(val string) bool {
	_, err := ulid.ParseStrict(val)
	if err != nil {
		return false
	}
	return true
}

func (id ID) String() string {
	return ulid.ULID(id).String()
}

func (id ID) IsZero() bool {
	return id == NilID
}

func (id ID) Compare(y ID) int {
	return ulid.ULID(id).Compare(ulid.ULID(y))
}

// MarshalText returns the IDX as UTF-8-encoded text. Implementing this allows us to use IDX
// as a map key when marshalling JSON. See https://pkg.go.dev/encoding#TextMarshaler
func (id ID) MarshalText() ([]byte, error) {
	return ulid.ULID(id).MarshalText()
}

// UnmarshalText populates the byte slice with the ObjectID. Implementing this allows us to use ObjectID
// as a map key when unmarshalling JSON. See https://pkg.go.dev/encoding#TextUnmarshaler
func (id *ID) UnmarshalText(b []byte) error {
	// The ulid UnmarshalText runs in non-strict mode,
	// therefore doing a strict check of characters to avoid passing un-allowed characters
	if dec[b[0]] == 0xFF ||
		dec[b[1]] == 0xFF ||
		dec[b[2]] == 0xFF ||
		dec[b[3]] == 0xFF ||
		dec[b[4]] == 0xFF ||
		dec[b[5]] == 0xFF ||
		dec[b[6]] == 0xFF ||
		dec[b[7]] == 0xFF ||
		dec[b[8]] == 0xFF ||
		dec[b[9]] == 0xFF ||
		dec[b[10]] == 0xFF ||
		dec[b[11]] == 0xFF ||
		dec[b[12]] == 0xFF ||
		dec[b[13]] == 0xFF ||
		dec[b[14]] == 0xFF ||
		dec[b[15]] == 0xFF ||
		dec[b[16]] == 0xFF ||
		dec[b[17]] == 0xFF ||
		dec[b[18]] == 0xFF ||
		dec[b[19]] == 0xFF ||
		dec[b[20]] == 0xFF ||
		dec[b[21]] == 0xFF ||
		dec[b[22]] == 0xFF ||
		dec[b[23]] == 0xFF ||
		dec[b[24]] == 0xFF ||
		dec[b[25]] == 0xFF {
		return ulid.ErrInvalidCharacters
	}
	return (*ulid.ULID)(id).UnmarshalText(b)
}

// MarshalJSON returns the IDX as a string
func (id ID) MarshalJSON() ([]byte, error) {
	return json.Marshal(id.String())
}

// UnmarshalJSON populates the byte slice with the IDX. If the byte slice is 16 bytes long, it
// will be populated with the hex representation of the IDX. If the byte slice is twelve bytes
// long, it will be populated with the BSON representation of the IDX. This method also accepts empty strings and
// decodes them as NilID. For any other inputs, an error will be returned.
func (id *ID) UnmarshalJSON(b []byte) error {
	idLen := len(b)
	if idLen == 2 && b[0] == 0x22 && b[1] == 0x22 {
		return nil
	}
	// If the value is "null"
	if idLen == 4 && b[0] == 0x6E && b[1] == 0x75 && b[2] == 0x6C && b[3] == 0x6C {
		return nil
	}
	if idLen == 28 {
		if err := id.UnmarshalText(b[1:27]); err != nil {
			return err
		}
		return nil
	}
	return ulid.ErrDataSize
}

// Scan implements the sql.Scanner interface. It supports scanning
// a string or byte slice.
func (id *ID) Scan(src interface{}) error {
	// If value is nil, set the ID to NilID
	if src == nil {
		copy(id[:], NilID[:])
	}
	return (*ulid.ULID)(id).Scan(src)
}

// Value implements the sql/driver.Valuer interface, returning the ID as a
// slice of bytes, by invoking MarshalBinary. If your use case requires a string
// representation instead, you can create a wrapper type that calls String()
// instead.
//
//	type stringValuer idx.ID
//
//	func (v stringValuer) Value() (driver.Value, error) {
//	    return idx.ID(v).String(), nil
//	}
//
//	// Example usage.
//	db.Exec("...", stringValuer(id))
//
// All valid ULIDs, including zero-value ULIDs, return a valid Value with a nil
// error. If your use case requires zero-value ULIDs to return a non-nil error,
// you can create a wrapper type that special-cases this behavior.
//
//	var zeroValueULID idx.ID
//
//	type invalidZeroValuer idx.ID
//
//	func (v invalidZeroValuer) Value() (driver.Value, error) {
//	    if idx.ID(v).Compare(zeroValueULID) == 0 {
//	        return nil, fmt.Errorf("zero value")
//	    }
//	    return idx.ID(v).Value()
//	}
//
//	// Example usage.
//	db.Exec("...", invalidZeroValuer(id))
func (id ID) Value() (driver.Value, error) {
	// If the ID is NilID, return nil
	if id == NilID {
		return nil, nil
	}
	return ulid.ULID(id).Value()
}

// Byte to index table for O(1) lookups when unmarshaling.
// We use 0xFF as sentinel value for invalid indexes.
var dec = [...]byte{
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x00, 0x01,
	0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E,
	0x0F, 0x10, 0x11, 0xFF, 0x12, 0x13, 0xFF, 0x14, 0x15, 0xFF,
	0x16, 0x17, 0x18, 0x19, 0x1A, 0xFF, 0x1B, 0x1C, 0x1D, 0x1E,
	0x1F, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x0A, 0x0B, 0x0C,
	0x0D, 0x0E, 0x0F, 0x10, 0x11, 0xFF, 0x12, 0x13, 0xFF, 0x14,
	0x15, 0xFF, 0x16, 0x17, 0x18, 0x19, 0x1A, 0xFF, 0x1B, 0x1C,
	0x1D, 0x1E, 0x1F, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
}
