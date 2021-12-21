alter table paths drop column task_hash;
ALTER TABLE paths
    RENAME trace TO path;