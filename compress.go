package gws

import (
	"bytes"
	"compress/flate"
	"encoding/binary"
	"github.com/lxzan/gws/internal"
	"io"
	"math"
)

func newCompressor(level int) *compressor {
	fw, _ := flate.NewWriter(nil, level)
	return &compressor{fw: fw}
}

// 压缩器
type compressor struct {
	buffer *bytes.Buffer
	fw     *flate.Writer
}

func (c *compressor) Close() {
	_bpool.Put(c.buffer)
}

// Compress 压缩
func (c *compressor) Compress(content *bytes.Buffer) (*bytes.Buffer, error) {
	c.buffer = _bpool.Get(content.Len())
	c.fw.Reset(c.buffer)
	if err := internal.WriteN(c.fw, content.Bytes(), content.Len()); err != nil {
		return nil, err
	}
	if err := c.fw.Flush(); err != nil {
		return nil, err
	}

	if n := c.buffer.Len(); n >= 4 {
		compressedContent := c.buffer.Bytes()
		if tail := compressedContent[n-4:]; binary.BigEndian.Uint32(tail) == math.MaxUint16 {
			c.buffer.Truncate(n - 4)
		}
	}
	return c.buffer, nil
}

func newDecompressor() *decompressor { return &decompressor{fr: flate.NewReader(nil)} }

type decompressor struct {
	fr io.ReadCloser
}

// Decompress 解压
func (c *decompressor) Decompress(payload *bytes.Buffer) (*bytes.Buffer, error) {
	_, _ = payload.Write(internal.FlateTail)
	resetter := c.fr.(flate.Resetter)
	_ = resetter.Reset(payload, nil) // must return a null pointer

	var buf = _bpool.Get(3 * payload.Len())
	_, err := io.Copy(buf, c.fr)
	_bpool.Put(payload)
	return buf, err
}
