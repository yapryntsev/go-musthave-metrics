CREATE TYPE metric_type AS enum ('counter', 'gauge');

CREATE TABLE metrics
(
    id    TEXT             not null,
    type  metric_type      not null,
    delta BIGINT,
    value DOUBLE PRECISION,
    PRIMARY KEY (id, type)
);