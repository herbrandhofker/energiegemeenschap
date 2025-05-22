-- Set timezone to Europe/Amsterdam
SET timezone = 'Europe/Amsterdam';

-- Create owners table
CREATE TABLE IF NOT EXISTS tibber.owners (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create homes table
CREATE TABLE IF NOT EXISTS tibber.homes (
    id VARCHAR(255) PRIMARY KEY,
    owner_id INTEGER REFERENCES tibber.owners(id),
    name VARCHAR(255) NOT NULL,
    address VARCHAR(255) NOT NULL,
    has_production_capability BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create prices table
CREATE TABLE IF NOT EXISTS tibber.prices (
    id SERIAL PRIMARY KEY,
    home_id VARCHAR(255) REFERENCES tibber.homes(id),
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    total_price DECIMAL(10,4) NOT NULL,
    energy_price DECIMAL(10,4) NOT NULL,
    tax_price DECIMAL(10,4) NOT NULL,
    currency VARCHAR(10) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(home_id, timestamp)
);

-- Create consumption table
CREATE TABLE IF NOT EXISTS tibber.consumption (
    id SERIAL PRIMARY KEY,
    home_id VARCHAR(255) REFERENCES tibber.homes(id),
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    consumption DECIMAL(10,4) NOT NULL,
    unit VARCHAR(10) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(home_id, timestamp)
);

-- Create production table
CREATE TABLE IF NOT EXISTS tibber.production (
    id SERIAL PRIMARY KEY,
    home_id VARCHAR(255) REFERENCES tibber.homes(id),
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    production DECIMAL(10,4) NOT NULL,
    unit VARCHAR(10) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(home_id, timestamp)
);

-- Create real_time_measurements table
CREATE TABLE IF NOT EXISTS tibber.real_time_measurements (
    id SERIAL PRIMARY KEY,
    home_id VARCHAR(255) REFERENCES tibber.homes(id),
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    power DECIMAL(10,4) NOT NULL,
    power_production DECIMAL(10,4) NOT NULL,
    accumulated_consumption DECIMAL(10,4) NOT NULL,
    accumulated_production DECIMAL(10,4) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(home_id, timestamp)
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_prices_home_id ON tibber.prices(home_id);
CREATE INDEX IF NOT EXISTS idx_prices_timestamp ON tibber.prices(timestamp);
CREATE INDEX IF NOT EXISTS idx_consumption_home_id ON tibber.consumption(home_id);
CREATE INDEX IF NOT EXISTS idx_consumption_timestamp ON tibber.consumption(timestamp);
CREATE INDEX IF NOT EXISTS idx_production_home_id ON tibber.production(home_id);
CREATE INDEX IF NOT EXISTS idx_production_timestamp ON tibber.production(timestamp);
CREATE INDEX IF NOT EXISTS idx_real_time_measurements_home_id ON tibber.real_time_measurements(home_id);
CREATE INDEX IF NOT EXISTS idx_real_time_measurements_timestamp ON tibber.real_time_measurements(timestamp); 