create table balances
(
    id                 bigserial
        primary key,
    vending_machine_id bigint        not null
        constraint balances_vending_machine_id_foreign
        references vending_machines
        on delete cascade,
    transaction_id     bigint        not null
        constraint balances_transaction_id_foreign
        references transactions
        on delete cascade,
    amount             numeric(8, 2) not null,
    type               boolean       not null,
    created_at         timestamp(0),
    updated_at         timestamp(0)
);