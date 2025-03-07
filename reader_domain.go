package geoip2

import (
	"errors"
	"io/ioutil"
	"net"
	"strconv"
)

type DomainReader struct {
	*reader
}

func (r *DomainReader) Lookup(ip net.IP) (string, error) {
	offset, err := r.getOffset(ip)
	if err != nil {
		return "", err
	}
	dataType, size, offset, err := readControl(r.decoderBuffer, offset)
	if err != nil {
		return "", err
	}
	result := &Domain{}
	switch dataType {
	case dataTypeMap:
		_, err = readDomainMap(result, r.decoderBuffer, size, offset)
		if err != nil {
			return "", err
		}
	case dataTypePointer:
		pointer, _, err := readPointer(r.decoderBuffer, size, offset)
		if err != nil {
			return "", err
		}
		dataType, size, offset, err := readControl(r.decoderBuffer, pointer)
		if err != nil {
			return "", err
		}
		if dataType != dataTypeMap {
			return "", errors.New("invalid Domain pointer type: " + strconv.Itoa(int(dataType)))
		}
		_, err = readDomainMap(result, r.decoderBuffer, size, offset)
		if err != nil {
			return "", err
		}
	default:
		return "", errors.New("invalid Domain type: " + strconv.Itoa(int(dataType)))
	}
	return result.Domain, nil
}

// NewDomainReaderWithType creates a new DomainReader that accepts MMDB files with a custom database type. Note that
// DomainReader only implements the fields provided by MaxMind GeoIP2-Domain databases, and will ignore other fields.
// It is up to the developer to ensure that the database provides a compatible selection of fields.
func NewDomainReaderWithType(buffer []byte, expectedTypes ...string) (*DomainReader, error) {
	reader, err := newReader(buffer)
	if err != nil {
		return nil, err
	}
	if !isExpectedDatabaseType(reader.metadata.DatabaseType, expectedTypes...) {
		return nil, errors.New("wrong MaxMind DB Domain type: " + reader.metadata.DatabaseType)
	}
	return &DomainReader{
		reader: reader,
	}, nil
}

func NewDomainReader(buffer []byte) (*DomainReader, error) {
	return NewDomainReaderWithType(buffer, "GeoIP2-Domain")
}

func NewDomainReaderFromFile(filename string) (*DomainReader, error) {
	buffer, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return NewDomainReader(buffer)
}
