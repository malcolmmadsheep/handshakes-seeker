ALTER TABLE tasks_queue
ADD COLUMN id serial primary key,
    ADD COLUMN body TEXT;