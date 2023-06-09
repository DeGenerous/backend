DROP FUNCTION IF EXISTS get_prompt(category varchar);
DROP TABLE IF EXISTS prompts;

DROP FUNCTION IF EXISTS get_image(storyId varchar(42));
DROP TABLE IF EXISTS stories;

DROP FUNCTION IF EXISTS get_nonce(wallet varchar);
DROP INDEX IF EXISTS user_wallet;
DROP TABLE IF EXISTS users;

create table users (
    id bigserial primary key unique,
    wallet varchar(42) unique not null,
    nonce varchar(36) not null,
    email varchar not null default '',
    display_name varchar(100) not null default '',
    profile_picture varchar default '',
    created timestamp default now(),
    used_bonus int default 0,
    bonus int default 0
);

create index user_wallet on users(wallet);

CREATE FUNCTION get_nonce(wallet varchar(42))
RETURNS varchar(36)
LANGUAGE plpgsql
AS $$
DECLARE num int;
    ret varchar;
BEGIN
    SELECT count(id) into num FROM users WHERE users.wallet = get_nonce.wallet;
    if num != 1 then
        return NULL;
    end if;

    SELECT nonce into ret from users WHERE users.wallet = get_nonce.wallet;
    return ret;
END
$$;

create table stories (
    id varchar(42) primary key unique not null,
    user_id bigint references users(id) not null,
    step int default 1,
    image boolean default false,
    bonus bool default false,
    created timestamp default now()
);

CREATE FUNCTION get_image(storyId varchar(42))
    RETURNS boolean
    LANGUAGE plpgsql
AS $$
DECLARE ret boolean;
BEGIN
    SELECT image into ret FROM stories WHERE id = get_image.storyId;
    UPDATE stories SET image = true WHERE id = get_image.storyId;
    return ret;
END
$$;

create table prompts (
    id bigserial primary key unique not null,
    category varchar not null,
    prompt text not null
);

CREATE FUNCTION get_prompt(category varchar)
    RETURNS RECORD
    LANGUAGE plpgsql
AS $$
DECLARE ret RECORD;
BEGIN
    SELECT id, prompt into ret FROM prompts WHERE prompts.category = get_prompt.category ORDER BY random() limit 1;
    return ret;
END
$$;
