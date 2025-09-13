package output

// Message represents a simple text message with optional styling
type Message struct {
	Text  string
	Level string // success, error, warning, info, or empty for plain text
}

// NewMessage creates a new message with default info level
func NewMessage(text string) *Message {
	return &Message{
		Text:  text,
		Level: "info",
	}
}

// NewPlainMessage creates a new message with no styling
func NewPlainMessage(text string) *Message {
	return &Message{
		Text: text,
	}
}

// WithLevel sets the message level
func (m *Message) WithLevel(level string) *Message {
	m.Level = level
	return m
}