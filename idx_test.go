package idx

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/oklog/ulid/v2"
)

func TestNewID(t *testing.T) {
	id := NewID()
	if id == NilID {
		t.Fatalf("%s is NilId", id.String())
	}
}

func TestCompare(t *testing.T) {
	if NilID.Compare(NilID) != 0 {
		t.Fatalf("NilID should be equal to NilID")
	}
	if NilID.Compare(NotNullNilID) != -1 {
		t.Fatalf("NilID should be less than NotNullNilId")
	}
	if NotNullNilID.Compare(NilID) != 1 {
		t.Fatalf("NotNullNilID should be greater than NilID")
	}
	if NotNullNilID.Compare(NotNullNilID) != 0 {
		t.Fatalf("NotNullNilID should be equal to NotNullNilID")
	}
	if NotNullNilID.Compare([16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2}) != -1 {
		t.Fatalf("NotNullNilID should be less")
	}
}

func TestFromString(t *testing.T) {
	id := NewID()
	idFromStr, err := FromString(id.String())
	if err != nil {
		t.Fatalf("Got error while creating ID from String %v", err)
	}
	if id != idFromStr || id.String() != idFromStr.String() {
		t.Fatalf("Original ID (%s) did not match with generated ID (%s)", id.String(), idFromStr.String())
	}
	invalidIds := []string{"null", "wrong", "00000", "01HAJ2Q3T69IJMMBDNAMVZ3FQB"}
	for _, val := range invalidIds {
		_, err = FromString(val)
		if err == nil {
			t.Fatalf("Was expecting error, not there was no error")
		}
	}
}

func TestIsValidID(t *testing.T) {
	id := NewID()
	if !IsValidID(id.String()) {
		t.Fatal("Expecting to be valid, but it's invalid")
	}

	invalidIds := []string{"null", "wrong", "00000", "01HAJ2Q3T69IJMMBDNAMVZ3FQB"}
	for _, val := range invalidIds {
		if IsValidID(val) {
			t.Fatalf("Expecting to be invalid, but it's valid")
		}
	}
}

func TestID_MarshalJSON(t *testing.T) {
	type IdTestStruct struct {
		Id ID `json:"id"`
	}
	id := NewID()
	jsonVal, err := json.Marshal(&IdTestStruct{Id: id})
	if err != nil {
		t.Fatalf("Got error while marshaling to JSON %v", err)
	}
	if string(jsonVal) != fmt.Sprintf(`{"id":"%s"}`, id.String()) {
		t.Fatalf("Original ID (%s) did not match with the ID in generated JSON %s", id.String(), string(jsonVal))
	}
}

func TestID_UnmarshalJSON(t *testing.T) {
	type IdTestStruct struct {
		ID ID `json:"id"`
	}
	id := NewID()
	jsonStrs := []string{
		`{"id":"01HAK8JPF7S0SFMJ2X96W37WXI"}`,
		`{"id":null}`,
		`{"id":""}`,
		fmt.Sprintf(`{"id":"%s"}`, id.String()),
	}
	idVals := []ID{
		NilID,
		NilID,
		NilID,
		id,
	}
	errVals := []error{
		ulid.ErrInvalidCharacters,
		nil,
		nil,
		nil,
	}
	unmVal := IdTestStruct{}
	for index, str := range jsonStrs {
		if err := json.Unmarshal([]byte(str), &unmVal); !errors.Is(err, errVals[index]) {
			t.Fatalf("Error did not match expectation %v : %v", err, errVals[index])
		}
		if unmVal.ID != idVals[index] {
			t.Fatalf("Original ID (%s) did not match with the ID from JSON %s %d", idVals[index].String(), unmVal.ID.String(), index)
		}
	}
}
