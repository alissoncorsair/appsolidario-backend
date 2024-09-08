-- Step 1: Add author_name column to the existing tables
ALTER TABLE posts
ADD COLUMN author_name VARCHAR(200) NOT NULL;

ALTER TABLE comments
ADD COLUMN author_name VARCHAR(200) NOT NULL;

-- Step 2: Temporarily remove foreign key constraints referencing posts table
ALTER TABLE post_photos DROP CONSTRAINT post_photos_post_id_fkey;
ALTER TABLE comments DROP CONSTRAINT comments_post_id_fkey;

-- Step 3: Create a new posts table with author_name in the third position
CREATE TABLE posts_new (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    author_name VARCHAR(200) NOT NULL,  -- Moved to third position
    title VARCHAR(200) NOT NULL,
    description TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Step 4: Migrate data from old posts table to the new posts table
-- Using `u.name` as author_name from the users table
INSERT INTO posts_new (id, user_id, author_name, title, description, created_at, updated_at)
SELECT p.id, p.user_id, u.name, p.title, p.description, p.created_at, p.updated_at
FROM posts p
JOIN users u ON p.user_id = u.id;

-- Step 5: Rename the old posts table and the new posts table
ALTER TABLE posts RENAME TO posts_old;
ALTER TABLE posts_new RENAME TO posts;

-- Step 6: Recreate foreign key constraints on post_photos and comments to reference the new posts table
ALTER TABLE post_photos
ADD CONSTRAINT post_photos_post_id_fkey FOREIGN KEY (post_id) REFERENCES posts(id);

ALTER TABLE comments
ADD CONSTRAINT comments_post_id_fkey FOREIGN KEY (post_id) REFERENCES posts(id);

-- Step 7: Drop the old posts table (optional)
DROP TABLE posts_old;


-- Step 8: Create a new comments table with author_name in the correct position
CREATE TABLE comments_new (
    id SERIAL PRIMARY KEY,
    post_id INT NOT NULL,
    user_id INT NOT NULL,
    author_name VARCHAR(200) NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Step 9: Migrate data from old comments table to the new comments table
-- Using `u.name` as author_name from the users table
INSERT INTO comments_new (id, post_id, user_id, author_name, content, created_at, updated_at)
SELECT c.id, c.post_id, c.user_id, u.name, c.content, c.created_at, c.updated_at
FROM comments c
JOIN users u ON c.user_id = u.id;

-- Step 10: Rename the old comments table and the new comments table
ALTER TABLE comments RENAME TO comments_old;
ALTER TABLE comments_new RENAME TO comments;

-- Step 11: Drop the old comments table (optional)
DROP TABLE comments_old;
