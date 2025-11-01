CREATE TABLE IF NOT EXISTS movies (
    id    SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    year  INT  NOT NULL
);

CREATE TABLE IF NOT EXISTS actors (
        id       SERIAL PRIMARY KEY,
        movie_id INT NOT NULL REFERENCES movies(id) ON DELETE CASCADE,
    name     TEXT NOT NULL
    );

INSERT INTO movies (title, year) VALUES
        ('Inception', 2010),
        ('Interstellar', 2014),
        ('The Dark Knight', 2008)
    ON CONFLICT DO NOTHING;

INSERT INTO actors (movie_id, name)
SELECT id, 'Actor ' || id::text FROM movies
    ON CONFLICT DO NOTHING;
