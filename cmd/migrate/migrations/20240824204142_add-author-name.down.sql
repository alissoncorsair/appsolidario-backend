DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'posts' AND column_name = 'author_name') THEN
        ALTER TABLE posts DROP COLUMN author_name;
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'comments' AND column_name = 'author_name') THEN
        ALTER TABLE comments DROP COLUMN author_name;
    END IF;
END $$;