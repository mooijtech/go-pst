package writer

import (
	pst "github.com/mooijtech/go-pst/v6/pkg"
	"github.com/rotisserie/eris"
)

// BTreeOnHeapWriter writes a BTree-On-Heap
type BTreeOnHeapWriter struct {
	// HeapOnNodeWriter represents the HeapOnNodeWriter.
	HeapOnNodeWriter *HeapOnNodeWriter
}

// NewBTreeOnHeapWriter creates a new BTreeOnHeapWriter.
func NewBTreeOnHeapWriter(heapOnNodeWriter *HeapOnNodeWriter) *BTreeOnHeapWriter {
	return &BTreeOnHeapWriter{HeapOnNodeWriter: heapOnNodeWriter}
}

// Write writes the BTree-on-Heap.
// References https://github.com/mooijtech/go-pst/blob/main/docs/README.md#inserting-into-the-bth
func (btreeOnHeapWriter *BTreeOnHeapWriter) Write() error {
	if err := btreeOnHeapWriter.HeapOnNodeWriter.Write(); err != nil {
		return eris.Wrap(err, "failed to write Heap-on-Node")
	}
	if err := btreeOnHeapWriter.WriteHeader(); err != nil {
		return eris.Wrap(err, "failed to write BTree-On-Heap header")
	}

	return nil
}

// WriteHeader writes the BTree-on-Heap header.
func (btreeOnHeapWriter *BTreeOnHeapWriter) WriteHeader() error {
	header := make([]byte, 0) // TODO - Correct size

	WriteBuffer([]byte{byte(pst.SignatureTypeBTreeOnHeap)}, header) // MUST be bTypeBTH.

	return nil
}
