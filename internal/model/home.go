package model

// HomeWithPrices combines Home and PriceInfo
type HomeWithPrices struct {
	Home   *Home
	Prices PriceInfo
}

// HomeWithConsumption combines Home and Consumption data
type HomeWithConsumption struct {
	Home        *Home
	Consumption []Consumption
}

// HomeWithProduction combines Home and Production data
type HomeWithProduction struct {
	Home       *Home
	Production []Production
}
