BEGIN TRANSACTION;

CREATE TABLE IF NOT EXISTS withdrawals (
   id SERIAL PRIMARY KEY,
   user_id INTEGER NOT NULL,
   order_id INTEGER NOT NULL,
   sum DECIMAL(10, 2) NOT NULL,
   processed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
   FOREIGN KEY (user_id) REFERENCES users(id),
   FOREIGN KEY (order_id) REFERENCES orders(id)
);

COMMIT;