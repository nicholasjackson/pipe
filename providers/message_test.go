package providers

import (
	"testing"

	"github.com/matryer/is"
)

func TestGenerateMessageID(t *testing.T) {
	is := is.New(t)

	m1 := NewMessage()

	is.True(m1.ID != "")     // id should not be empty
	is.Equal(36, len(m1.ID)) // id should be 36 characters long
}

func TestGenerateRandomMessageID(t *testing.T) {
	is := is.New(t)

	m1 := NewMessage()
	m2 := NewMessage()

	is.True(m1.ID != m2.ID) // id1 should not be equal to id2
}

func TestSetsDefaultTimeStamp(t *testing.T) {
	is := is.New(t)

	m1 := NewMessage()

	is.True(m1.Timestamp > 1) // timestamp should be set
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
