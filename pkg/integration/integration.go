package integration

import (
	"context"
	"time"
)

// Integration represents an external service that can be integrated
type Integration interface {
	// Name returns the human-readable name of the integration
	Name() string

	// Slug returns a unique identifier for the integration
	Slug() string

	// Version returns the version of the integration
	Version() string

	// Description provides a human-readable description
	Description() string

	// AuthType returns the type of authentication this integration uses
	AuthType() AuthType

	// GetAuthFields returns the fields needed for authentication
	GetAuthFields() []AuthField

	// Authenticate performs the authentication process
	Authenticate(ctx context.Context, authData map[string]string) (Credentials, error)

	// Triggers returns all available triggers for this integration
	Triggers() []Trigger

	// Actions returns all available actions for this integration
	Actions() []Action
}

// AuthType represents different authentication methods
type AuthType string

const (
	AuthTypeOAuth2 AuthType = "oauth2"
	AuthTypeAPIKey AuthType = "api_key"
	AuthTypeBasic  AuthType = "basic"
	AuthTypeCustom AuthType = "custom"
)

// AuthField represents a field needed for authentication
type AuthField struct {
	Key         string
	Label       string
	Type        string
	Required    bool
	Placeholder string
	Help        string
}

// Credentials represents authentication credentials for an integration
type Credentials interface {
	// Valid checks if the credentials are still valid
	Valid() bool

	// ExpiresAt returns when the credentials expire, if applicable
	ExpiresAt() *time.Time

	// Refresh attempts to refresh the credentials
	Refresh(ctx context.Context) error

	// AsMap returns the credentials as a map for storage
	AsMap() map[string]string
}

// Trigger represents an event that can trigger a workflow
type Trigger interface {
	// ID returns the unique identifier for this trigger
	ID() string

	// Name returns the human-readable name
	Name() string

	// Description provides a human-readable description
	Description() string

	// InputFields returns fields that configure this trigger
	InputFields() []Field

	// OutputFields describes the output data structure
	OutputFields() []Field

	// Execute registers or polls for the trigger
	Execute(ctx context.Context, input map[string]interface{}, credentials Credentials) (TriggerJob, error)
}

// TriggerJob represents a running trigger job
type TriggerJob interface {
	// ID returns the unique identifier for this job
	ID() string

	// Status returns the current status
	Status() JobStatus

	// Stop cancels the trigger job
	Stop(ctx context.Context) error

	// Events returns a channel of trigger events
	Events() <-chan Event
}

// Action represents an operation that can be performed
type Action interface {
	// ID returns the unique identifier for this action
	ID() string

	// Name returns the human-readable name
	Name() string

	// Description provides a human-readable description
	Description() string

	// InputFields returns fields that configure this action
	InputFields() []Field

	// OutputFields describes the output data structure
	OutputFields() []Field

	// Execute performs the action
	Execute(ctx context.Context, input map[string]interface{}, credentials Credentials) (map[string]interface{}, error)
}

// Field represents a data field in an integration
type Field struct {
	Key         string
	Label       string
	Type        FieldType
	Required    bool
	Default     interface{}
	Choices     []Choice
	Placeholder string
	Help        string
	Computed    bool
	DependsOn   []string
}

// FieldType represents different types of fields
type FieldType string

const (
	FieldTypeString      FieldType = "string"
	FieldTypeNumber      FieldType = "number"
	FieldTypeBoolean     FieldType = "boolean"
	FieldTypeObject      FieldType = "object"
	FieldTypeArray       FieldType = "array"
	FieldTypeDateTime    FieldType = "datetime"
	FieldTypeSelect      FieldType = "select"
	FieldTypeMultiSelect FieldType = "multiselect"
)

// Choice represents an option for a select field
type Choice struct {
	Value interface{}
	Label string
}

// Event represents data from a trigger event
type Event struct {
	ID        string
	Timestamp time.Time
	Data      map[string]interface{}
	Raw       []byte
}

// JobStatus represents the status of a trigger job
type JobStatus string

const (
	JobStatusRunning JobStatus = "running"
	JobStatusStopped JobStatus = "stopped"
	JobStatusError   JobStatus = "error"
)
