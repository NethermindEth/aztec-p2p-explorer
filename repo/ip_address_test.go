package repo

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	coretypes "github.com/NethermindEth/aztec-p2p-explorer/core/types"
	"github.com/NethermindEth/aztec-p2p-explorer/database/models"
)

func TestCreateIPAddress(t *testing.T) {
	t.Parallel()

	repo, tearDown := setupTestRepo(t)
	t.Cleanup(tearDown)

	ip := &coretypes.IPInfo{
		IPAddress: "192.168.1.1",
		GeoInfo: &coretypes.GeoInfo{
			Continent:      "TestContinent",
			ContinentCode:  "TC",
			Country:        "TestCountry",
			CountryISO:     "CO ",
			City:           "TestCity",
			Latitude:       52.52,
			Longitude:      13.405,
			ASNumber:       11111,
			ASOrganization: "TestAS",
		},
	}

	err := repo.CreateIPAddress(context.Background(), ip)
	require.NoError(t, err)

	// Verify the IP address was inserted correctly
	dbIPAddress, err := models.IPAddresses(models.IPAddressWhere.IPAddress.EQ("192.168.1.1")).One(context.Background(), repo.db)
	require.NoError(t, err)
	assert.Equal(t, ip.IPAddress, dbIPAddress.IPAddress)
	assert.Equal(t, ip.Latitude, dbIPAddress.Latitude.Float64)
	assert.Equal(t, ip.Longitude, dbIPAddress.Longitude.Float64)

	// Verify the continent was inserted correctly
	dbContinent, err := models.Continents(models.ContinentWhere.ContinentName.EQ("TestContinent")).One(context.Background(), repo.db)
	require.NoError(t, err)
	assert.Equal(t, ip.Continent, dbContinent.ContinentName)
	assert.Equal(t, ip.ContinentCode, dbContinent.Code)

	// Verify the country was inserted correctly
	dbCountry, err := models.Countries(models.CountryWhere.CountryName.EQ("TestCountry")).One(context.Background(), repo.db)
	require.NoError(t, err)
	assert.Equal(t, ip.Country, dbCountry.CountryName)
	assert.Equal(t, ip.CountryISO, dbCountry.IsoCode)

	// Verify the city was inserted correctly
	dbCity, err := models.Cities(models.CityWhere.CityName.EQ("TestCity")).One(context.Background(), repo.db)
	require.NoError(t, err)
	assert.Equal(t, ip.City, dbCity.CityName)

	// Verify the autonomous system was inserted correctly
	dbAS, err := models.AutonomousSystems(models.AutonomousSystemWhere.AsName.EQ("TestAS")).One(context.Background(), repo.db)
	require.NoError(t, err)
	assert.Equal(t, ip.ASOrganization, dbAS.AsName)
	assert.Equal(t, int(ip.ASNumber), dbAS.AsNumber) //nolint:gosec // ignore integer overflow
}
