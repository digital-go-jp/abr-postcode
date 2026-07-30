package model

// Address is the town and city information both address endpoints return.
// It is embedded rather than repeated so the JSON field order stays identical
// across the two responses.
type Address struct {
	LgCode       string `json:"lg_code"`
	MachiazaID   string `json:"machiaza_id"`
	Pref         string `json:"pref"`
	County       string `json:"county"`
	City         string `json:"city"`
	Ward         string `json:"ward"`
	KyotoSt      string `json:"kyoto_st"`
	OazaCho      string `json:"oaza_cho"`
	Chome        string `json:"chome"`
	Koaza        string `json:"koaza"`
	MachiazaDist string `json:"machiaza_dist"`
}

// NewAddress combines a town with the city that owns it.
func NewAddress(town Town, city City) Address {
	return Address{
		LgCode:       town.LgCode,
		MachiazaID:   town.MachiazaID,
		Pref:         city.Pref,
		County:       city.County,
		City:         city.City,
		Ward:         city.Ward,
		KyotoSt:      town.KyotoSt,
		OazaCho:      town.OazaCho,
		Chome:        town.Chome,
		Koaza:        town.Koaza,
		MachiazaDist: town.MachiazaDist,
	}
}

// AddressResponse represents address information returned by /post_code endpoint.
type AddressResponse struct {
	Address
	PostCode string `json:"post_code"`
}

// MachiazaResponse represents address information returned by /machiaza endpoint.
type MachiazaResponse struct {
	Address
	PostCodes []string `json:"post_codes"`
}
