package pipes

import (
	"errors"
	"io"
	"sync"
)

var (
	// ErrBroken is returned if a pipe is closed before all data is written to it.
	ErrBroken = errors.New("broken pipe")
)

type block struct {
	position int64
	data     []byte
}

// BlockPipe allows blocks of data to be piped out of order to a reader. It
// implements the WriterAt and Reader interfaces. Calls to WriterAt may be
// executed in parallel. Blocks written by WriterAt may not overlap. Doing so
// will result in undefiend behavior. Calls to read may not be done in
// parallel. The writer must close the pipe on completion. It may pass an error
// to the reader by calling CloseWithError. Read may return zero bytes. It will
// return an error when no more data is to be read. EOF indicates that all data
// was written successfully.
type BlockPipe struct {
	buffer   map[int64]*block // block buffer, holds blocks that haven't been read
	position int64            // current position of the reader
	current  *block           // current block being read
	err      error            // error to return on read
	cond     *sync.Cond
}

// NewBlocks creates a new block pipe.
func NewBlocks() *BlockPipe {
	return &BlockPipe{
		buffer: make(map[int64]*block),
		cond:   sync.NewCond(&sync.Mutex{}),
	}
}

// WriteAt writes a block of data at the given offset. It always returns the
// length of buf and nil.
func (p *BlockPipe) WriteAt(buf []byte, offset int64) (int, error) {
	p.cond.L.Lock()
	if p.buffer == nil {
	}
	blk := &block{
		position: offset,
		data:     make([]byte, len(buf)),
	}
	copy(blk.data, buf)
	p.buffer[offset] = blk
	p.cond.L.Unlock()
	p.cond.Broadcast()
	return len(buf), nil
}

// Read up to len(buf) bytes into buf. Blocks until data is ready to be read.
// Returns the number of bytes read and a nil error on success. Returns 0 and
// io.EOF when no more bytes are available. If CloseWithError is called then
// returns 0 and the provided error. If the pipe is closed before all blocks
// are written then ErrBroken is returned.
func (p *BlockPipe) Read(buf []byte) (int, error) {
	p.cond.L.Lock()
	defer p.cond.L.Unlock()

	for {
		if p.isAllBlocksRead() {
			return 0, p.err
		}
		if p.isCurrentBlockRead() {
			p.current = p.buffer[p.position]
			if p.current == nil {
				// wait for the next block to become available
				p.cond.Wait()
				continue
			}
			delete(p.buffer, p.position)
		}
		break
	}

	blkOffset := int(p.position - p.current.position)
	copySize := len(p.current.data) - blkOffset
	bufSize := len(buf)
	if copySize > bufSize {
		copySize = bufSize
	}
	copy(buf[0:copySize], p.current.data[blkOffset:blkOffset+copySize])
	p.position += int64(copySize)
	return copySize, nil
}

func (p *BlockPipe) isCurrentBlockRead() bool {
	if p.current == nil {
		return true
	}
	return p.position >= int64(len(p.current.data))+p.current.position
}

func (p *BlockPipe) isAllBlocksRead() bool {
	if p.err == io.EOF {
		endPos := p.position
		if p.current != nil {
			endPos = p.current.position + int64(len(p.current.data))
		}
		return len(p.buffer) == 0 && p.position >= endPos
	} else if p.err != nil {
		return true
	}
	return false
}

// CloseWithError closes the pipe and causes Read to return err the next time
// Read is called.
func (p *BlockPipe) CloseWithError(err error) {
	p.cond.L.Lock()
	p.err = err

	// if EOF, ensure the remaining blocks are contiguous
	if p.err == io.EOF {
		pos := p.position
		if p.current != nil {
			pos = p.current.position + int64(len(p.current.data))
		}

		var count int
		for p.buffer[pos] != nil {
			pos += int64(len(p.buffer[pos].data))
			count++
		}
		if count != len(p.buffer) {
			p.err = ErrBroken
		}
	}

	p.cond.L.Unlock()
	p.cond.Broadcast()
}

// Close the pipe. Read will consume the remaining blocks and return EOF on
// completion.
func (p *BlockPipe) Close() error {
	p.CloseWithError(io.EOF)
	return nil
}
