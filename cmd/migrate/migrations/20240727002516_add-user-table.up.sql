-- Create Role Table
CREATE TABLE "role" (
  id SERIAL PRIMARY KEY,
  name VARCHAR(50) UNIQUE NOT NULL
);


-- Create User Table
CREATE TABLE "user" (
  id SERIAL PRIMARY KEY,
  name VARCHAR(100) NOT NULL,
  email VARCHAR(100) UNIQUE NOT NULL,
  password CHAR(60) NOT NULL,
  status SMALLINT NOT NULL CHECK (status IN (0, 1)), -- 0 for inactive, 1 for active
  description TEXT,
  postal_code CHAR(8),
  street VARCHAR(255) NOT NULL,
  city VARCHAR(100) NOT NULL,
  state VARCHAR(100) NOT NULL,
  cpf CHAR(11) UNIQUE NOT NULL,
  role_id INT REFERENCES role(id),
  points INT DEFAULT 0,
  registration_date DATE NOT NULL DEFAULT CURRENT_DATE,
  birth_date DATE NOT NULL
);


-- Create ProfilePicture Table
CREATE TABLE "profile_picture" (
  id SERIAL PRIMARY KEY,
  user_id INT REFERENCES "user"(id) ON DELETE CASCADE,
  path VARCHAR(255) NOT NULL,
  uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create refresh token table
CREATE TABLE "refresh_token" (
  id SERIAL PRIMARY KEY,
  user_id INT REFERENCES "user"(id) ON DELETE CASCADE,
  token VARCHAR(255) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  expires_at TIMESTAMP NOT NULL
);
