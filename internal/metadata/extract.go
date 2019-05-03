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
	underscore := strings.Split(name, "_")
	dash := strings.Split(underscore[2], "-")

	return underscore[1], underscore[0], dash[0], nil
}
