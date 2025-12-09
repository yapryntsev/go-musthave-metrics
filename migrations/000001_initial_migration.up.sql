CREATE TYPE metric_type AS enum ('counter', 'gauge');

CREATE TABLE metrics
(
    id    TEXT primary key not null,
    type  metric_type      not null,
    delta integer,
    value double precision
);