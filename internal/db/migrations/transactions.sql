create table transactions
(
    id bigserial primary key,
    txn_id     bigint  not null,
    txn_date   integer,
    result     integer               not null,
    sum        double precision      not null,
    comment    varchar(255),
    status     boolean default false not null,
    created_at timestamp(0),
    updated_at timestamp(0)
);