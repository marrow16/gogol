package imaging

import (
	"bytes"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"image"
	"image/png"
	"io"
	"strings"
)

func PngEncode(w io.Writer, img image.Image, metadata [][2]string) error {
	metadata = append(metadata[:len(metadata):len(metadata)], [2]string{"Producer", "GoGoL"})
	pw := newPngEncoder(w, img, metadata)
	return pw.encode()
}

func newPngEncoder(w io.Writer, img image.Image, metadata [][2]string) *pngEncoder {
	var err error
	sanitizedMetadata := make([][2]string, 0, len(metadata))
	for _, meta := range metadata {
		if len(meta[0]) == 0 {
			err = errors.New("metadata keyword must not be empty")
			break
		}
		if len(meta[1]) == 0 {
			// simply ignore metadata with no value...
			continue
		}
		if len(meta[0]) > 79 {
			meta[0] = meta[0][:79]
		}
		if strings.IndexByte(meta[0], 0) >= 0 {
			err = errors.New("metadata keyword has 0 byte terminator")
			break
		}
		sanitizedMetadata = append(sanitizedMetadata, meta)
	}
	return &pngEncoder{
		w:        w,
		tail:     make([]byte, 0, pngIENDSize),
		img:      img,
		metadata: sanitizedMetadata,
		err:      err,
	}
}

type pngEncoder struct {
	w        io.Writer
	tail     []byte
	img      image.Image
	metadata [][2]string
	err      error
}

const pngIENDSize = 12

var (
	pngIEND = []byte{
		0x00, 0x00, 0x00, 0x00,
		'I', 'E', 'N', 'D',
		0xae, 0x42, 0x60, 0x82}
	pngITXT = []byte{'i', 'T', 'X', 't'}
)

func (e *pngEncoder) encode() error {
	if e.err != nil {
		return e.err
	}
	if len(e.metadata) == 0 {
		// no metadata - so just write png as normal...
		return png.Encode(e.w, e.img)
	}
	// encode the actual png (exc. IEND)...
	if err := png.Encode(e, e.img); err != nil {
		return err
	}
	// write the metadata...
	for _, meta := range e.metadata {
		e.writePngITXt(meta[0], meta[1])
	}
	if e.err != nil {
		return e.err
	}
	// write the IEND...
	_, e.err = e.w.Write(pngIEND)
	return e.err
}

func (e *pngEncoder) Write(p []byte) (int, error) {
	n := len(p)
	data := append(e.tail, p...)
	if idx := bytes.Index(data, pngIEND); idx >= 0 {
		if _, e.err = e.w.Write(data[:idx]); e.err != nil {
			return 0, e.err
		}
		e.tail = append(e.tail[:0], pngIEND...)
		return n, nil
	}
	// retain enough bytes to detect an IEND split across Write calls...
	keep := min(len(data), pngIENDSize-1)
	write := len(data) - keep
	if write > 0 {
		if _, e.err = e.w.Write(data[:write]); e.err != nil {
			return 0, e.err
		}
	}
	e.tail = append(e.tail[:0], data[write:]...)
	return n, nil
}

func (e *pngEncoder) writePngITXt(key, value string) {
	if e.err != nil {
		return
	}
	data := make([]byte, 0, len(key)+len(value)+5)
	// append the keyword...
	data = append(data, key...)
	// append keyword terminator, compression flag/method, empty language tag and empty translated keyword...
	data = append(data, 0, 0, 0, 0, 0)
	data = append(data, value...)
	// write the iTXt length...
	var length [4]byte
	binary.BigEndian.PutUint32(length[:], uint32(len(data)))
	if _, e.err = e.w.Write(length[:]); e.err != nil {
		return
	}
	// write the chunk type...
	if _, e.err = e.w.Write(pngITXT); e.err != nil {
		return
	}
	// write the actual data...
	if _, e.err = e.w.Write(data); e.err != nil {
		return
	}
	// calc checksum...
	crc := crc32.NewIEEE()
	if _, e.err = crc.Write(pngITXT); e.err != nil {
		return
	}
	if _, e.err = crc.Write(data); e.err != nil {
		return
	}
	var checksum [4]byte
	binary.BigEndian.PutUint32(checksum[:], crc.Sum32())
	// write checksum...
	_, e.err = e.w.Write(checksum[:])
}
