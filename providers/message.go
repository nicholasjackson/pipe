package providers

import "github.com/google/uuid"

type Message struct {
	// ID is a random message ID
	ID string
	// ParentID allows a message to be traced through a pipeline, actions, success and fail messages
	// will have the a parent ID equal to the recieved message ID
	ParentID string
	// Data is an arbitrary data payload for the message
	Data []byte
	// ContentType for the message if known
	ContentType string
	// Timestamp for the message
	Timestamp int64
	// Redelivered is set if this message is a re-delivery
	Redelivered bool
	// Sequence number for the message
	Sequence uint64
}

// Ack acknowledged receipt and processing of a message
func (m *Message) Ack() {

}

// Copy the message to a new instance
func (m *Message) Copy() Message {
	return *m
}

// GenerateRandomMessageID generates a random UUID to spec RFC 4122
func GenerateRandomMessageID() string {
	u, err := uuid.NewUUID()
	if err != nil {
		panic(err) // this should not occur, panic as we have no handling for this
	}

	return u.String()
}
