# postgresql-migrations

> Example migrations - naming convention and up/down pattern (golang-migrate)

## app/shared/infrastructure/postgresql/migrations/000001_initial_schema.up.sql

```sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS sample_table (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS template_table (
    id VARCHAR(36) PRIMARY KEY
);
```

---

## app/shared/infrastructure/postgresql/migrations/000001_initial_schema.down.sql

```sql
DROP TABLE IF EXISTS "sample_table";
DROP TABLE IF EXISTS "template_table";
```
