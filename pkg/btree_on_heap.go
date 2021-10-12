// Package pst
// This file is part of go-pst (https://github.com/mooijtech/go-pst)
// Copyright (C) 2021 Marten Mooij (https://www.mooijtech.com/)
package pst

// BTreeOnHeapHeader represents the b-tree on heap header.
type BTreeOnHeapHeader struct {
	TableType int
	KeySize   int
	ValueSize int
	Levels    int
	HIDRoot   int
}

// GetBTreeOnHeapHeader returns the btree on heap header.
func (pstFile *File) GetBTreeOnHeapHeader(heapOnNode HeapOnNode, localDescriptors []LocalDescriptor, formatType string, encryptionType string) (BTreeOnHeapHeader, error) {
	// All tables should have a BTree-on-Heap header at HID 0x20 (HID User Root from the Heap-on-Node header).
	hidUserRoot, err := heapOnNode.GetHIDUserRoot()

	if err != nil {
		return BTreeOnHeapHeader{}, err
	}

	inputStream, err := pstFile.NewHeapOnNodeInputStreamFromHNID(hidUserRoot, heapOnNode, localDescriptors, formatType, encryptionType)

	if err != nil {
		return BTreeOnHeapHeader{}, err
	}

	btreeOnHeapTableType, err := inputStream.SeekAndReadUint16(1, 0)

	if err != nil {
		return BTreeOnHeapHeader{}, err
	}

	btreeOnHeapKeySize, err := inputStream.SeekAndReadUint16(1, 1)

	if err != nil {
		return BTreeOnHeapHeader{}, err
	}

	btreeOnHeapValueSize, err := inputStream.SeekAndReadUint16(1, 2)

	if err != nil {
		return BTreeOnHeapHeader{}, err
	}

	btreeOnHeapLevels, err := inputStream.SeekAndReadUint16(1, 3)

	if err != nil {
		return BTreeOnHeapHeader{}, err
	}

	btreeOnHeapHIDRoot, err := inputStream.SeekAndReadUint32(4, 4)

	if err != nil {
		return BTreeOnHeapHeader{}, err
	}

	return BTreeOnHeapHeader{
		TableType: btreeOnHeapTableType,
		KeySize:   btreeOnHeapKeySize,
		ValueSize: btreeOnHeapValueSize,
		Levels:    btreeOnHeapLevels,
		HIDRoot:   btreeOnHeapHIDRoot,
	}, nil
}
