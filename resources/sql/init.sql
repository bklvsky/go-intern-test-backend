CREATE TABLE users
(
    id INT PRIMARY KEY,
    balance DECIMAL (18,2),
    reserve DECIMAL (18,2)
);

INSERT INTO users VALUES (1, 100.0, 0);
INSERT INTO users VALUES (2, 20.5, 20);