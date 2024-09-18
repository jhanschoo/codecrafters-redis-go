// Package respreader provides a Parser for the RESP protocol.
package rdbreader

import (
	"bufio"
	"encoding/binary"
	"errors"
	"log"
	"strconv"
	"time"
)

const (
	version11MagicString = "REDIS0011"
)

var (
	millisecondsInSecond     = uint64(time.Second / time.Millisecond)
	nanosecondsInMillisecond = uint64(time.Millisecond / time.Nanosecond)
)

var (
	ErrorUnsupportedRDBVersion     = errors.New("unsupported RDB version")
	ErrorUnknownOpcode             = errors.New("unknown opcode")
	ErrorUnsupportedOpcode         = errors.New("unsupported opcode")
	ErrorUnsupportedValueType      = errors.New("unsupported value type")
	ErrorMissingEOF                = errors.New("missing EOF")
	ErrorInsufficientBytesRead     = errors.New("insufficient bytes read")
	ErrorUnsupportedLengthEncoding = errors.New("unsupported length encoding")
	ErrorUnsupportedStringEncoding = errors.New("unsupported string encoding")
	ErrorIncorrectDbSizes          = errors.New("incomplete db sizes")
	ErrorDanglingExpiry            = errors.New("dangling expiry")
)

type ValueData struct {
	Value     string
	ExpiresAt time.Time
}

type readerState struct {
	br               *bufio.Reader
	aux              map[string]string
	dbs              map[int64]map[string]ValueData
	currentDb        int64
	currentKvExpiry  time.Time
	entriesLeft      int64
	expiryFieldsLeft int64
}

func ReadRDB(br *bufio.Reader) (map[int64]map[string]ValueData, error) {
	st := &readerState{
		br,
		make(map[string]string),
		make(map[int64]map[string]ValueData),
		-1,
		time.Time{},
		-1,
		-1,
	}
	st, err := st.readRDB()
	if err != nil {
		return nil, err
	}
	return st.dbs, nil
}

func (st *readerState) readMagicString() (*readerState, error) {
	log.Println("readMagicString")
	magicCandidate := make([]byte, len(version11MagicString))
	n, err := st.br.Read(magicCandidate)
	if err != nil {
		return nil, err
	}
	if n != len(version11MagicString) {
		return nil, ErrorInsufficientBytesRead
	}
	if string(magicCandidate) != version11MagicString {
		return nil, ErrorUnsupportedRDBVersion
	}
	log.Printf("readMagicString: success, read magic string: %s\n", version11MagicString)
	return st, nil
}

func (st *readerState) readAuxField() (*readerState, error) {
	log.Println("readAuxField")
	key, err := readString(st.br)
	if err != nil {
		return nil, err
	}
	value, err := readString(st.br)
	if err != nil {
		return nil, err
	}
	st.aux[key] = value
	log.Printf("readAuxField: success, read key: %s, value: %s\n", key, value)
	return st, nil
}

func (st *readerState) readResizeDb() (*readerState, error) {
	log.Println("readResizeDb")
	if st.entriesLeft > 0 || st.expiryFieldsLeft > 0 {
		return nil, ErrorIncorrectDbSizes
	}
	var err error
	st.entriesLeft, err = readLength(st.br)
	if err != nil {
		return nil, err
	}
	st.expiryFieldsLeft, err = readLength(st.br)
	if err != nil {
		return nil, err
	}
	st.dbs[st.currentDb] = make(map[string]ValueData, st.entriesLeft)
	log.Printf("readResizeDb: success, expecting db %d to contain %d entries and %d expiry fields\n", st.currentDb, st.entriesLeft, st.expiryFieldsLeft)
	return st, nil
}

func (st *readerState) readExpireTime() (*readerState, error) {
	log.Println("readExpireTime")
	if st.expiryFieldsLeft == 0 {
		return nil, ErrorIncorrectDbSizes
	}
	st.expiryFieldsLeft--
	bs, err := readBytes(st.br, 4)
	if err != nil {
		return nil, err
	}
	st.currentKvExpiry = time.Unix(int64(binary.LittleEndian.Uint32(bs)), 0)
	log.Printf("readExpireTime: success, expecting next entry to have expiry time: %s\n", st.currentKvExpiry)
	return st, nil
}

