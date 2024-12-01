package device

import (
	"os"
)

// OpenDummyDevice opens file from the name
func OpenDummyDevice(name string) (*os.File, string, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, name, err
	}

	return f, name, nil
}
