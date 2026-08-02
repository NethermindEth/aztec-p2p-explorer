package repo

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"

	coretypes "github.com/NethermindEth/aztec-p2p-explorer/core/types"
	"github.com/NethermindEth/aztec-p2p-explorer/database/models"
)

func (r *PeerRepository) CreateIPAddress(ctx context.Context, ipInfo *coretypes.IPInfo) error {
	return r.withTx(ctx, func(tx *sql.Tx) error {
		continentID, err := r.insertContinent(ctx, tx, ipInfo.ContinentCode, ipInfo.Continent)
		if err != nil {
			return err
		}

		countryID, err := r.upsertCountry(ctx, tx, ipInfo.Country, ipInfo.CountryISO, null.NewInt(continentID, continentID != 0))
		if err != nil {
			return err
		}

		cityID, err := r.upsertCity(ctx, tx, ipInfo.City, null.NewInt(countryID, countryID != 0), null.NewInt(continentID, continentID != 0))
		if err != nil {
			return err
		}

		asID, err := r.insertASO(ctx, tx, ipInfo.ASOrganization, int(ipInfo.ASNumber)) //nolint:gosec // ignore integer overflow
		if err != nil {
			return err
		}

		_, err = r.upsertIPAddress(ctx, tx, ipInfo.IPAddress,
			ipInfo.Port,
			null.NewInt(countryID, countryID != 0),
			null.NewInt(continentID, continentID != 0),
			null.NewInt(cityID, cityID != 0),
			null.NewInt(asID, asID != 0),
			null.NewFloat64(ipInfo.Latitude, ipInfo.Latitude != 0),
			null.NewFloat64(ipInfo.Longitude, ipInfo.Longitude != 0),
		)

		return err
	})
}

// upsertIPAddress upserts an IP address record and returns the IP address object
func (r *PeerRepository) upsertIPAddress(
	ctx context.Context,
	exec boil.ContextExecutor,
	ip string,
	port int,
	countryID null.Int,
	continentID null.Int,
	cityID null.Int,
	asID null.Int,
	latitude null.Float64,
	longitude null.Float64,
) (*models.IPAddress, error) {
	ipAddr := &models.IPAddress{
		IPAddress:   ip,
		Port:        port,
		CountryID:   countryID,
		ContinentID: continentID,
		CityID:      cityID,
		AsID:        asID,
		Latitude:    latitude,
		Longitude:   longitude,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	columns := []string{models.IPAddressColumns.IPAddress}
	for _, field := range []struct {
		valid  bool
		column string
	}{
		{countryID.Valid, models.IPAddressColumns.CountryID},
		{continentID.Valid, models.IPAddressColumns.ContinentID},
		{cityID.Valid, models.IPAddressColumns.CityID},
		{asID.Valid, models.IPAddressColumns.AsID},
		{latitude.Valid, models.IPAddressColumns.Latitude},
		{longitude.Valid, models.IPAddressColumns.Longitude},
	} {
		if field.valid {
			columns = append(columns, field.column)
		}
	}

	if len(columns) > 1 {
		columns = append(columns, models.IPAddressColumns.UpdatedAt)
	}

	err := ipAddr.Upsert(ctx, exec, true,
		[]string{models.IPAddressColumns.IPAddress, models.IPAddressColumns.Port},
		boil.Whitelist(columns...), boil.Infer())
	if err != nil {
		return nil, fmt.Errorf("upsert ip address: %w", err)
	}

	r.logger.Debug("Upserted IP address", "ip", ipAddr)
	return ipAddr, nil
}

// insertContinent inserts a continent record if not exists and returns the continent ID
func (r *PeerRepository) insertContinent(ctx context.Context, exec boil.ContextExecutor, code, name string) (int, error) {
	if code == "" || name == "" {
		return 0, nil
	}

	c := &models.Continent{
		ContinentName: name,
		Code:          code,
	}

	err := c.Upsert(ctx, exec, true, []string{models.ContinentColumns.Code},
		boil.Whitelist(models.ContinentColumns.Code), boil.Infer())
	if err != nil {
		return 0, fmt.Errorf("insert continent: %w", err)
	}

	return c.ID, nil
}

// upsertCountry upserts a country record and returns the country ID
func (r *PeerRepository) upsertCountry(
	ctx context.Context,
	exec boil.ContextExecutor,
	name string,
	isoCode string,
	continentID null.Int,
) (int, error) {
	if name == "" || isoCode == "" {
		return 0, nil
	}

	c := &models.Country{
		CountryName: name,
		IsoCode:     isoCode,
		ContinentID: continentID,
	}

	columns := []string{models.CountryColumns.IsoCode}
	if continentID.Valid {
		columns = append(columns, models.CountryColumns.ContinentID)
	}

	err := c.Upsert(ctx, exec, true, []string{models.CountryColumns.IsoCode},
		boil.Whitelist(columns...), boil.Infer())
	if err != nil {
		return 0, fmt.Errorf("insert country: %w", err)
	}

	return c.ID, nil
}

// upsertCity upserts a city record and returns the city ID
func (r *PeerRepository) upsertCity(
	ctx context.Context,
	exec boil.ContextExecutor,
	name string,
	countryID null.Int,
	continentID null.Int,
) (int, error) {
	if name == "" {
		return 0, nil
	}

	c := &models.City{
		CityName:    name,
		CountryID:   countryID,
		ContinentID: continentID,
	}

	columns := []string{models.CityColumns.CityName}
	if countryID.Valid {
		columns = append(columns, models.CityColumns.CountryID)
	}

	if continentID.Valid {
		columns = append(columns, models.CityColumns.ContinentID)
	}

	err := c.Upsert(ctx, exec, true, []string{models.CityColumns.CityName, models.CityColumns.CountryID, models.CityColumns.ContinentID},
		boil.Whitelist(columns...), boil.Infer())
	if err != nil {
		return 0, fmt.Errorf("insert city: %w", err)
	}

	return c.ID, nil
}

// insertASO inserts an autonomous system organisation record if not exists and returns the ASO ID
func (r *PeerRepository) insertASO(
	ctx context.Context,
	exec boil.ContextExecutor,
	name string,
	num int,
) (int, error) {
	if name == "" || num == 0 {
		return 0, nil
	}

	aso := &models.AutonomousSystem{
		AsName:   name,
		AsNumber: num,
	}

	err := aso.Upsert(ctx, exec, true, []string{models.AutonomousSystemColumns.AsNumber},
		boil.Whitelist(models.AutonomousSystemColumns.AsNumber), boil.Infer())
	if err != nil {
		return 0, fmt.Errorf("insert aso: %w", err)
	}

	return aso.ID, nil
}
