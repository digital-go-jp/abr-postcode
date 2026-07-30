package model

type City struct {
	LgCode string `json:"lg_code"`
	Pref   string `json:"pref"`
	County string `json:"county"`
	City   string `json:"city"`
	Ward   string `json:"ward"`
}
