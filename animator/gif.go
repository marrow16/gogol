package animator

import (
	"compress/lzw"
	"errors"
	"image"
	"image/color"
	"io"
	"math/bits"
)

type gifEncoder struct {
	w                io.Writer
	err              error
	buf              [256]byte
	width, height    int
	palette          []color.RGBA
	globalColorTable [3 * 256]byte
	loopCount        int
	delay            int
}

const (
	fColorTable           = 1 << 7
	sTrailer              = 0x3B
	sImageDescriptor      = 0x2C
	sExtension            = 0x21
	gcLabel               = 0xF9
	gcBlockSize           = 0x04
	disposal         byte = 0x00
)

func (e *gifEncoder) end() error {
	e.writeByte(sTrailer)
	return e.err
}

func (e *gifEncoder) writeHeader() {
	if e.err != nil {
		return
	}
	e.writeString("GIF89a")
	// Logical screen width and height.
	lEPutUint16(e.buf[0:2], uint16(e.width))
	lEPutUint16(e.buf[2:4], uint16(e.height))
	e.write(e.buf[:4])

	paddedSize := log2(len(e.palette))
	e.buf[0] = fColorTable | uint8(paddedSize)
	e.buf[1] = 0x00
	e.buf[2] = 0x00 // Pixel Aspect Ratio.
	e.write(e.buf[:3])
	e.encodeColorTable(paddedSize)

	if e.loopCount >= 0 {
		// add animation loop info
		e.buf[0] = 0x21
		e.buf[1] = 0xff
		e.buf[2] = 0x0b
		e.write(e.buf[:3])
		e.writeString("NETSCAPE2.0")
		e.buf[0] = 0x03
		e.buf[1] = 0x01
		lEPutUint16(e.buf[2:4], uint16(e.loopCount))
		e.buf[4] = 0x00
		e.write(e.buf[:5])
	}
}

func (e *gifEncoder) encodeColorTable(paddedSize int) {
	for i, c := range e.palette {
		e.globalColorTable[3*i+0] = c.R
		e.globalColorTable[3*i+1] = c.G
		e.globalColorTable[3*i+2] = c.B
	}
	n := 1 << (paddedSize + 1)
	if n > len(e.palette) {
		// Pad with black.
		clear(e.globalColorTable[3*len(e.palette) : 3*n])
	}
	e.write(e.globalColorTable[:3*n])
}

func (e *gifEncoder) writeString(s string) {
	if e.err != nil {
		return
	}
	_, e.err = io.WriteString(e.w, s)
}

func (e *gifEncoder) writeByte(b byte) {
	if e.err != nil {
		return
	}
	e.write([]byte{b})
}

func (e *gifEncoder) write(p []byte) {
	if e.err != nil {
		return
	}
	_, e.err = e.w.Write(p)
}

func lEPutUint16(b []byte, v uint16) {
	_ = b[1] // early bounds check to guarantee safety of writes below
	b[0] = byte(v)
	b[1] = byte(v >> 8)
}

func log2(x int) int {
	if x < 2 {
		return 0
	}
	return bits.Len(uint(x-1)) - 1
}

func (e *gifEncoder) writeImageBlock(pm *image.Paletted) {
	if e.err != nil {
		return
	}

	b := pm.Bounds()
	if b.Min.X < 0 || b.Max.X >= 1<<16 || b.Min.Y < 0 || b.Max.Y >= 1<<16 {
		e.err = errors.New("gif: image block is too large to encode")
		return
	}

	e.buf[0] = sExtension  // Extension Introducer.
	e.buf[1] = gcLabel     // Graphic Control Label.
	e.buf[2] = gcBlockSize // Block Size.
	e.buf[3] = 0x00 | disposal<<2
	lEPutUint16(e.buf[4:6], uint16(e.delay)) // Delay Time (1/100ths of a second)

	// Transparent color index.
	e.buf[6] = 0x00
	e.buf[7] = 0x00 // Block Terminator.
	e.write(e.buf[:8])
	e.buf[0] = sImageDescriptor
	lEPutUint16(e.buf[1:3], uint16(b.Min.X))
	lEPutUint16(e.buf[3:5], uint16(b.Min.Y))
	lEPutUint16(e.buf[5:7], uint16(b.Dx()))
	lEPutUint16(e.buf[7:9], uint16(b.Dy()))
	e.write(e.buf[:9])

	paddedSize := log2(len(pm.Palette)) // Size of Local Color Table: 2^(1+n).
	e.writeByte(0)                      // Use the global color table.

	litWidth := paddedSize + 1
	if litWidth < 2 {
		litWidth = 2
	}
	e.writeByte(uint8(litWidth)) // LZW Minimum Code Size.

	bw := blockWriter{e: e}
	bw.setup()
	lzww := lzw.NewWriter(bw, lzw.LSB, litWidth)
	if dx := b.Dx(); dx == pm.Stride {
		_, e.err = lzww.Write(pm.Pix[:dx*b.Dy()])
		if e.err != nil {
			_ = lzww.Close()
			return
		}
	} else {
		for i, y := 0, b.Min.Y; y < b.Max.Y; i, y = i+pm.Stride, y+1 {
			_, e.err = lzww.Write(pm.Pix[i : i+dx])
			if e.err != nil {
				_ = lzww.Close()
				return
			}
		}
	}
	_ = lzww.Close() // flush to bw
	bw.close()       // flush to e.w
}

type blockWriter struct {
	e *gifEncoder
}

func (b blockWriter) setup() {
	b.e.buf[0] = 0
}

func (b blockWriter) close() {
	// Write the block terminator (0x00), either by itself, or along with a
	// pending sub-block.
	if b.e.buf[0] == 0 {
		b.e.writeByte(0)
	} else {
		n := uint(b.e.buf[0])
		b.e.buf[n+1] = 0
		b.e.write(b.e.buf[:n+2])
	}
}

// blockWriter must be an io.Writer for lzw.NewWriter, but this is never
// actually called.
func (b blockWriter) Write(data []byte) (int, error) {
	for i, c := range data {
		if err := b.WriteByte(c); err != nil {
			return i, err
		}
	}
	return len(data), nil
}

func (b blockWriter) WriteByte(c byte) error {
	if b.e.err != nil {
		return b.e.err
	}

	// Append c to buffered sub-block.
	b.e.buf[0]++
	b.e.buf[b.e.buf[0]] = c
	if b.e.buf[0] < 255 {
		return nil
	}

	// Flush block
	b.e.write(b.e.buf[:256])
	b.e.buf[0] = 0
	return b.e.err
}
