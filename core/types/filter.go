package types

import (
	"reflect"
)

// PeerFilter represents the filter options for peers
type PeerFilter struct {
	PeerID            string   `query:"id"`
	Continents        []string `query:"continents"`
	ContinentsCodes   []string `query:"continents_codes"`
	Countries         []string `query:"countries"`
	CountriesISOCodes []string `query:"countries_codes"`
	Cities            []string `query:"cities"`
	ASOrganizations   []string `query:"as_names"`
	ASNumbers         []string `query:"as_numbers"`
	ClientNames       []string `query:"clients"`
	SyncStatus        *bool    `query:"synced"`
	SpecVersion       string   `query:"-"` // Not bound from query params, set programmatically
}

func CompareFilters(f1, f2 *PeerFilter) bool {
	return reflect.DeepEqual(f1, f2)
}
