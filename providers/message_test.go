package providers

import (
	"testing"

	"github.com/matryer/is"
)

func TestGenerateMessageID(t *testing.T) {
	is := is.New(t)

	id1 := GenerateRandomMessageID()

	is.True(id1 != "")     // id should not be empty
	is.Equal(16, len(id1)) // id should be 16 characters long
}

func TestGenerateRandomMessageID(t *testing.T) {
	is := is.New(t)

	id1 := GenerateRandomMessageID()
	id2 := GenerateRandomMessageID()

	is.True(id1 != id2) // id1 should not be equal to id2
}

func TestCopyDuplicatesAMessage(t *testing.T) {
	is := is.New(t)

	msg := Message{
		ID:          "id",
		ParentID:    "parentid",
		Data:        []byte("abc"),
		ContentType: "text",
		Timestamp:   12345,
		Redelivered: true,
		Sequence:    124,
	}

	copy := msg.Copy()

	is.Equal(msg, copy) // copy should be a duplicate of the original message
}
