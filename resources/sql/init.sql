CREATE TABLE users
(
    id INT PRIMARY KEY,
    balance DECIMAL (18,2),
    reserve DECIMAL (18,2)
);

CREATE TABLE transactions
(
    id SERIAL PRIMARY KEY,
    order_id INT NOT NULL,
    user_id INT REFERENCES users(id),
    service_id INT,
    cost DECIMAL (18,2),
    time_st TIMESTAMP,
    note VARCHAR(255)
);

INSERT INTO users VALUES (1, 100.0, 0);
INSERT INTO users VALUES (2, 20.5, 20);