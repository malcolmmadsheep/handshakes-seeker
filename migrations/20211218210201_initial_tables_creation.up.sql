create table if not exists connections (
    id serial primary key,
    data_source VARCHAR(32) not null,
    source_url VARCHAR(255) not null,
    destination_url VARCHAR(255)
);