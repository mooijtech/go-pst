package writer

// WriteBuffer is a helper for copying (appending) bytes without specifying the offset.
func WriteBuffer(inputBuffer []byte, outputBuffer []byte) {
	copy(outputBuffer[len(outputBuffer):len(outputBuffer)+len(inputBuffer)], inputBuffer)
}
