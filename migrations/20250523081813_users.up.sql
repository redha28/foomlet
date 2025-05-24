-- USERS
CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
  firstname VARCHAR(255),
  lastname VARCHAR(255),
  pin VARCHAR(255),
  phone VARCHAR(20) UNIQUE,
  address VARCHAR,
  created_at TIMESTAMP DEFAULT NOW() NULL,
  updated_at TIMESTAMP DEFAULT NOW() NULL
);