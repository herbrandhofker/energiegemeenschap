-- Create owners table
CREATE TABLE IF NOT EXISTS owners (
    id SERIAL PRIMARY KEY,
    name TEXT,
    first_name TEXT,
    middle_name TEXT,
    last_name TEXT,
    address_1 TEXT,
    address_2 TEXT,
    address_3 TEXT,
    city TEXT,
    postal_code TEXT,
    country TEXT,
    latitude TEXT,
    longitude TEXT,
    email TEXT UNIQUE,
    mobile TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create homes table
CREATE TABLE IF NOT EXISTS homes (
    id UUID PRIMARY KEY,
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
    city TEXT,
    postal_code TEXT,
    country TEXT,
    latitude DOUBLE PRECISION,
    longitude DOUBLE PRECISION,
    consumption_ean TEXT,
    grid_company TEXT,
    grid_area_code TEXT,
    price_area_code TEXT,
    production_ean TEXT,
    energy_tax_type TEXT,
    vat_type TEXT,
    estimated_annual_consumption INTEGER,
    real_time_consumption_enabled BOOLEAN,
    owner_id INTEGER REFERENCES owners(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create prices table
CREATE TABLE IF NOT EXISTS prices (
    id SERIAL PRIMARY KEY,
    home_id UUID REFERENCES homes(id),
    price_date DATE,
    hour_of_day INTEGER,
    total DECIMAL(10,4),
    energy DECIMAL(10,4),
    tax DECIMAL(10,4),
    currency TEXT,
    level TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (home_id, price_date, hour_of_day),
    CHECK (hour_of_day >= 0 AND hour_of_day < 24)
);

-- Create consumption table
CREATE TABLE IF NOT EXISTS consumption (
    id SERIAL PRIMARY KEY,
    home_id UUID NOT NULL REFERENCES homes(id),
    from_time TIMESTAMP WITH TIME ZONE NOT NULL,
    to_time TIMESTAMP WITH TIME ZONE NOT NULL,
    consumption DECIMAL(10, 2) NOT NULL,
    cost DECIMAL(10, 2) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(home_id, from_time)
);

-- Create production table
CREATE TABLE IF NOT EXISTS production (
    id SERIAL PRIMARY KEY,
    home_id UUID NOT NULL REFERENCES homes(id),
    from_time TIMESTAMP WITH TIME ZONE NOT NULL,
    to_time TIMESTAMP WITH TIME ZONE NOT NULL,
    production DECIMAL(10, 2) NOT NULL,
    profit DECIMAL(10, 2) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(home_id, from_time)
);

-- Create real_time_measurements table
CREATE TABLE IF NOT EXISTS real_time_measurements (
    id SERIAL PRIMARY KEY,
    home_id UUID REFERENCES homes(id),
    timestamp TIMESTAMP WITH TIME ZONE,
    power DECIMAL(10,4),
    power_production DECIMAL(10,4),
    min_power DECIMAL(10,4),
    average_power DECIMAL(10,4),
    max_power DECIMAL(10,4),
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
    UNIQUE (home_id, timestamp)
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_prices_home_id ON prices(home_id);
CREATE INDEX IF NOT EXISTS idx_prices_price_date ON prices(price_date);
CREATE INDEX IF NOT EXISTS idx_prices_hour_of_day ON prices(hour_of_day);
CREATE INDEX IF NOT EXISTS idx_consumption_home_id ON consumption(home_id);
CREATE INDEX IF NOT EXISTS idx_consumption_from_time ON consumption(from_time);
CREATE INDEX IF NOT EXISTS idx_production_home_id ON production(home_id);
CREATE INDEX IF NOT EXISTS idx_production_from_time ON production(from_time);
CREATE INDEX IF NOT EXISTS idx_real_time_measurements_home_id ON real_time_measurements(home_id);
CREATE INDEX IF NOT EXISTS idx_real_time_measurements_timestamp ON real_time_measurements(timestamp); 