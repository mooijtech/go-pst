package writer

import "github.com/rotisserie/eris"

// MessageStoreWriter represents a writer for Message Stores.
type MessageStoreWriter struct {
	// PropertyContextWriter represents the pst.PropertyContext writer.
	PropertyContextWriter *PropertyContextWriter
}

// TODO - properties.MessageStore

// NewMessageStoreWriter creates a new MessageStoreWriter.
func NewMessageStoreWriter(propertyContextWriter *PropertyContextWriter) *MessageStoreWriter {
	return &MessageStoreWriter{
		PropertyContextWriter: propertyContextWriter,
	}
}

// Write writes the Message Store.
func (messageStoreWriter *MessageStoreWriter) Write() error {
	if err := messageStoreWriter.PropertyContextWriter.Write(); err != nil {
		return eris.Wrap(err, "failed to write Message Store")
	}

	return nil
}