func (st *readerState) readExpireTimeMs() (*readerState, error) {
	log.Println("readExpireTimeMs")
	if st.expiryFieldsLeft == 0 {
		return nil, ErrorIncorrectDbSizes
	}
	st.expiryFieldsLeft--
	bs, err := readBytes(st.br, 8)
	if err != nil {
		return nil, err
	}
	ms := binary.LittleEndian.Uint64(bs)
	sec := int64(ms / millisecondsInSecond)
	nsec := int64((ms % millisecondsInSecond) * nanosecondsInMillisecond)
	st.currentKvExpiry = time.Unix(sec, nsec)
	log.Printf("readExpireTimeMs: success, expecting next entry to have expiry time: %s\n", st.currentKvExpiry)
	return st, nil
}

func (st *readerState) readSelectDb() (*readerState, error) {
	log.Println("readSelectDb")
	if st.entriesLeft > 0 || st.expiryFieldsLeft > 0 {
		return nil, ErrorIncorrectDbSizes
	}
	db, err := readLength(st.br)
	if err != nil {
		return nil, err
	}
	st.dbs[db] = make(map[string]ValueData)
	st.currentDb = db
	log.Printf("readSelectDb: success, selected db %d\n", db)
	return st, nil
}

func (st *readerState) readEof() (*readerState, error) {
	log.Println("readEof")
	if st.entriesLeft > 0 || st.expiryFieldsLeft > 0 {
		return nil, ErrorIncorrectDbSizes
	}
	if !st.currentKvExpiry.IsZero() {
		return nil, ErrorDanglingExpiry
	}
	log.Println("readEof: success, read EOF")
	return st, nil
}

func (st *readerState) readKv(typeByte byte) (*readerState, error) {
	log.Println("readKv")
	if st.entriesLeft == 0 {
		return nil, ErrorIncorrectDbSizes
	}
	st.entriesLeft--
	k, err := readString(st.br)
	if err != nil {
		return nil, err
	}
	switch typeByte {
	case 0: // string
		v, err := readString(st.br)
		if err != nil {
			return nil, err
		}
		st.dbs[st.currentDb][k] = ValueData{v, st.currentKvExpiry}
		log.Printf("readKv: success, read key: %s, value: %s, with expiry %s\n", k, v, st.currentKvExpiry)
	default:
		return nil, ErrorUnsupportedValueType
	}
	st.currentKvExpiry = time.Time{}
	return st, nil
}

func (st *readerState) readRDB() (*readerState, error) {
	st, err := st.readMagicString()
	if err != nil {
		return nil, err
	}
	for {
		opcode, err := st.br.ReadByte()
		if err != nil {
			return nil, err
		}
		switch opcode {
		// 0xFA: AUX
		case 0xFA:
			if _, err = st.readAuxField(); err != nil {
				return nil, err
			}
		// 0xFB: RESIZEDB
		case 0xFB:
			if _, err = st.readResizeDb(); err != nil {
				return nil, err
			}
		// 0xFC: EXPIRETIMEMS
		case 0xFC:
			if _, err = st.readExpireTimeMs(); err != nil {
				return nil, err
			}
		// 0xFD: EXPIRETIME
		case 0xFD:
			if _, err = st.readExpireTime(); err != nil {
				return nil, err
			}
		// 0xFE: SELECTDB
		case 0xFE:
			if _, err = st.readSelectDb(); err != nil {
				return nil, err
			}
		// 0xFF: EOF
		case 0xFF:
			if _, err = st.readEof(); err != nil {
				return nil, err
			} else { // explicit else to visually signal distinct handling of case
				return st, nil
			}
		default:
			if _, err = st.readKv(opcode); err != nil {
				return nil, err
			}
		}
	}
}

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
