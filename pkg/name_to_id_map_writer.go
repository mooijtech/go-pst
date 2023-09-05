package pst

import "io"

// NameToIDMapWriter defines a writer for the Name-to-ID-Map.
// References https://github.com/mooijtech/go-pst/blob/main/docs/README.md#name-to-id-map
type NameToIDMapWriter struct {
	PropertyContextWriter *PropertyContextWriter
}

// NewNameToIDMapWriter creates a new NameToIDMapWriter.
func NewNameToIDMapWriter() *NameToIDMapWriter {
	return &NameToIDMapWriter{
		PropertyContextWriter: NewPropertyContextWriter(),
	}
}

func (nameToIDMapWriter *NameToIDMapWriter) WriteTo(writer io.Writer) (int64, error) {
	// The minimum requirement for the Name-to-ID Map is a PC node with a single property PidTagNameidBucketCount set to a value of 251 (0xFB)
	return nameToIDMapWriter.PropertyContextWriter.
}