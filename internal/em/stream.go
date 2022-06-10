package em

import (
	"bufio"
	"bytes"
	"encoding/binary"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/pkg/errors"
)

type streamRW struct {
	s   network.Stream
	buf *bytes.Buffer
	rw  *bufio.ReadWriter
}

func NewStreamRW(s network.Stream) *streamRW {
	rw := &streamRW{
		s:   s,
		buf: new(bytes.Buffer),
		rw:  bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s)),
	}
	return rw
}

func (c *streamRW) Write(b []byte) error {
	c.buf.Reset()
	c.buf.Grow(4)
	binary.LittleEndian.PutUint32(c.buf.Bytes()[:4], uint32(len(b)))

	if _, err := c.rw.Write(c.buf.Bytes()[:4]); err != nil {
		return errors.Wrap(err, "transmitting len")
	}

	if _, err := c.rw.Write(b); err != nil {
		return errors.Wrap(err, "transmitting data")
	}

	return c.rw.Flush()
}

func (c *streamRW) Read() ([]byte, error) {
	c.buf.Reset()
	c.buf.Grow(4)
	if _, err := c.rw.Read(c.buf.Bytes()[:4]); err != nil {
		return nil, err
	}
	len := binary.LittleEndian.Uint32(c.buf.Bytes()[:4])

	c.buf.Reset()
	c.buf.Grow(int(len))
	n, err := c.rw.Read(c.buf.Bytes()[:len])
	if err != nil {
		return nil, err
	}

	return c.buf.Bytes()[:n], nil
}
