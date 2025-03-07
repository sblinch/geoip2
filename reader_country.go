package geoip2

import (
	"errors"
	"io/ioutil"
	"net"
	"strconv"
)

type CountryReader struct {
	*reader
}

func (r *CountryReader) Lookup(ip net.IP) (*CountryResult, error) {
	offset, err := r.getOffset(ip)
	if err != nil {
		return nil, err
	}
	dataType, size, offset, err := readControl(r.decoderBuffer, offset)
	if err != nil {
		return nil, err
	}
	if dataType != dataTypeMap {
		return nil, errors.New("invalid Country type: " + strconv.Itoa(int(dataType)))
	}
	var key []byte
	result := &CountryResult{}
	for i := uint(0); i < size; i++ {
		key, offset, err = readMapKey(r.decoderBuffer, offset)
		if err != nil {
			return nil, err
		}
		switch b2s(key) {
		case "continent":
			offset, err = readContinent(&result.Continent, r.decoderBuffer, offset)
			if err != nil {
				return nil, err
			}
		case "country":
			offset, err = readCountry(&result.Country, r.decoderBuffer, offset)
			if err != nil {
				return nil, err
			}
		case "registered_country":
			offset, err = readCountry(&result.RegisteredCountry, r.decoderBuffer, offset)
			if err != nil {
				return nil, err
			}
		case "represented_country":
			offset, err = readCountry(&result.RepresentedCountry, r.decoderBuffer, offset)
			if err != nil {
				return nil, err
			}
		case "traits":
			offset, err = readTraits(&result.Traits, r.decoderBuffer, offset)
			if err != nil {
				return nil, err
			}
		default:
			return nil, errors.New("unknown Country response key: " + string(key) + ", type: " + strconv.Itoa(int(dataType)))
		}
	}
	return result, nil
}

// NewCountryReaderWithType creates a new CountryReader that accepts MMDB files with a custom database type. Note that
// CountryReader only implements the fields provided by MaxMind Geo*-Country databases, and will ignore other fields.
// It is up to the developer to ensure that the database provides a compatible selection of fields.
func NewCountryReaderWithType(buffer []byte, expectedTypes ...string) (*CountryReader, error) {
	reader, err := newReader(buffer)
	if err != nil {
		return nil, err
	}
	if !isExpectedDatabaseType(reader.metadata.DatabaseType, expectedTypes...) {
		return nil, errors.New("wrong MaxMind DB Country type: " + reader.metadata.DatabaseType)
	}
	return &CountryReader{
		reader: reader,
	}, nil
}

func NewCountryReader(buffer []byte) (*CountryReader, error) {
	return NewCountryReaderWithType(buffer, "GeoIP2-Country", "GeoLite2-Country", "Geoacumen-Country", "DBIP-Country", "DBIP-Country-Lite")
}

func NewCountryReaderFromFile(filename string) (*CountryReader, error) {
	buffer, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return NewCountryReader(buffer)
}
