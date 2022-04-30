package smarttrak

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
)

type Header struct {
	Name string
}

func ReadHeader(r io.Reader) (*Header, error) {
	var magic [4]byte
	_, err := r.Read(magic[:])
	if err != nil {
		return nil, err
	}

	// skip 16 bytes "CTravelTrakCEDoc "
	_, err = readExact(r, 16)
	if err != nil {
		return nil, err
	}

	name, err := readString(r)
	if err != nil {
		return nil, err
	}

	// skip 38 bytes
	_, err = readExact(r, 38)
	if err != nil {
		return nil, err
	}

	suitType, err := readString(r)
	if err != nil {
		return nil, err
	}
	log.Println("Suit type:", suitType)

	// skip 2 bytes
	_, err = readExact(r, 2)
	if err != nil {
		return nil, err
	}

	weather, err := readString(r)
	if err != nil {
		return nil, err
	}
	log.Println("Weather:", weather)

	// skip 27 bytes
	_, err = readExact(r, 27)
	if err != nil {
		return nil, err
	}

	return &Header{
		Name: name,
	}, nil
}

func readString(r io.Reader) (string, error) {
	header, err := readExact(r, 4)
	if err != nil {
		return "", err
	}

	if !bytes.Equal(header[:3], []byte{0xFF, 0xFE, 0xFF}) {
		return "", fmt.Errorf("unexpected string header")
	}

	length := int(header[3])
	var runes []rune
	for i := 0; i < length; i++ {
		char, err := readUint16(r)
		if err != nil {
			return "", err
		}
		runes = append(runes, rune(char))
	}

	return string(runes), nil
}

func readUint16(r io.Reader) (uint16, error) {
	data, err := readExact(r, 2)
	if err != nil {
		return 0, err
	}

	return binary.LittleEndian.Uint16(data), nil
}

func readExact(r io.Reader, length int) ([]byte, error) {
	data := make([]byte, length)

	var read int
	for read < length {
		n, err := r.Read(data[read:])
		if err != nil {
			return nil, err
		}
		if n == 0 {
			return nil, io.EOF
		}
		read += n
	}

	return data, nil
}
