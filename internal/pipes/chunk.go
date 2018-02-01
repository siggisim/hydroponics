package pipes

import (
	"io"
)

// Chunk is a fixed size byte buffer.
type Chunk struct {
	data []byte
	len  int
}

// NewChunk creates a chunk of the requested size.
func NewChunk(size int) *Chunk {
	return &Chunk{
		data: make([]byte, size),
	}
}

// Write data to the chunk. Returns the number of bytes written to the chunk.
// If the chunk has no capacity then 0 and io.EOF are returned.
func (c *Chunk) Write(buf []byte) (int, error) {
	if c.Full() {
		return 0, io.EOF
	}
	bufSize := len(buf)
	copySize := min(c.Cap()-c.Len(), bufSize)
	copy(c.data[c.len:copySize], buf[0:copySize])
	c.len += copySize
	return copySize, nil
}

// Return the current bytes in the chunk.
func (c *Chunk) Bytes() []byte {
	return c.data[:c.len]
}

// Len returns the number of bytes written to the chunk.
func (c *Chunk) Len() int {
	return c.len
}

// Cap returns the maximum number of bytes the chunk can store.
func (c *Chunk) Cap() int {
	return cap(c.data)
}

// Full returns true if the chunk is at capacity.
func (c *Chunk) Full() bool {
	return c.Len() == c.Cap()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
