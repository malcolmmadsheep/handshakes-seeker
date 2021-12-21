alter table paths
add column task_hash VARCHAR(32) unique not null;
ALTER TABLE paths
    RENAME path TO trace;
alter type path_status
add value 'cancelled';