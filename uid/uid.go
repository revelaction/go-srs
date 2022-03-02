package uid

// UID represents an unique id process
type UID interface {
	Create() string
	Validate(uid string) error
}
