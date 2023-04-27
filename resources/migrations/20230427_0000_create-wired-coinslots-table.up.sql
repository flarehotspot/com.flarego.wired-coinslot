CREATE TABLE IF NOT EXISTS wired_coinslots (
    id INT AUTO_INCREMENT PRIMARY KEY,
    alias VARCHAR(255),

    coin_pin INT NOT NULL,
    coin_inhibit_pin INT NOT NULL,
    coin_relay_active BOOLEAN NOT NULL DEFAULT TRUE,
    coin_relay_delay_sec INT DEFAULT 0,
    coin_bouncetime INT DEFAULT 0,

    bill_pin INT,
    bill_inhibit_pin INT NOT NULL DEFAULT 0,
    bill_relay_active BOOLEAN NOT NULL DEFAULT TRUE,
    bill_relay_delay_sec INT NOT NULL DEFAULT 0,
    bill_bouncetime INT NOT NULL DEFAULT 0,

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
