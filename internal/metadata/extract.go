package metadata

import (
	"fmt"
	"os"
	"strings"
)

// Extract metadata from file.
func Extract(file os.FileInfo) (string, string, string, error) {
	if !Validate(file) {
		return "", "", "", fmt.Errorf("invalid file format: %s", file.Name())
	}

	// Remove the ".log" from the file.
	name := strings.Replace(file.Name(), ".log", "", 1)

	// Split the string down so we can return their metadata.
	sl := strings.Split(name, "_")

	return sl[1], sl[0], sl[2], nil
}
