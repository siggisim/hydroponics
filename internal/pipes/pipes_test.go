package pipes_test

import (
	"bytes"
	"errors"
	"io"
	"sync"
	"testing"

	"github.com/zenreach/hydroponics/internal/pipes"
)

func TestReadOneBlock(t *testing.T) {
	want := []byte("hello")
	blocks := [][]byte{want}
	pipe := pipes.NewBlocks()

	go writeBlocks(t, pipe, blocks)
	assertReadOnce(t, pipe, want)
}

func TestReadMultipleBlocks(t *testing.T) {
	want := []byte("hello")
	blocks := [][]byte{[]byte("he"), []byte("llo")}
	pipe := pipes.NewBlocks()

	go writeBlocks(t, pipe, blocks)
	assertRead(t, pipe, want)
}

func TestCloseWithError(t *testing.T) {
	want := errors.New("oops!")

	pipe := pipes.NewBlocks()
	pipe.WriteAt([]byte("hello"), 0)

	buf := make([]byte, 2)
	pipe.Read(buf)
	pipe.CloseWithError(want)

	n, have := pipe.Read(buf)
	if have != want {
		t.Errorf("incorrect error: \"%s\" != \"%s\"", have, want)
	} else if n != 0 {
		t.Errorf("returned incorrect count: %d != 0", n)
	}
}

func TestIncompleteWrites(t *testing.T) {
	pipe := pipes.NewBlocks()
	pipe.WriteAt([]byte("he"), 0)
	pipe.WriteAt([]byte("o"), 4)
	pipe.Close()

	buf := make([]byte, 10)
	n, err := pipe.Read(buf)
	if err != pipes.ErrBroken {
		t.Errorf("incorrect error: \"%s\" != \"%s\"", err, pipes.ErrBroken)
	} else if n != 0 {
		t.Errorf("returned incorrect count: %d != 0", n)
	}
}

func writeBlocks(t *testing.T, wr *pipes.BlockPipe, blocks [][]byte) {
	wg := &sync.WaitGroup{}
	wg.Add(len(blocks))

	var offset int64
	for i := range blocks {
		block := blocks[i]
		go func(block []byte, offset int64) {
			n, err := wr.WriteAt(block, offset)
			if err != nil {
				t.Errorf("error while writing block \"%s\": %s", block, err)
			} else if n != len(block) {
				t.Errorf("entire block not written: %d != %d", n, len(block))
			}
			wg.Done()
		}(block, offset)
		offset += int64(len(block))
	}
	wg.Wait()
	err := wr.Close()
	if err != nil {
		t.Errorf("error while closing: %s", err)
	}
}

func assertBytesEqual(t *testing.T, left, right []byte) {
	// all tests use strings converted to bytes, so this is safe
	if left == nil && right != nil {
		t.Errorf("bytes not equal: nil != \"%s\"", right)
	} else if left != nil && right == nil {
		t.Errorf("bytes not equal: \"%s\" != nil", left)
	} else if string(left) != string(right) {
		t.Errorf("bytes not equal: \"%s\" != \"%s\"", left, right)
	}
}

func assertRead(t *testing.T, rdr io.Reader, expect []byte) {
	buf := &bytes.Buffer{}
	n, err := io.Copy(buf, rdr)
	if err != nil {
		t.Errorf("read error: %s", err)
		return
	}
	if n != int64(len(expect)) {
		t.Errorf("read incorrect byte count: %d != %d", n, len(expect))
	}
	assertBytesEqual(t, buf.Bytes(), expect)
}

func assertReadOnce(t *testing.T, rdr io.Reader, expect []byte) {
	buf := make([]byte, len(expect))
	n, err := rdr.Read(buf)
	if err != nil {
		t.Errorf("read error: %s", err)
		return
	}
	if n != len(expect) {
		t.Errorf("read incorrect byte count: %d != %d", n, len(expect))
	}
	assertBytesEqual(t, buf, expect)
}
