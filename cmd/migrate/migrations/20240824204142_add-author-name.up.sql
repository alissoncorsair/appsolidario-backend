ALTER TABLE posts
ADD COLUMN author_name VARCHAR(200) NOT NULL;

ALTER TABLE comments
ADD COLUMN author_name VARCHAR(200) NOT NULL;