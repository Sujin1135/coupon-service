create table coupons
(
    id           varchar(36)                         not null
        primary key,
    name         varchar(20)                         not null,
    issue_amount int                                 not null,
    issued_at    timestamp                           not null,
    expires_at   timestamp                           not null,
    created_at   timestamp default CURRENT_TIMESTAMP not null,
    modified_at  timestamp default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP,
    deleted_at   timestamp                           null
);

create table issued_coupons
(
    id          varchar(36)                         not null
        primary key,
    coupon_id   varchar(36)                         not null,
    code        varchar(10)                         not null,
    created_at  timestamp default CURRENT_TIMESTAMP not null,
    modified_at timestamp default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP,
    deleted_at  timestamp                           null,
    constraint issued_coupons_coupon_id_code_uniq
        unique (coupon_id, code),
    constraint issued_coupons_coupon_id_fk
        foreign key (coupon_id) references coupons (id)
);
