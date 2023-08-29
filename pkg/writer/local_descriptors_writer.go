package writer

// LocalDescriptorsWriter represents a writer for Local Descriptors.
// (B-Tree Nodes pointing to other B-Tree Nodes)
type LocalDescriptorsWriter struct {
}

// NewLocalDescriptorsWriter creates a new LocalDescriptorsWriter.
func NewLocalDescriptorsWriter() *LocalDescriptorsWriter {
	return &LocalDescriptorsWriter{}
}

// Write writes the Local Descriptors.
func (localDescriptorsWriter *LocalDescriptorsWriter) Write() error {
	return nil
}
