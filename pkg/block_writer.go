package pst

import (
	"bytes"
	"github.com/rotisserie/eris"
	"io"
)

// BlockWriter represents a writer for XBlocks and XXBlocks.
// References https://github.com/mooijtech/go-pst/blob/main/docs/README.md#data-tree
type BlockWriter struct {
	// FormatType represents the FormatType used while writing.
	FormatType FormatType
	// BlockWriteChannel represents a Go channel used for writing blocks.
	BlockWriteChannel chan Identifier
	// BlockWriteCallback represents the callback which is called once a block is written.
	BlockWriteCallback chan int
}

// NewBlockWriter creates a new BlockWriter.
func NewBlockWriter(formatType FormatType) *BlockWriter {
	return &BlockWriter{
		FormatType: formatType,
	}
}

// WriteTo writes the blocks (XBlocks and XXBlocks).
func (blockWriter *BlockWriter) WriteTo(writer io.Writer) (int64, error) {
	return blockWriter.WriteXBlock(writer)
}

// WriteXBlock writes the XBlock.
// References https://github.com/mooijtech/go-pst/blob/main/docs/README.md#xblock
func (blockWriter *BlockWriter) WriteXBlock(writer io.Writer) (int64, error) {
	// 1+1+2+4+identifiers+padding+blockTrailer
	xBlockBuffer := bytes.NewBuffer(make([]byte, 0))

	// Block type; MUST be set to 0x01 to indicate an XBLOCK or XXBLOCK.
	xBlockBuffer.WriteByte(byte(1))
	// MUST be set to 0x01 to indicate an XBLOCK.
	xBlockBuffer.WriteByte(byte(1))
	// The count of identifiers in the XBLOCK.
	xBlockBuffer.Write(GetUint16(uint16(len(blockWriter.Identifiers))))
	// Total count of bytes of all the external data stored in the data blocks referenced by XBLOCK.
	// TODO
	// Array of identifiers that reference data blocks.
	// The size is equal to the number of entries indicated by cEnt multiplied by the size of a BID
	// (8 bytes for Unicode PST files, 4 bytes for ANSI PST files).
	for _, identifier := range blockWriter.Identifiers {
		switch blockWriter.FormatType {
		case FormatTypeUnicode:
			xBlockBuffer.Write(GetUint64(uint64(identifier)))
		case FormatTypeANSI:
			xBlockBuffer.Write(GetUint32(uint32(identifier)))
		default:
			return 0, ErrFormatTypeUnsupported
		}
	}
	// This field is present if the total size of all the other fields is not a multiple of 64.
	// The size of this field is the smallest number of bytes required to make the size of the XBLOCK a multiple of 64.
	// TODO -
	// A BLOCKTRAILER structure
	if _, err := blockWriter.WriteBlockTrailer(xBlockBuffer); err != nil {
		return 0, eris.Wrap(err, "failed to write block trailer")
	}

	return xBlockBuffer.WriteTo(writer)
}

// WriteBlockTrailer writes the block trailer.
// References https://github.com/mooijtech/go-pst/blob/main/docs/README.md#blocktrailer
func (blockWriter *BlockWriter) WriteBlockTrailer(writer io.Writer) (int64, error) {
	// The amount of data, in bytes, contained within the data section of the block.
	// This value does not include the block trailer or any unused bytes that can
	// exist after the end of the data and before the start of the block trailer.
	// TODO -
	// Block signature. See section 5.5 for the algorithm to calculate the block signature.
	// TODO -
	// 32-bit CRC of the cb bytes of raw data, see section 5.3 for the algorithm to calculate the CRC.
	// Note the locations of the dwCRC and bid are differs between the Unicode and ANSI version of this structure.
	// TODO -
	// The BID (section 2.2.2.2) of the data block.

	return 0, nil
}
