-- Create schema if it doesn't exist
CREATE SCHEMA IF NOT EXISTS tibber;

-- Create tibber_tokens table
CREATE TABLE IF NOT EXISTS tibber.tibber_tokens (
    id SERIAL PRIMARY KEY,
    token TEXT NOT NULL UNIQUE,
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create homes table
CREATE TABLE IF NOT EXISTS tibber.homes (
    id TEXT PRIMARY KEY,
    type TEXT,
    size INTEGER,
    app_nickname TEXT,
    app_avatar TEXT,
    main_fuse_size INTEGER,
    number_of_residents INTEGER,
    time_zone TEXT,
    address_1 TEXT,
    address_2 TEXT,
    address_3 TEXT,
    postal_code TEXT,
    city TEXT,
    country TEXT,
    latitude TEXT,
    longitude TEXT,
    consumption_ean TEXT,
    grid_company TEXT,
    grid_area_code TEXT,
    price_area_code TEXT,
    production_ean TEXT,
    energy_tax_type TEXT,
    vat_type TEXT,
    estimated_annual_consumption DOUBLE PRECISION,
    real_time_consumption_enabled BOOLEAN,
    owner_id INTEGER,
    token_id INTEGER REFERENCES tibber.tibber_tokens(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create prices table
CREATE TABLE IF NOT EXISTS tibber.prices (
    id SERIAL PRIMARY KEY,
    home_id TEXT REFERENCES tibber.homes(id),
    price_date DATE NOT NULL,
    hour_of_day INTEGER NOT NULL,
    total DECIMAL(10,4) NOT NULL,
    energy DECIMAL(10,4) NOT NULL,
    tax DECIMAL(10,4) NOT NULL,
    currency VARCHAR(10) NOT NULL,
    level VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(home_id, price_date, hour_of_day)
);

-- Create consumption table
CREATE TABLE IF NOT EXISTS tibber.consumption (
    id SERIAL PRIMARY KEY,
    home_id TEXT REFERENCES tibber.homes(id),
    from_time TIMESTAMP WITH TIME ZONE,
    to_time TIMESTAMP WITH TIME ZONE,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    consumption DECIMAL(10,4) NOT NULL,
    cost DECIMAL(10,4),
    currency VARCHAR(10),
    unit VARCHAR(10) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(home_id, from_time)
);

-- Create production table
CREATE TABLE IF NOT EXISTS tibber.production (
    id SERIAL PRIMARY KEY,
    home_id TEXT REFERENCES tibber.homes(id),
    from_time TIMESTAMP WITH TIME ZONE,
    to_time TIMESTAMP WITH TIME ZONE,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    production DECIMAL(10,4) NOT NULL,
    profit DECIMAL(10,4),
    currency VARCHAR(10),
    unit VARCHAR(10),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(home_id, from_time)
);

-- Create real_time_measurements table
CREATE TABLE IF NOT EXISTS tibber.real_time_measurements (
    id SERIAL PRIMARY KEY,
    home_id TEXT REFERENCES tibber.homes(id),
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    power DECIMAL(10,4) NOT NULL,
    min_power DECIMAL(10,4),
    average_power DECIMAL(10,4),
    max_power DECIMAL(10,4),
    power_production DECIMAL(10,4),
    max_power_production DECIMAL(10,4),
    accumulated_consumption DECIMAL(10,4),
    accumulated_production DECIMAL(10,4),
    last_meter_consumption DECIMAL(10,4),
    last_meter_production DECIMAL(10,4),
    current_l1 DECIMAL(10,4),
    current_l2 DECIMAL(10,4),
    current_l3 DECIMAL(10,4),
    voltage_phase1 DECIMAL(10,4),
    voltage_phase2 DECIMAL(10,4),
    voltage_phase3 DECIMAL(10,4),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(home_id, timestamp)
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_prices_home_id ON tibber.prices(home_id);
CREATE INDEX IF NOT EXISTS idx_prices_date ON tibber.prices(price_date);
CREATE INDEX IF NOT EXISTS idx_consumption_home_id ON tibber.consumption(home_id);
CREATE INDEX IF NOT EXISTS idx_consumption_from_time ON tibber.consumption(from_time);
CREATE INDEX IF NOT EXISTS idx_production_home_id ON tibber.production(home_id);
CREATE INDEX IF NOT EXISTS idx_production_from_time ON tibber.production(from_time);
CREATE INDEX IF NOT EXISTS idx_real_time_measurements_home_id ON tibber.real_time_measurements(home_id);
CREATE INDEX IF NOT EXISTS idx_real_time_measurements_timestamp ON tibber.real_time_measurements(timestamp); 