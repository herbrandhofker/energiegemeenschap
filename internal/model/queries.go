package model

// GraphQL query constants
const (
	// UserQuery retrieves basic user information and homes
	UserQuery = `
        query {
            viewer {
                name
                login
                userId
                accountType
                homes {
                    id
                    timeZone
                    type
                    size
                    numberOfResidents
                    appNickname
                }
            }
        }
    `

	// HomeDetailsQuery retrieves detailed information about homes
	HomeDetailsQuery = `
        query {
            viewer {
                homes {
                    id
                    type
                    size
                    appNickname
                    appAvatar
                    mainFuseSize
                    address {
                        address1
                        address2
                        postalCode
                        city
                        country
                        latitude
                        longitude
                    }
                    meteringPointData {
                        consumptionEan
                        gridCompany
                        gridAreaCode
                        priceAreaCode
                        productionEan
                        energyTaxType
                        vatType
                        estimatedAnnualConsumption
                    }
                    features {
                        realTimeConsumptionEnabled
                    }
                }
            }
        }
    `

	// ConsumptionQuery retrieves consumption data for a home
	ConsumptionQuery = `
        query ($homeId: ID!, $resolution: EnergyResolution!, $last: Int!) {
            viewer {
                home(id: $homeId) {
                    consumption(resolution: $resolution, last: $last) {
                        nodes {
                            from
                            to
                            cost
                            unitPrice
                            unitPriceVAT
                            consumption
                            consumptionUnit
                            currency
                        }
                    }
                }
            }
        }
    `

	// ProductionQuery retrieves production data for a home
	ProductionQuery = `
        query ($homeId: ID!, $resolution: EnergyResolution!, $last: Int!) {
            viewer {
                home(id: $homeId) {
                    production(resolution: $resolution, last: $last) {
                        nodes {
                            from
                            to
                            profit
                            unitPrice
                            unitPriceVAT
                            production
                            productionUnit
                            currency
                        }
                    }
                }
            }
        }
    `

	// PriceQuery retrieves current and future price information
	PriceQuery = `
        query {
            viewer {
                homes {
                    id
                    currentSubscription {
                        priceInfo {
                            current {
                                total
                                energy
                                tax
                                startsAt
                                level
                                currency
                            }
                            today {
                                total
                                energy
                                tax
                                startsAt
                                level
                                currency
                            }
                            tomorrow {
                                total
                                energy
                                tax
                                startsAt
                                level
                                currency
                            }
                        }
                    }
                }
            }
        }
    `

	MeasurementQuery = `
        query {
            viewer {
                home(id: "homeId") {
                subscription {
                    liveMeasurement {
                        timestamp
                        power
                        powerProduction
                        lastMeterConsumption
                        lastMeterProduction
                        accumulatedConsumption
                        accumulatedProduction
                        accumulatedConsumptionLastHour
                        accumulatedProductionLastHour
                        accumulatedCost
                        accumulatedReward
                        currency
                        minPower
                        averagePower
                        maxPower
                        minPowerProduction
                        maxPowerProduction
                        powerFactor
                        voltagePhase1
                        voltagePhase2
                        voltagePhase3
                        currentL1
                        currentL2
                        currentL3
                        signalStrength
                    }
                }
            }
        }
    
`
)

