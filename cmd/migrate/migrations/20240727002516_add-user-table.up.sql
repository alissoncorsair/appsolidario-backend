-- Create Role Table
CREATE TABLE "roles" (
  id SERIAL PRIMARY KEY,
  name VARCHAR(50) UNIQUE NOT NULL
);

-- Insert default roles role payee 1 role payer 2
INSERT INTO "roles" (name) VALUES ('payee'), ('payer');


-- Create User Table
CREATE TABLE "users" (
  id SERIAL PRIMARY KEY,
  name VARCHAR(100) NOT NULL,
  surname VARCHAR(100) NOT NULL,
  email VARCHAR(100) UNIQUE NOT NULL,
  password CHAR(60) NOT NULL,
  status SMALLINT NOT NULL CHECK (status IN (0, 1)), -- 0 for inactive, 1 for active
  description TEXT,
  postal_code CHAR(8),
  city VARCHAR(100) NOT NULL,
  state VARCHAR(100) NOT NULL,
  cpf CHAR(11) UNIQUE NOT NULL,
  role_id INT REFERENCES "roles"(id),
  points INT DEFAULT 0,
  birth_date DATE NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);



-- Create ProfilePicture Table
CREATE TABLE "profile_pictures" (
  id SERIAL PRIMARY KEY,
  user_id INT REFERENCES "users"(id) ON DELETE CASCADE,
  path VARCHAR(255) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_user_id ON "profile_pictures"(user_id);


-- Create refresh token table
-- CREATE TABLE "refresh_tokens" (
--   id SERIAL PRIMARY KEY,
--   user_id INT REFERENCES "users"(id) ON DELETE CASCADE,
--   token VARCHAR(255) NOT NULL,
--   created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
--   expires_at TIMESTAMP NOT NULL
-- );
