package maxmind

import (
	"net"
	"os"
	"path/filepath"

	"github.com/oschwald/geoip2-golang"

	"github.com/NethermindEth/aztec-p2p-explorer/core/types"
)

const (
	asnFileName  = "GeoLite2-ASN.mmdb"
	cityFileName = "GeoLite2-City.mmdb"
)

type Client interface {
	GetAddrInfo(ip string) (*types.GeoInfo, error)
	Close() error
}

type MaxMindClient struct {
	asnReader  *geoip2.Reader
	cityReader *geoip2.Reader
}

func NewClient(maxmindDir string) (*MaxMindClient, error) {
	asnPath := filepath.Join(maxmindDir, asnFileName)
	cityPath := filepath.Join(maxmindDir, cityFileName)

	if _, err := os.Stat(asnPath); os.IsNotExist(err) {
		return nil, err
	}

	if _, err := os.Stat(cityPath); os.IsNotExist(err) {
		return nil, err
	}

	asnReader, err := geoip2.Open(asnPath)
	if err != nil {
		return nil, err
	}

	cityReader, err := geoip2.Open(cityPath)
	if err != nil {
		return nil, err
	}

	return &MaxMindClient{
		asnReader:  asnReader,
		cityReader: cityReader,
	}, nil
}

func (c *MaxMindClient) GetAddrInfo(ip string) (*types.GeoInfo, error) {
	resIP := net.ParseIP(ip)
	asn, err := c.asnReader.ASN(resIP)
	if err != nil {
		return nil, err
	}

	city, err := c.cityReader.City(resIP)
	if err != nil {
		return nil, err
	}

	return &types.GeoInfo{
		ASOrganization: asn.AutonomousSystemOrganization,
		ASNumber:       asn.AutonomousSystemNumber,
		City:           city.City.Names["en"],
		Country:        city.Country.Names["en"],
		CountryISO:     city.Country.IsoCode,
		Continent:      city.Continent.Names["en"],
		ContinentCode:  city.Continent.Code,
		Latitude:       city.Location.Latitude,
		Longitude:      city.Location.Longitude,
	}, nil
}

func (c *MaxMindClient) Close() error {
	err := c.asnReader.Close()
	if err != nil {
		return err
	}

	err = c.cityReader.Close()
	if err != nil {
		return err
	}

	return nil
}

type mockClient struct{}

func NewMockClient() *mockClient {
	return &mockClient{}
}

func (c *mockClient) GetAddrInfo(ip string) (*types.GeoInfo, error) {
	return &types.GeoInfo{
		ASOrganization: "Deutsche Telekom AG",
		ASNumber:       3320,
		City:           "Berlin",
		Country:        "Germany",
		CountryISO:     "DE",
		Continent:      "Europe",
		ContinentCode:  "EU",
		Latitude:       52.52,
		Longitude:      13.405,
	}, nil
}

func (c *mockClient) Close() error {
	return nil
}
