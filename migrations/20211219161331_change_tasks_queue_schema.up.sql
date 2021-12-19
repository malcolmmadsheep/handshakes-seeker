drop table if exists tasks_queue;
create table tasks_queue (
    id VARCHAR(32) PRIMARY key,
    origin_task_id VARCHAR(32),
    data_source VARCHAR(32) NOT NULL,
    source_url VARCHAR(255) not null,
    dest_url VARCHAR(255) not null,
    cursor VARCHAR(255) not null,
    created_at timestamp default current_timestamp
);