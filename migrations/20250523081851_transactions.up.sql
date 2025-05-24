CREATE TABLE transactions (
  id UUID PRIMARY KEY,
  wallet_id UUID REFERENCES wallets(id),
  transaction_type_id INT REFERENCES transaction_types(id),
  amount MONEY NOT NULL,
  balance_before MONEY NOT NULL,
  balance_after MONEY NOT NULL,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);