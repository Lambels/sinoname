package sinoname

import "fmt"

// MessagePacket represents a message passed through the pipeline.
type MessagePacket struct {
	// Copy of the actuall message.
	Message string
	// Number of changes the message has gone through.
	Changes int
	// Skip indicates a countdown on how many layers the message has to skip.
	Skip int
}

func (m MessagePacket) String() string {
	return fmt.Sprintf("%v , Changes: %v , Skips: %v", m.Message, m.Changes, m.Skip)
}

func (m *MessagePacket) setAndIncrement(v string) {
	m.Message = v
	m.Changes++
}
