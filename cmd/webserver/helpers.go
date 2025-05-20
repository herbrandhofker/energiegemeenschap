package main

import (
	"fmt"
	"time"
	"ws/internal/model"
)

// sortConsumptionByDate sorteert consumptiegegevens op datum (nieuwste eerst)
func sortConsumptionByDate(consumption []model.Consumption) []model.Consumption {
	// Maak een kopie om te sorteren
	sorted := make([]model.Consumption, len(consumption))
	copy(sorted, consumption)

	// Sorteer op datum (nieuwste eerst) - direct string vergelijking
	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			// Vergelijk direct de strings - ISO 8601 formaat is lexicografisch sorteerbaar
			if sorted[j].From < sorted[j+1].From {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	return sorted
}

// sortConsumptionByDateOldestFirst sorteert consumptiegegevens op datum (oudste eerst)
func sortConsumptionByDateOldestFirst(consumption []model.Consumption) []model.Consumption {
	// Maak een kopie om te sorteren
	sorted := make([]model.Consumption, len(consumption))
	copy(sorted, consumption)

	// Sorteer op datum (oudste eerst) - direct string vergelijking
	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			// Vergelijk direct de strings - ISO 8601 formaat is lexicografisch sorteerbaar
			if sorted[j].From > sorted[j+1].From {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	return sorted
}

// sortProductionByDate sorteert productiegegevens op datum (nieuwste eerst)
func sortProductionByDate(production []model.Production) []model.Production {
	// Maak een kopie om te sorteren
	sorted := make([]model.Production, len(production))
	copy(sorted, production)

	// Sorteer op datum (nieuwste eerst) - direct string vergelijking
	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			// Vergelijk direct de strings - ISO 8601 formaat is lexicografisch sorteerbaar
			if sorted[j].From < sorted[j+1].From {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	return sorted
}

// sortProductionByDateOldestFirst sorteert productiegegevens op datum (oudste eerst)
func sortProductionByDateOldestFirst(production []model.Production) []model.Production {
	// Maak een kopie om te sorteren
	sorted := make([]model.Production, len(production))
	copy(sorted, production)

	// Sorteer op datum (oudste eerst) - direct string vergelijking
	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			// Vergelijk direct de strings - ISO 8601 formaat is lexicografisch sorteerbaar
			if sorted[j].From > sorted[j+1].From {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	return sorted
}

// findHomeByID zoekt naar een home met de gegeven ID en geeft een fout terug als niet gevonden
func (wd *WebDashboard) findHomeByID(homeID string) (*model.Home, error) {
	if homeID == "" {
		return nil, fmt.Errorf("geen home ID opgegeven")
	}

	for i, home := range wd.Homes {
		if home.Id == homeID {
			return &wd.Homes[i], nil // Return pointer naar het element in de array
		}
	}

	return nil, fmt.Errorf("home met ID '%s' niet gevonden", homeID)
}

// calculateEndTime berekent de eindtijd (1 uur na starttijd) voor prijsgeldigheid
func calculateEndTime(startTimeStr string) (string, error) {
	layout := "2006-01-02T15:04:05.000-07:00"

	// Datumstring omzetten naar time.Time
	parsedTime, err := time.Parse(layout, startTimeStr)
	if err != nil {
		return "", fmt.Errorf("kan tijd niet parsen: %w", err)
	}

	// Bereken eindtijd (1 uur later) en formatteer als RFC3339
	return parsedTime.Add(time.Hour).Format(time.RFC3339), nil
}
