CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

drop table if exists books;
drop table if exists authors;

create table authors (
    id uuid DEFAULT uuid_generate_v4(),
    name varchar not null,
    created_at timestamp not null,
    primary key(id)
);

create table books (
    id uuid DEFAULT uuid_generate_v4(),
    name varchar not null,
    genre varchar not null,
    created_at timestamp not null,
    author_id uuid,
    primary key (id),
    constraint fk_author
        foreign key (author_id)
            references authors(id)
);