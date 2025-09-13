package lipbalm

// Message represents a simple text message with semantic styling
type Message struct {
	Text  string
	Level string // success, error, warning, info, or empty for plain text
}

// Common message levels
const (
	LevelSuccess = "success"
	LevelError   = "error"
	LevelWarning = "warning"
	LevelInfo    = "info"
)

// NewMessage creates a new message with the specified level
func NewMessage(text string, level string) *Message {
	return &Message{
		Text:  text,
		Level: level,
	}
}

// NewSuccessMessage creates a success message
func NewSuccessMessage(text string) *Message {
	return NewMessage(text, LevelSuccess)
}

// NewErrorMessage creates an error message
func NewErrorMessage(text string) *Message {
	return NewMessage(text, LevelError)
}

// NewWarningMessage creates a warning message
func NewWarningMessage(text string) *Message {
	return NewMessage(text, LevelWarning)
}

// NewInfoMessage creates an info message
func NewInfoMessage(text string) *Message {
	return NewMessage(text, LevelInfo)
}

// NewPlainMessage creates a message with no styling
func NewPlainMessage(text string) *Message {
	return &Message{Text: text}
}