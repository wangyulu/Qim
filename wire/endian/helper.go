package endian

import (
	"encoding/binary"
	"io"
)

var Default = binary.LittleEndian

func ReadUint8(r io.Reader) (uint8, error) {
	var bytes = make([]byte, 1)
	if _, err := io.ReadFull(r, bytes); err != nil {
		return 0, err
	}

	return uint8(bytes[0]), nil
}

func ReadUint16(r io.Reader) (uint16, error) {
	var bytes = make([]byte, 2)
	if _, err := io.ReadFull(r, bytes); err != nil {
		return 0, err
	}

	return Default.Uint16(bytes), nil
}

func ReadUint32(r io.Reader) (uint32, error) {
	var bytes = make([]byte, 4)
	if _, err := io.ReadFull(r, bytes); err != nil {
		return 0, err
	}

	return Default.Uint32(bytes), nil
}

func ReadBytes(r io.Reader) ([]byte, error) {
	bufLen, err := ReadUint32(r)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, bufLen)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}

	return buf, nil
}

func ReadFixedBytes(len int, r io.Reader) ([]byte, error) {
	buf := make([]byte, len)

	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}

	return buf, nil
}

func WriteUint8(w io.Writer, val uint8) error {
	buf := []byte{byte(val)}

	if _, err := w.Write(buf); err != nil {
		return err
	}

	return nil
}

func WriteUint16(w io.Writer, val uint16) error {
	buf := make([]byte, 2)

	Default.PutUint16(buf, val)

	if _, err := w.Write(buf); err != nil {
		return err
	}

	return nil
}

func WriteUint32(w io.Writer, val uint32) error {
	buf := make([]byte, 4)

	Default.PutUint32(buf, val)

	if _, err := w.Write(buf); err != nil {
		return err
	}

	return nil
}

func WriteBytes(w io.Writer, buf []byte) error {
	bufLen := len(buf)

	err := WriteUint32(w, uint32(bufLen))
	if err != nil {
		return err
	}

	_, err = w.Write(buf)
	if err != nil {
		return err
	}

	return nil
}

func WriteShortBytes(w io.Writer, buf []byte) error {
	bufLen := len(buf)

	if err := WriteUint16(w, uint16(bufLen)); err != nil {
		return err
	}

	_, err := w.Write(buf)

	return err
}

func ReadShortBytes(r io.Reader) ([]byte, error) {
	bufLen, err := ReadUint16(r)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, bufLen)

	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}

	return buf, err
}

func ReadShortString(r io.Reader) (string, error) {
	buf, err := ReadShortBytes(r)
	if err != nil {
		return "", err
	}

	return string(buf), nil
}
