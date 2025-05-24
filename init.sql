-- This file runs when PostgreSQL container starts
-- Create any additional configurations if needed

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Set timezone
SET timezone = 'UTC+7';
