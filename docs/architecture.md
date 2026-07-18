- **Contain change through isolation, not abstraction.** The "database-agnostic" NFR was debated and *dropped*. The durable principle is a storage layer with a small blast radius — not a costly upfront abstraction to swap databases we'll never swap.
- **Global feeds + subscriptions join table.** Feeds are stored once in a shared table; a subscriptions join table links users to feeds. Per-user feed duplication was explicitly rejected.
- **Fail-fast on config.** Missing or malformed env vars must cause immediate startup failure — never silent misbehavior or defaults that mask a broken setup.

## Data model
```mermaid
erDiagram
    USERS {
        bigint id PK "GENERATED ALWAYS AS IDENTITY"
        text username UK "NOT NULL"
        text email UK "NOT NULL; UNIQUE on lower(email) via expression index"
        timestamptz created_at "NOT NULL DEFAULT now()"
        timestamptz updated_at "NOT NULL"
    }

    FEED {
        bigint id PK "GENERATED ALWAYS AS IDENTITY"
        text feed_link UK "NOT NULL; RSS feed URL"
        text page_link; homepage URL"
        text title "NOT NULL"
        timestamptz created_at "NOT NULL DEFAULT now()"
        timestamptz fetched_at "NULL until first fetch"
        timestamptz updated_at "feed contents update time, per publisher"
    }

    ITEM {
        bigint id PK "GENERATED ALWAYS AS IDENTITY"
        bigint feed_id FK, UK "UNIQUE(feed_id, guid)"
        text guid UK "NOT NULL; fallback = item link when feed omits GUID"
        text title "NOT NULL"
        text description
        text link "NOT NULL"
        timestamptz published_at "NULL — feeds omit or lie; sort fallback = created_at"
        timestamptz created_at "NOT NULL DEFAULT now(); when fetched first"
        timestamptz updated_at "item update time, per publisher"
    }

    SUBSCRIPTION {
        bigint user_id PK, FK "composite PK(user_id, feed_id)"
        bigint feed_id PK, FK "composite PK(user_id, feed_id)"
        timestamptz created_at "NOT NULL DEFAULT now()"
    }

    READSTATUS {
        bigint user_id PK, FK "composite PK(user_id, item_id)"
        bigint item_id PK, FK "composite PK(user_id, item_id)"
        timestamptz read_at "NOT NULL DEFAULT now()"
    }


    USERS ||--o{ SUBSCRIPTION : "subscribes"
    FEED ||--o{ SUBSCRIPTION : "gets subscribed"
    FEED ||--o{ ITEM : "contains"
    USERS ||--o{ READSTATUS : "reads"
    ITEM ||--o{ READSTATUS : "gets read"
```

For item.guid, if guid field is empty in a retrieved item, use item.link.
To sort items use item.updated_at -> item.published_at -> item.created_at. This is a fallback sequence if some fields are not defined. Might ignore item.updated_at in the future if find it too annoying.

### DDL
```SQL
CREATE TABLE IF NOT EXISTS users (
    id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    username text UNIQUE NOT NULL,
    email text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS users_lower_email_idx ON users (lower(email));

CREATE TABLE IF NOT EXISTS feeds (
    id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    feed_link text UNIQUE NOT NULL, -- RSS feed URL
    page_link text, -- homepage URL
    title text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    fetched_at timestamptz, -- NULL until first fetch
    updated_at timestamptz -- feed contents update time, per publisher
);

CREATE TABLE IF NOT EXISTS items (
    id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    feed_id bigint NOT NULL REFERENCES feeds(id) ON DELETE CASCADE,
    guid text NOT NULL,
    title text NOT NULL,
    description text,
    link text NOT NULL, 
    created_at timestamptz NOT NULL DEFAULT now(),
    published_at timestamptz,
    updated_at timestamptz, -- item update time, per publisher

    CONSTRAINT feed_id_guid UNIQUE (feed_id, guid)
);

CREATE TABLE IF NOT EXISTS subscriptions (
    user_id bigint REFERENCES users(id) ON DELETE CASCADE,
    feed_id bigint REFERENCES feeds(id) ON DELETE CASCADE,
    created_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT user_id_feed_id PRIMARY KEY (user_id, feed_id)
);

CREATE TABLE IF NOT EXISTS read_status (
    user_id bigint REFERENCES users(id) ON DELETE CASCADE,
    item_id bigint REFERENCES items(id) ON DELETE CASCADE,
    read_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT user_id_item_id PRIMARY KEY (user_id, item_id)
);
```

## Open questions
- Should HTML in titles, descriptions and contents be escaped? Sanitize them (e.g. with bluemonday Go library) when displaying.