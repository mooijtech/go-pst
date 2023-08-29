package writer

import (
	"github.com/mooijtech/go-pst/v6/pkg/properties"
	"github.com/rotisserie/eris"
)

// MessageWriter represents a message that can be written to a PST file.
type MessageWriter struct {
	// Properties represents the properties of a pst.Message.
	Properties *properties.Message
	// Attachments represents the attachments of a pst.Message.
	Attachments []*AttachmentWriter

	// PropertyContextWriter writes the pst.PropertyContext of a pst.Message.
	PropertyContextWriter *PropertyContextWriter
}

// NewMessageWriter creates a new MessageWriter.
func NewMessageWriter(properties *properties.Message, attachments []*AttachmentWriter) *MessageWriter {
	return &MessageWriter{
		Properties:            properties,
		Attachments:           attachments,
		PropertyContextWriter: NewPropertyContextWriter(properties),
	}
}

// Write writes the message property context.
func (messageWriter *MessageWriter) Write() error {
	// Write Property Context.
	if err := messageWriter.PropertyContextWriter.Write(); err != nil {
		return eris.Wrap(err, "failed to write Property Context")
	}

	// Write attachments.
	for _, attachmentWriter := range messageWriter.Attachments {
		if err := attachmentWriter.Write(); err != nil {
			return eris.Wrap(err, "failed to write attachment")
		}
	}

	return nil
}
