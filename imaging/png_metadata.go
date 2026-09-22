package imaging

import (
	"bytes"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"io"
)

func PngMetadata(r io.Reader) ([][2]string, error) {
	var signature [8]byte
	if _, err := io.ReadFull(r, signature[:]); err != nil {
		return nil, err
	}
	if !bytes.Equal(signature[:], pngSignature) {
		return nil, errors.New("invalid PNG signature")
	}
	var metadata [][2]string
	var header [8]byte
	for {
		// read the chunk header...
		if _, err := io.ReadFull(r, header[:]); err != nil {
			return nil, err
		}
		length := binary.BigEndian.Uint32(header[:4])
		chunkType := header[4:]

		// read text metadata chunk...
		if bytes.Equal(chunkType, pngITXT) || bytes.Equal(chunkType, pngTEXT) {
			data := make([]byte, length)
			if _, err := io.ReadFull(r, data); err != nil {
				return nil, err
			}
			if err := validatePngChunk(r, chunkType, data); err != nil {
				return nil, err
			}
			var meta [2]string
			var err error
			if bytes.Equal(chunkType, pngITXT) {
				meta, err = parsePngITXt(data)
			} else {
				meta, err = parsePngText(data)
			}
			if err != nil {
				return nil, err
			}
			metadata = append(metadata, meta)
			continue
		}
		// stop at IEND...
		if bytes.Equal(chunkType, pngIENDType) {
			if length != 0 {
				return nil, errors.New("invalid PNG IEND")
			}
			if err := validatePngChunk(r, chunkType, nil); err != nil {
				return nil, err
			}
			return metadata, nil
		}
		// skip unwanted chunk data and CRC...
		if _, err := io.CopyN(io.Discard, r, int64(length)+4); err != nil {
			return nil, err
		}
	}
}

var (
	pngSignature = []byte{
		0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}
	pngIENDType = []byte{'I', 'E', 'N', 'D'}
	pngTEXT     = []byte{'t', 'E', 'X', 't'}
)

func validatePngChunk(r io.Reader, chunkType, data []byte) error {
	var checksum [4]byte
	if _, err := io.ReadFull(r, checksum[:]); err != nil {
		return err
	}
	crc := crc32.NewIEEE()
	_, _ = crc.Write(chunkType)
	_, _ = crc.Write(data)
	if binary.BigEndian.Uint32(checksum[:]) != crc.Sum32() {
		return errors.New("invalid PNG chunk checksum")
	}
	return nil
}

func parsePngITXt(data []byte) ([2]string, error) {
	var result [2]string
	// read keyword...
	idx := bytes.IndexByte(data, 0)
	if idx <= 0 || idx > 79 {
		return result, errors.New("invalid PNG iTXt keyword")
	}
	result[0] = string(data[:idx])
	data = data[idx+1:]
	// read compression flag and method...
	if len(data) < 2 {
		return result, errors.New("invalid PNG iTXt chunk")
	}
	compressed := data[0]
	method := data[1]
	data = data[2:]
	if compressed != 0 {
		return result, errors.New("compressed PNG iTXt is not supported")
	}
	if method != 0 {
		return result, errors.New("invalid PNG iTXt compression method")
	}
	// skip language tag...
	idx = bytes.IndexByte(data, 0)
	if idx < 0 {
		return result, errors.New("invalid PNG iTXt language tag")
	}
	data = data[idx+1:]
	// skip translated keyword...
	idx = bytes.IndexByte(data, 0)
	if idx < 0 {
		return result, errors.New("invalid PNG iTXt translated keyword")
	}
	data = data[idx+1:]
	// remaining data is UTF-8 text...
	result[1] = string(data)
	return result, nil
}

func parsePngText(data []byte) ([2]string, error) {
	var result [2]string
	idx := bytes.IndexByte(data, 0)
	if idx <= 0 || idx > 79 {
		return result, errors.New("invalid PNG tEXt keyword")
	}
	result[0] = latin1String(data[:idx])
	result[1] = latin1String(data[idx+1:])
	return result, nil
}

func latin1String(data []byte) string {
	result := make([]rune, len(data))
	for i, b := range data {
		result[i] = rune(b)
	}
	return string(result)
}
