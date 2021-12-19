create type path_status as enum (
    'not_started',
    'in_progress',
    'found',
    'not_found'
);
create table if not exists paths (
    id serial primary key,
    data_source VARCHAR(32) not null,
    source_url varchar(255) not null,
    destination_url VARCHAR(255),
    status path_status,
    path TEXT
);