// Package respreader provides a globally shared Reader to initialize the global db of the application in an unsafe manner from a *bufio.Reader.
package rdbreader

import (
	"bufio"
	"encoding/binary"
	"errors"
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
	ErrorUnsupportedDbIndex        = errors.New("unsupported db index")
)

type ValueData struct {
	Value     string
	ExpiresAt time.Time
}

type ResetStateFunc = func(sizeHint int64)
type UnsafeKvSetterFunc = func(key, value string, expiresAt time.Time)

type readerState struct {
	br               *bufio.Reader
	aux              map[string]string
	currentKvExpiry  time.Time
	entriesLeft      int64
	expiryFieldsLeft int64
	stateResetter    ResetStateFunc
	unsafeKvSetter   UnsafeKvSetterFunc
}

var st readerState

func ReadRDBToState(br *bufio.Reader, stateResetter ResetStateFunc, unsafeKvSetter UnsafeKvSetterFunc) error {
	stateResetter(0)
	st = readerState{
		br,
		make(map[string]string),
		time.Time{},
		-1,
		-1,
		stateResetter,
		unsafeKvSetter,
	}
	return st.readRDB()
}

func (st *readerState) readMagicString() error {
	magicCandidate := make([]byte, len(version11MagicString))
	n, err := st.br.Read(magicCandidate)
	if err != nil {
		return err
	}
	if n != len(version11MagicString) {
		return ErrorInsufficientBytesRead
	}
	if string(magicCandidate) != version11MagicString {
		return ErrorUnsupportedRDBVersion
	}
	return nil
}

func (st *readerState) readAuxField() error {
	key, err := readString(st.br)
	if err != nil {
		return err
	}
	value, err := readString(st.br)
	if err != nil {
		return err
	}
	st.aux[key] = value
	return nil
}

func (st *readerState) readResizeDb() error {
	if st.entriesLeft > 0 || st.expiryFieldsLeft > 0 {
		return ErrorIncorrectDbSizes
	}
	var err error
	st.entriesLeft, err = readLength(st.br)
	if err != nil {
		return err
	}
	st.expiryFieldsLeft, err = readLength(st.br)
	if err != nil {
		return err
	}
	st.stateResetter(st.entriesLeft)
	return nil
}

func (st *readerState) readExpireTime() error {
	if st.expiryFieldsLeft == 0 {
		return ErrorIncorrectDbSizes
	}
	if !st.currentKvExpiry.IsZero() {
		return ErrorDanglingExpiry
	}
	st.expiryFieldsLeft--
	bs, err := readBytes(st.br, 4)
	if err != nil {
		return err
	}
	st.currentKvExpiry = time.Unix(int64(binary.LittleEndian.Uint32(bs)), 0)
	return nil
}

func (st *readerState) readExpireTimeMs() error {
	if st.expiryFieldsLeft == 0 {
		return ErrorIncorrectDbSizes
	}
	if !st.currentKvExpiry.IsZero() {
		return ErrorDanglingExpiry
	}
	st.expiryFieldsLeft--
	bs, err := readBytes(st.br, 8)
	if err != nil {
		return err
	}
	ms := binary.LittleEndian.Uint64(bs)
	sec := int64(ms / millisecondsInSecond)
	nsec := int64((ms % millisecondsInSecond) * nanosecondsInMillisecond)
	st.currentKvExpiry = time.Unix(sec, nsec)
	return nil
}

func (st *readerState) readSelectDb() error {
	if st.entriesLeft > 0 || st.expiryFieldsLeft > 0 {
		return ErrorIncorrectDbSizes
	}
	db, err := readLength(st.br)
	if err != nil {
		return err
	}
	if db != 0 {
		return ErrorUnsupportedDbIndex
	}
	return nil
}

func (st *readerState) readEof() error {
	if st.entriesLeft > 0 || st.expiryFieldsLeft > 0 {
		return ErrorIncorrectDbSizes
	}
	if !st.currentKvExpiry.IsZero() {
		return ErrorDanglingExpiry
	}
	return nil
}

func (st *readerState) readKv(typeByte byte) error {
	if st.entriesLeft == 0 {
		return ErrorIncorrectDbSizes
	}
	st.entriesLeft--
	k, err := readString(st.br)
	if err != nil {
		return err
	}
	switch typeByte {
	case 0: // string
		v, err := readString(st.br)
		if err != nil {
			return err
		}
		st.unsafeKvSetter(k, v, st.currentKvExpiry)
	default:
		return ErrorUnsupportedValueType
	}

	// reset value-specific state
	st.currentKvExpiry = time.Time{}
	return nil
}

func (st *readerState) readRDB() error {
	if err := st.readMagicString(); err != nil {
		return err
	}
	for {
		opcode, err := st.br.ReadByte()
		if err != nil {
			return err
		}
		switch opcode {
		// 0xFA: AUX
		case 0xFA:
			if err = st.readAuxField(); err != nil {
				return err
			}
		// 0xFB: RESIZEDB
		case 0xFB:
			if err = st.readResizeDb(); err != nil {
				return err
			}
		// 0xFC: EXPIRETIMEMS
		case 0xFC:
			if err = st.readExpireTimeMs(); err != nil {
				return err
			}
		// 0xFD: EXPIRETIME
		case 0xFD:
			if err = st.readExpireTime(); err != nil {
				return err
			}
		// 0xFE: SELECTDB
		case 0xFE:
			if err = st.readSelectDb(); err != nil {
				return err
			}
		// 0xFF: EOF
		case 0xFF:
			if err = st.readEof(); err != nil {
				return err
			} else { // explicit else to visually signal distinct handling of case
				return nil
			}
		default:
			if err = st.readKv(opcode); err != nil {
				return err
			}
		}
	}
}
