create table if not exists users (
    username varchar(20) unique not null,
    firstname varchar(20),
    lastname varchar(20),
    description varchar(100),
    registration_timestamp timestamptz default current_timestamp,
    followers varchar(20)[] default array[]::varchar(20)[],
    following varchar(20)[] default array[]::varchar(20)[],
    tweets integer[] default array[]::integer[],
    timeline integer[] default array[]::integer[]
);

create sequence if not exists tweet_id_seq;

create table if not exists tweets (
    id bigint default nextval('tweet_id_seq'),
    parent_id integer default 0,
    creator varchar(20) not null,
    content varchar(280) not null,
    creation_timestamp timestamptz default current_timestamp
);
