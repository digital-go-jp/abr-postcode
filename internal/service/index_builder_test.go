package service

import (
	"abr-postcode/internal/model"
	"reflect"
	"testing"
)

func TestBuildIndexes(t *testing.T) {
	cities := []model.City{
		{LgCode: "011011", Pref: "北海道", City: "札幌市", Ward: "中央区"},
		{LgCode: "131016", Pref: "東京都", City: "千代田区"},
	}

	towns := []model.Town{
		{LgCode: "011011", MachiazaID: "0001001", OazaCho: "旭ケ丘", Chome: "1丁目"},
		{LgCode: "131016", MachiazaID: "0006000", OazaCho: "千代田"},
	}

	postCodeMappings := []model.PostCodeMapping{
		{PostCode: "0640941", LgCode: "011011", MachiazaID: "0001001"},
		{PostCode: "1000001", LgCode: "131016", MachiazaID: "0006000"},
	}

	data := BuildIndexes(cities, towns, postCodeMappings)

	t.Run("cities indexed correctly", func(t *testing.T) {
		if len(data.Cities) != 2 {
			t.Errorf("expected 2 cities, got %d", len(data.Cities))
		}
		if city, ok := data.Cities["011011"]; !ok {
			t.Error("city 011011 not found")
		} else if city.Pref != "北海道" {
			t.Errorf("expected pref 北海道, got %s", city.Pref)
		}
	})

	t.Run("towns indexed correctly", func(t *testing.T) {
		if len(data.Towns) != 2 {
			t.Errorf("expected 2 towns, got %d", len(data.Towns))
		}
		key := "0110110001001"
		if town, ok := data.Towns[key]; !ok {
			t.Errorf("town %s not found", key)
		} else if town.OazaCho != "旭ケ丘" {
			t.Errorf("expected oaza_cho 旭ケ丘, got %s", town.OazaCho)
		}
	})

	t.Run("post code mappings indexed correctly", func(t *testing.T) {
		if len(data.PostCodeMappings) != 2 {
			t.Errorf("expected 2 post codes, got %d", len(data.PostCodeMappings))
		}
		if mappings, ok := data.PostCodeMappings["0640941"]; !ok {
			t.Error("post code 0640941 not found")
		} else if len(mappings) != 1 {
			t.Errorf("expected 1 mapping, got %d", len(mappings))
		}
	})

	t.Run("reverse index built correctly", func(t *testing.T) {
		key := "0110110001001"
		if postCodes, ok := data.TownToPostCodes[key]; !ok {
			t.Errorf("reverse index for %s not found", key)
		} else if len(postCodes) != 1 || postCodes[0] != "0640941" {
			t.Errorf("expected [0640941], got %v", postCodes)
		}
	})
}

func TestBuildIndexes_Empty(t *testing.T) {
	data := BuildIndexes(nil, nil, nil)

	if len(data.Cities) != 0 {
		t.Errorf("expected 0 cities, got %d", len(data.Cities))
	}
	if len(data.Towns) != 0 {
		t.Errorf("expected 0 towns, got %d", len(data.Towns))
	}
	if len(data.PostCodeMappings) != 0 {
		t.Errorf("expected 0 post codes, got %d", len(data.PostCodeMappings))
	}
}

func TestBuildIndexes_MultiplePostCodesPerTown(t *testing.T) {
	cities := []model.City{
		{LgCode: "342131", Pref: "広島県", City: "廿日市市"},
	}

	towns := []model.Town{
		{LgCode: "342131", MachiazaID: "0069000", OazaCho: "宮島町"},
	}

	// Same town has multiple post codes, listed here out of order
	postCodeMappings := []model.PostCodeMapping{
		{PostCode: "7390503", LgCode: "342131", MachiazaID: "0069000"},
		{PostCode: "7390501", LgCode: "342131", MachiazaID: "0069000"},
		{PostCode: "7390502", LgCode: "342131", MachiazaID: "0069000"},
	}

	data := BuildIndexes(cities, towns, postCodeMappings)

	key := "3421310069000"
	postCodes := data.TownToPostCodes[key]

	expected := []string{"7390501", "7390502", "7390503"}
	if !reflect.DeepEqual(postCodes, expected) {
		t.Errorf("expected sorted post codes %v, got %v", expected, postCodes)
	}
}

func TestBuildIndexes_SamePostCodeMultipleTowns(t *testing.T) {
	cities := []model.City{
		{LgCode: "131016", Pref: "東京都", City: "千代田区"},
	}

	towns := []model.Town{
		{LgCode: "131016", MachiazaID: "0001001", OazaCho: "内幸町", Chome: "一丁目"},
		{LgCode: "131016", MachiazaID: "0001002", OazaCho: "内幸町", Chome: "二丁目"},
	}

	// Same post code maps to multiple towns
	postCodeMappings := []model.PostCodeMapping{
		{PostCode: "1000011", LgCode: "131016", MachiazaID: "0001001"},
		{PostCode: "1000011", LgCode: "131016", MachiazaID: "0001002"},
	}

	data := BuildIndexes(cities, towns, postCodeMappings)

	mappings := data.PostCodeMappings["1000011"]
	if len(mappings) != 2 {
		t.Errorf("expected 2 mappings for post code, got %d", len(mappings))
	}
}
