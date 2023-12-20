create table vending_machines
(
    id                  bigserial primary key,
    name                varchar(255) not null,
    no                  varchar(255) not null,
    location            varchar(255),
    type                varchar(255),
    status              boolean,
    last_maintenance_at timestamp(0),
    created_at          timestamp(0),
    updated_at          timestamp(0),
    deleted_at          timestamp(0)
);