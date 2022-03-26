package assets

type Asset struct {
	Id int
	Name string
	UnitName string
	Decimals int
}

//TODO: add algod method parameter
//TODO: what is params structure
//TODO: should export fetch?
//TODO: what about __call__, __hash__, __repr__ methods?
func (s *Asset) Fetch(algod interface{}){

	var params map[string]interface{}

	if s.Id > 0 {
		//params = algod.asset_info(self.id)['params']
	} else {

		params = map[string]interface{} {
			"name": "Algo",
			"unit-name": "ALGO",
			"decimals": 6,
		}

	}

	s.Name = params["name"].(string)
	s.UnitName = params["unit-name"].(string)
	s.Decimals = params["decimals"].(int)

}


//TODO: what about __mul__, __add__, __sub__, __gt__, __lt__, __eq__, __repr__ methods?
type AssetAmount struct {
	Asset Asset
	Amount int
}

