CREATE SEQUENCE website_id_seq;
CREATE TABLE IF NOT EXISTS websites (
    id INTEGER PRIMARY KEY DEFAULT nextval('website_id_seq'),
    main_url VARCHAR(255) NOT NULL,
    url_pattern VARCHAR(255) NOT NULL,
    title_pattern VARCHAR(255) NOT NULL,
    text_pattern VARCHAR(255) NOT NULL,
    process_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE SEQUENCE news_id_seq;
CREATE TABLE IF NOT EXISTS news (
    id INTEGER PRIMARY KEY DEFAULT nextval('news_id_seq'),
    website_id INTEGER NOT NULL,
    url VARCHAR(255) NOT NULL,
    title VARCHAR(255) NOT NULL,
    text TEXT NOT NULL,
    FOREIGN KEY (website_id) REFERENCES websites (id) ON DELETE CASCADE
);