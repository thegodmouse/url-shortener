CREATE DATABASE IF NOT EXISTS url_shortener;

USE url_shortener;

CREATE TABLE IF NOT EXISTS short_urls
(
    id         INTEGER                             NOT NULL AUTO_INCREMENT,
    url        VARCHAR(2083)                       NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    expire_at  TIMESTAMP                           NOT NULL,
    is_deleted BOOLEAN   DEFAULT FALSE             NOT NULL,
    PRIMARY KEY (id)
);