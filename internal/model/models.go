package model

import (
	"time"
)

// User represents a Tibber user with associated data
type User struct {
	ID          string    `json:"userId"`
	Name        string    `json:"name"`
	Email       string    `json:"login"`
	AccountType string    `json:"accountType"`
	LastLogin   time.Time `json:"lastLogin"`
	Homes       []Home    `json:"homes"`
	APIToken    string    `json:"-"` // Don't include this in JSON responses
}

// LegalEntity
type Owner struct {
	Name        string `json:"name"`
	FirstName   string `json:"firstName"`
	MiddleName  string `json:"middleName"`
	LastName    string `json:"lastName"`
	AccountType string `json:"accountType"`
	IsCompany   bool   `json:"isCompany"`
}

// MeteringPointData represents metering point data for a home
type MeteringPointData struct {
	ConsumptionEan             string  `json:"consumptionEan"`
	GridCompany                string  `json:"gridCompany"`
	GridAreaCode               string  `json:"gridAreaCode"`
	PriceAreaCode              string  `json:"priceAreaCode"`
	ProductionEan              string  `json:"productionEan,omitempty"`
	EnergyTaxType              string  `json:"energyTaxType"`
	VatType                    string  `json:"vatType"`
	EstimatedAnnualConsumption float64 `json:"estimatedAnnualConsumption"`
}

// Home represents a home associated with a Tibber user
type Home struct {
	Id                  string            `json:"id"`
	Type                string            `json:"type"` // HOUSE, APARTMENT, etc.
	Size                int               `json:"size"`
	Address             Address           `json:"address"`
	NumberOfResidents   int               `json:"numberOfResidents"`
	TimeZone            string            `json:"timeZone"`
	AppNickname         string            `json:"appNickname,omitempty"`
	AppAvatar           string            `json:"appAvatar,omitempty"`
	MainFuseSize        int               `json:"mainFuseSize,omitempty"`
	MeteringPointData   MeteringPointData `json:"meteringPointData,omitempty"`
	Features            HomeFeatures      `json:"features"`
	Consumption         []Consumption     `json:"consumption,omitempty"`
	Production          []Production      `json:"production,omitempty"`
	CurrentSubscription *Subscription     `json:"currentSubscription,omitempty"`
	Owner               *User             `json:"owner,omitempty"`
}

// Address represents the address of a home
type Address struct {
	Address1   string `json:"address1"`
	Address2   string `json:"address2"`
	Address3   string `json:"address3"`
	PostalCode string `json:"postalCode"`
	City       string `json:"city"`
	Country    string `json:"country"`
	Latitude   string `json:"latitude"`
	Longitude  string `json:"longitude"`
}

// HomeFeatures represents the features available for a home
type HomeFeatures struct {
	RealTimeConsumptionEnabled bool `json:"realTimeConsumptionEnabled"`
	// Remove the fields that don't exist in the API
	// We'll determine production capability based on whether productionEan exists
}

// HomeConsumptionEdge Consumption represents consumption data for a specific time period
type Consumption struct {
	From            string  `json:"from"`
	To              string  `json:"to"`
	Currency        string  `json:"currency"`
	Cost            float64 `json:"cost"`
	UnitPrice       float64 `json:"unitPrice"`
	UnitPriceVat    float64 `json:"unitPriceVat"`
	Consumption     float64 `json:"consumption"`
	ConsumptionUnit string  `json:"consumptionUnit"`
}

// Price represents price information
type Price struct {
	Total     float64 `json:"total"`
	Energy    float64 `json:"energy"`
	Tax       float64 `json:"tax"`
	StartTime string  `json:"startsAt"`
	EndTime   string  `json:"endsAt,omitempty"`
	Currency  string  `json:"currency"`
	Level     string  `json:"level,omitempty"` // VERY_CHEAP, CHEAP, NORMAL, EXPENSIVE, VERY_EXPENSIVE
}

// PriceInfo represents price information for different time periods
type PriceInfo struct {
	Current  Price   `json:"current"`
	Today    []Price `json:"today"`    // This must be a slice (array)
	Tomorrow []Price `json:"tomorrow"` // This must be a slice (array)
}

// Subscription represents a user's subscription details
type Subscription struct {
	ID          string    `json:"id"`
	Subscriber  Owner     `json:"subscriber"`
	Status      string    `json:"status"`
	PriceRating string    `json:"priceRating,omitempty"`
	PriceInfo   PriceInfo `json:"priceInfo,omitempty"`
}

// ConsumptionSummary summarizes consumption data
type ConsumptionSummary struct {
	From        string
	To          string
	Consumption float64
	Cost        float64
	Currency    string
}

// ToSummary converts Consumption to a simplified ConsumptionSummary
func (c *Consumption) ToSummary() ConsumptionSummary {
	return ConsumptionSummary{
		From:        c.From,
		To:          c.To,
		Consumption: c.Consumption,
		Cost:        c.Cost,
		Currency:    c.Currency,
	}
}

// Production represents energy production data for a specific time period
type Production struct {
	From           string  `json:"from"`
	To             string  `json:"to"`
	Profit         float64 `json:"profit"`
	UnitPrice      float64 `json:"unitPrice"`
	UnitPriceVAT   float64 `json:"unitPriceVAT"`
	Production     float64 `json:"production"`
	ProductionUnit string  `json:"productionUnit"`
	Currency       string  `json:"currency,omitempty"`
}

// ProductionSummary summarizes production data
type ProductionSummary struct {
	From       string
	To         string
	Production float64
	Profit     float64
	Currency   string
}

// ToSummary converts Production to a simplified ProductionSummary
func (p *Production) ToSummary() ProductionSummary {
	return ProductionSummary{
		From:       p.From,
		To:         p.To,
		Production: p.Production,
		Profit:     p.Profit,
		Currency:   p.Currency,
	}
}

// Measurement represents live power measurement data from Tibber
type Measurement struct {
	Timestamp              time.Time `json:"timestamp"`
	Power                  float64   `json:"power"`
	PowerProduction        float64   `json:"powerProduction"`
	MinPower               float64   `json:"minPower"`
	AveragePower           float64   `json:"averagePower"`
	MaxPower               float64   `json:"maxPower"`
	MaxPowerProduction     float64   `json:"maxPowerProduction"`
	CurrentL1              *float64  `json:"currentL1,omitempty"`
	CurrentL2              *float64  `json:"currentL2,omitempty"`
	CurrentL3              *float64  `json:"currentL3,omitempty"`
	VoltagePhase1          *float64  `json:"voltagePhase1,omitempty"`
	VoltagePhase2          *float64  `json:"voltagePhase2,omitempty"`
	VoltagePhase3          *float64  `json:"voltagePhase3,omitempty"`
	AccumulatedConsumption float64   `json:"accumulatedConsumption"`
	AccumulatedProduction  float64   `json:"accumulatedProduction"`
	LastMeterConsumption   float64   `json:"lastMeterConsumption"`
	LastMeterProduction    float64   `json:"lastMeterProduction"`
}

// MeasurementSummary provides a simpler view of measurement data
type MeasurementSummary struct {
	Timestamp   time.Time
	Power       float64
	IsProducing bool
	Production  float64
}

// ToSummary converts a full Measurement to a simplified MeasurementSummary
func (m *Measurement) ToSummary() MeasurementSummary {
	return MeasurementSummary{
		Timestamp:   m.Timestamp,
		Power:       m.Power,
		IsProducing: m.PowerProduction > 0,
		Production:  m.PowerProduction,
	}
}
