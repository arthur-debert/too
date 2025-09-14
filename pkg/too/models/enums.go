package models

// CommandType represents the category of a command
type CommandType string

const (
	CommandTypeCore  CommandType = "core"
	CommandTypeExtra CommandType = "extras"
	CommandTypeMisc  CommandType = "misc"
)

// AttributeType represents the type of attribute to mutate
type AttributeType string

const (
	AttributeCompletion AttributeType = "completion"
	AttributeText       AttributeType = "text"
	AttributeParent     AttributeType = "parent"
)