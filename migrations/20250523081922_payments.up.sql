CREATE TABLE payments (
  transaction_id UUID PRIMARY KEY REFERENCES transactions(id),
  user_id UUID REFERENCES users(id),
  amount INT,
  remarks VARCHAR,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);
