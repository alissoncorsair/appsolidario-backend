-- Create Role Table
CREATE TABLE Role (
  id SERIAL PRIMARY KEY,
  name VARCHAR(50) UNIQUE NOT NULL
);

-- Create User Table
CREATE TABLE "User" (
  id SERIAL PRIMARY KEY,
  name VARCHAR(100) NOT NULL,
  email VARCHAR(100) UNIQUE NOT NULL,
  password_hash CHAR(60) NOT NULL,
  status SMALLINT NOT NULL CHECK (status IN (0, 1)),
  description TEXT,
  postal_code CHAR(8),
  registration_date DATE NOT NULL,
  cpf CHAR(11) UNIQUE NOT NULL,
  role_id INT REFERENCES Role(id),
  points INT DEFAULT 0
);

-- Create ProfilePicture Table
CREATE TABLE ProfilePicture (
  id SERIAL PRIMARY KEY,
  user_id INT REFERENCES "User"(id) ON DELETE CASCADE,
  path VARCHAR(255) NOT NULL,
  uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);