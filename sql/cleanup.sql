-- Drop all tables in the tibber schema
DROP TABLE IF EXISTS tibber.real_time_measurements CASCADE;
DROP TABLE IF EXISTS tibber.production CASCADE;
DROP TABLE IF EXISTS tibber.consumption CASCADE;
DROP TABLE IF EXISTS tibber.prices CASCADE;
DROP TABLE IF EXISTS tibber.homes CASCADE;
DROP TABLE IF EXISTS tibber.owners CASCADE;
DROP TABLE IF EXISTS tibber.tibber_tokens CASCADE;

-- Drop the schema
DROP SCHEMA IF EXISTS tibber CASCADE; 