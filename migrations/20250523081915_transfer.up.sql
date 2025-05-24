CREATE TABLE transfer (
  transaction_id UUID PRIMARY KEY REFERENCES transactions(id),
  target_user UUID REFERENCES users(id),
  sender_user UUID REFERENCES users(id),
  remarks VARCHAR,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);