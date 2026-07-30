package service

import (
	"os"
	"path/filepath"
	"slices"

	"abr-postcode/internal/model"
)

type AddressData struct {
	Cities           map[string]model.City
	Towns            map[string]model.Town
	PostCodeMappings map[string][]model.PostCodeMapping
	TownToPostCodes  map[string][]string // lg_code+machiaza_id → sorted post_codes
}

// newAddressData sizes the indexes for townCount towns. Growing a map to
// hundreds of thousands of entries reallocates and rehashes it about twenty
// times over, and the last of those holds both tables at once, so the estimate
// is worth making even when it is rough.
//
// Only Towns and TownToPostCodes get it: there is one town per row of
// town.csv, whereas post codes repeat across towns heavily enough that a
// row-derived size would over-allocate PostCodeMappings several times over.
// Cities is small enough not to matter.
func newAddressData(townCount int) *AddressData {
	return &AddressData{
		Cities:           make(map[string]model.City),
		Towns:            make(map[string]model.Town, townCount),
		PostCodeMappings: make(map[string][]model.PostCodeMapping),
		TownToPostCodes:  make(map[string][]string, townCount),
	}
}

// bytesPerTownRow approximates a row of town.csv, so the number of towns can be
// guessed from the file size before reading it.
const bytesPerTownRow = 38

// estimateRows guesses how many rows path holds. An unreadable file reports 0,
// which only costs the size hint.
func estimateRows(path string, bytesPerRow int) int {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return int(info.Size()) / bytesPerRow
}

// LoadAddressData reads the three normalized CSVs in dataDir straight into the
// lookup indexes.
func LoadAddressData(dataDir string) (*AddressData, error) {
	townPath := filepath.Join(dataDir, "town.csv")
	data := newAddressData(estimateRows(townPath, bytesPerTownRow))

	err := forEachRow(filepath.Join(dataDir, "city.csv"), func(row *csvReader) {
		data.addCity(model.City{
			LgCode: row.value("lg_code"),
			Pref:   row.value("pref"),
			County: row.value("county"),
			City:   row.value("city"),
			Ward:   row.value("ward"),
		})
	})
	if err != nil {
		return nil, err
	}

	err = forEachRow(filepath.Join(dataDir, "town.csv"), func(row *csvReader) {
		data.addTown(model.Town{
			LgCode:       row.value("lg_code"),
			MachiazaID:   row.value("machiaza_id"),
			KyotoSt:      row.value("kyoto_st"),
			OazaCho:      row.value("oaza_cho"),
			Chome:        row.value("chome"),
			Koaza:        row.value("koaza"),
			MachiazaDist: row.value("machiaza_dist"),
		})
	})
	if err != nil {
		return nil, err
	}

	err = forEachRow(filepath.Join(dataDir, "post_code_mapping.csv"), func(row *csvReader) {
		data.addPostCodeMapping(model.PostCodeMapping{
			PostCode:   row.value("post_code"),
			LgCode:     row.value("lg_code"),
			MachiazaID: row.value("machiaza_id"),
		})
	})
	if err != nil {
		return nil, err
	}

	data.sortPostCodes()
	return data, nil
}

// BuildIndexes assembles the lookup indexes from records already in memory.
// Production reads the CSVs through LoadAddressData; this is the seam tests
// use to build a fixture without files.
func BuildIndexes(cities []model.City, towns []model.Town, postCodeMappings []model.PostCodeMapping) *AddressData {
	data := newAddressData(len(towns))

	for _, city := range cities {
		data.addCity(city)
	}
	for _, town := range towns {
		data.addTown(town)
	}
	for _, mapping := range postCodeMappings {
		data.addPostCodeMapping(mapping)
	}

	data.sortPostCodes()
	return data
}

func (d *AddressData) addCity(city model.City) {
	d.Cities[city.LgCode] = city
}

func (d *AddressData) addTown(town model.Town) {
	d.Towns[townKey(town.LgCode, town.MachiazaID)] = town
}

// addPostCodeMapping indexes the mapping by postal code and by town.
func (d *AddressData) addPostCodeMapping(mapping model.PostCodeMapping) {
	d.PostCodeMappings[mapping.PostCode] = append(d.PostCodeMappings[mapping.PostCode], mapping)

	key := townKey(mapping.LgCode, mapping.MachiazaID)
	d.TownToPostCodes[key] = append(d.TownToPostCodes[key], mapping.PostCode)
}

// townKey addresses Towns and TownToPostCodes. A machiaza ID is unique only
// within its municipality, so both codes together identify a town.
func townKey(lgCode, machiazaID string) string {
	return lgCode + machiazaID
}

// Address joins a town with the city that owns it, reporting false when
// either is missing.
func (d *AddressData) Address(lgCode, machiazaID string) (model.Address, bool) {
	town, ok := d.Towns[townKey(lgCode, machiazaID)]
	if !ok {
		return model.Address{}, false
	}
	city, ok := d.Cities[town.LgCode]
	if !ok {
		return model.Address{}, false
	}
	return model.NewAddress(town, city), true
}

// PostCodes returns the town's postal codes, already sorted.
func (d *AddressData) PostCodes(lgCode, machiazaID string) []string {
	return d.TownToPostCodes[townKey(lgCode, machiazaID)]
}

func (d *AddressData) sortPostCodes() {
	for key := range d.TownToPostCodes {
		slices.Sort(d.TownToPostCodes[key])
	}
}
