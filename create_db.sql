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
    publication_date date not null,
    created_at timestamp not null,
    author_id uuid,
    primary key (id),
    constraint fk_author
        foreign key (author_id)
            references authors(id)
);


WITH new_author AS (
    insert into authors (name, created_At)
    values ('Lovecraft', current_timestamp) returning id
) 
INSERT INTO books (name, genre, publication_date, created_at, author_id)
    SELECT 'Шепчущий во тьме', 'Хоррор', '1920-10-02', current_timestamp, id
        FROM new_author;
INSERT INTO books (name, genre, publication_date, created_at, author_id)
    SELECT 'Зов Ктулху', 'Хоррор', '1921-10-02', current_timestamp, id
        FROM authors where name = 'Lovecraft';