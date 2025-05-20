package service_db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"ws/internal/tibber"
)

// RealTimeService handles real-time measurement operations
type RealTimeService struct {
	DB *sql.DB
}

// StoreMeasurement stores a real-time measurement in the database
func (s *RealTimeService) StoreMeasurement(ctx context.Context, homeID string, measurement tibber.Measurement) error {
	query := `
		INSERT INTO real_time_measurements (
			home_id, timestamp, power, power_production,
			min_power, average_power, max_power, max_power_production,
			accumulated_consumption, accumulated_production,
			last_meter_consumption, last_meter_production,
			current_l1, current_l2, current_l3,
			voltage_phase1, voltage_phase2, voltage_phase3
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18
		)
		ON CONFLICT (home_id, timestamp) DO UPDATE SET
			power = EXCLUDED.power,
			power_production = EXCLUDED.power_production,
			min_power = EXCLUDED.min_power,
			average_power = EXCLUDED.average_power,
			max_power = EXCLUDED.max_power,
			max_power_production = EXCLUDED.max_power_production,
			accumulated_consumption = EXCLUDED.accumulated_consumption,
			accumulated_production = EXCLUDED.accumulated_production,
			last_meter_consumption = EXCLUDED.last_meter_consumption,
			last_meter_production = EXCLUDED.last_meter_production,
			current_l1 = EXCLUDED.current_l1,
			current_l2 = EXCLUDED.current_l2,
			current_l3 = EXCLUDED.current_l3,
			voltage_phase1 = EXCLUDED.voltage_phase1,
			voltage_phase2 = EXCLUDED.voltage_phase2,
			voltage_phase3 = EXCLUDED.voltage_phase3
	`

	_, err := s.DB.ExecContext(ctx, query,
		homeID,
		measurement.Timestamp,
		measurement.Power,
		measurement.PowerProduction,
		measurement.MinPower,
		measurement.AveragePower,
		measurement.MaxPower,
		measurement.MaxPowerProduction,
		measurement.AccumulatedConsumption,
		measurement.AccumulatedProduction,
		measurement.LastMeterConsumption,
		measurement.LastMeterProduction,
		measurement.CurrentL1,
		measurement.CurrentL2,
		measurement.CurrentL3,
		measurement.VoltagePhase1,
		measurement.VoltagePhase2,
		measurement.VoltagePhase3,
	)

	if err != nil {
		return fmt.Errorf("error storing real-time measurement: %w", err)
	}

	return nil
}

// GetLatestMeasurements returns the latest measurements for a specific home
func (s *RealTimeService) GetLatestMeasurements(ctx context.Context, homeID string, limit int) ([]tibber.Measurement, error) {
	query := `
		SELECT 
			timestamp, power, power_production,
			min_power, average_power, max_power, max_power_production,
			accumulated_consumption, accumulated_production,
			last_meter_consumption, last_meter_production,
			current_l1, current_l2, current_l3,
			voltage_phase1, voltage_phase2, voltage_phase3
		FROM real_time_measurements
		WHERE home_id = $1
		ORDER BY timestamp DESC
		LIMIT $2
	`

	rows, err := s.DB.QueryContext(ctx, query, homeID, limit)
	if err != nil {
		return nil, fmt.Errorf("error querying latest measurements: %w", err)
	}
	defer rows.Close()

	var measurements []tibber.Measurement
	for rows.Next() {
		var m tibber.Measurement
		var timestamp time.Time
		err := rows.Scan(
			&timestamp,
			&m.Power,
			&m.PowerProduction,
			&m.MinPower,
			&m.AveragePower,
			&m.MaxPower,
			&m.MaxPowerProduction,
			&m.AccumulatedConsumption,
			&m.AccumulatedProduction,
			&m.LastMeterConsumption,
			&m.LastMeterProduction,
			&m.CurrentL1,
			&m.CurrentL2,
			&m.CurrentL3,
			&m.VoltagePhase1,
			&m.VoltagePhase2,
			&m.VoltagePhase3,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning measurement: %w", err)
		}
		m.Timestamp = timestamp
		measurements = append(measurements, m)
	}

	return measurements, nil
}

// CleanupOldMeasurements verwijdert metingen ouder dan de opgegeven duur
func (s *RealTimeService) CleanupOldMeasurements(ctx context.Context, olderThan time.Duration) error {
	query := `
		DELETE FROM real_time_measurements 
		WHERE timestamp < NOW() - INTERVAL '1 day'
	`
	_, err := s.DB.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("error cleaning up old measurements: %w", err)
	}
	return nil
}
