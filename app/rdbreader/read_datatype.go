package rdbreader

import (
	"bufio"
	"encoding/binary"
	"errors"
	"strconv"
)

func readBytes(br *bufio.Reader, length int) ([]byte, error) {
	bs := make([]byte, length)
	n, err := br.Read(bs)
	if err != nil {
		return nil, err
	}
	if n != length {
		return nil, ErrorInsufficientBytesRead
	}
	return bs, nil
}

func readLength(br *bufio.Reader) (int64, error) {
	b, err := br.ReadByte()
	if err != nil {
		return 0, err
	}
	switch b >> 6 {
	case 0b00:
		return int64(b & 0b00111111), nil
	case 0b01:
		b2, err := br.ReadByte()
		if err != nil {
			return 0, err
		}
		return int64(b&0b00111111)<<8 + int64(b2), nil
	case 0b10:
		bs, err := readBytes(br, 4)
		if err != nil {
			return 0, err
		}
		return int64(int32(binary.BigEndian.Uint32(bs))), nil
	default:
		br.UnreadByte()
		return 0, ErrorUnsupportedLengthEncoding
	}
}

func readString(br *bufio.Reader) (string, error) {
	length, err := readLength(br)
	if errors.Is(err, ErrorUnsupportedLengthEncoding) {
		b, err := br.ReadByte()
		if err != nil {
			return "", err
		}
		switch b {
		case 0xC0:
			b1, err := br.ReadByte()
			if err != nil {
				return "", err
			}
			return strconv.FormatInt(int64(int8(b1)), 10), nil
		case 0xC1:
			bs, err := readBytes(br, 2)
			if err != nil {
				return "", err
			}
			return strconv.FormatInt(int64(int16(binary.LittleEndian.Uint16(bs))), 10), nil
		case 0xC2:
			bs, err := readBytes(br, 4)
			if err != nil {
				return "", err
			}
			return strconv.FormatInt(int64(int32(binary.LittleEndian.Uint32(bs))), 10), nil
		default:
			return "", ErrorUnsupportedStringEncoding
		}
	}
	bs, err := readBytes(br, int(length))
	if err != nil {
		return "", err
	}
	return string(bs), nil
}
