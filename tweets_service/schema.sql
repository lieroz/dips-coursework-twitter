create table if not exists tweets (
    id serial primary key,
    parent_id integer default 0,
    creator varchar(20) not null,
    content varchar(280) not null,
    creation_timestamp timestamptz default current_timestamp
);
