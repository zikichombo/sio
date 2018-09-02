package entry

import "errors"

var (
	// ErrInvalidEntryName is used on RegisterEntry to
	// enforce the use of known entry points.
	ErrInvalidEntryName = errors.New("unknown entry")
	// ErrUnsupported is returned when the entry in use
	// does not support the requested operation.
	ErrUnsupported = errors.New("unsupported")
	// ErrNoEntryAvailable indicates there are no entry ports
	// for the host.
	ErrNoEntryAvailable = errors.New("no entry available")
	// ErrEntryInUse indicates that the caller requested an
	// entry when another is in use.
	ErrEntryInUse = errors.New("entry in use")
)
