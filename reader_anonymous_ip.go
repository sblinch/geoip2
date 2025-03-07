package geoip2

import (
	"errors"
	"io/ioutil"
	"net"
	"strconv"
)

type AnonymousIPReader struct {
	*reader
}

func (r *AnonymousIPReader) Lookup(ip net.IP) (*AnonymousIP, error) {
	offset, err := r.getOffset(ip)
	if err != nil {
		return nil, err
	}
	dataType, size, offset, err := readControl(r.decoderBuffer, offset)
	if err != nil {
		return nil, err
	}
	result := &AnonymousIP{}
	switch dataType {
	case dataTypeMap:
		_, err = readAnonymousIPMap(result, r.decoderBuffer, size, offset)
		if err != nil {
			return nil, err
		}
	case dataTypePointer:
		pointer, _, err := readPointer(r.decoderBuffer, size, offset)
		if err != nil {
			return nil, err
		}
		dataType, size, offset, err := readControl(r.decoderBuffer, pointer)
		if err != nil {
			return nil, err
		}
		if dataType != dataTypeMap {
			return nil, errors.New("invalid Anonymous-IP pointer type: " + strconv.Itoa(int(dataType)))
		}
		_, err = readAnonymousIPMap(result, r.decoderBuffer, size, offset)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("invalid Anonymous-IP type: " + strconv.Itoa(int(dataType)))
	}
	return result, nil
}

// NewAnonymousIPReaderWithType creates a new AnonymousIPReader that accepts MMDB files with a custom database type.
// Note that AnonymousIPReader only implements the fields provided by MaxMind GeoIP2-Anonymous-IP databases, and will
// ignore other fields. It is up to the developer to ensure that the database provides a compatible selection of fields.
func NewAnonymousIPReaderWithType(buffer []byte, expectedTypes ...string) (*AnonymousIPReader, error) {
	reader, err := newReader(buffer)
	if err != nil {
		return nil, err
	}
	if !isExpectedDatabaseType(reader.metadata.DatabaseType, expectedTypes...) {
		return nil, errors.New("wrong MaxMind DB Anonymous-IP type: " + reader.metadata.DatabaseType)
	}
	return &AnonymousIPReader{
		reader: reader,
	}, nil
}

func NewAnonymousIPReader(buffer []byte) (*AnonymousIPReader, error) {
	return NewAnonymousIPReaderWithType(buffer, "GeoIP2-Anonymous-IP")
}

func NewAnonymousIPReaderFromFile(filename string) (*AnonymousIPReader, error) {
	buffer, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return NewAnonymousIPReader(buffer)
}
