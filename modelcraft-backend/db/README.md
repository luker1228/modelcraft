# Database Management

## Docker Deployment (Automatic)

### How It Works

When using Docker Compose, MySQL container automatically initializes the database using SQL files from the `db/schema/mysql/` directory:

1. **MySQL Initialization**: MySQL's official Docker image automatically executes `.sql`, `.sql.gz`, and `.sh` files found in `/docker-entrypoint-initdb.d`
2. **SQL Files**: All `.sql` files in `db/schema/mysql/` are executed in alphabetical order (01_project.sql, 02_database_cluster.sql, 03_model_domain.sql)
3. **Default Project**: The init script automatically creates a `default` project for backward compatibility
4. **Fresh Start**: Works on fresh MySQL container starts only (not on container restarts with existing data)

### Configuration

SQL files are mounted via Docker Compose:

```yaml
# In docker-compose.yml
volumes:
  - ./db/schema/mysql:/docker-entrypoint-initdb.d:ro
```

### Updating the Schema

1. Edit SQL files in `db/schema/mysql/`
2. Recreate MySQL container (remove existing volume for fresh start):
```bash
docker compose down -v
docker compose up -d
```

### Local Development

For local development, you can either:
- Use Docker Compose (recommended) - automatic initialization
- Enable application-level migration via config: Set `database.migrate_on_startup: true` (default)
- Manually execute SQL files using MySQL client:
```bash
mysql -h localhost -u root -p modelcraft < db/schema/mysql/01_project.sql
mysql -h localhost -u root -p modelcraft < db/schema/mysql/02_database_cluster.sql
mysql -h localhost -u root -p modelcraft < db/schema/mysql/03_model_domain.sql
```

## Schema Files

Schema files are located in `db/schema/mysql/`:
```
db/schema/mysql/
├── 01_project.sql           # Projects table with default project
├── 02_database_cluster.sql  # Database clusters table
└── 03_model_domain.sql      # Models, fields, relations, enums tables
```

The initialization script creates:
- `projects` table with default project
- `database_clusters` table
- `models` table
- `field_definitions` table
- `model_relations` table
- `model_enums` table
- `model_field_enum_associations` table

All tables support project isolation through `project_id` foreign key.

## Application-Level Migration (Optional)

The application can run migrations on startup instead of relying on Docker's init scripts:

### Configuration

Set the migration flag in your config file (configs/config.yaml or config.docker.yaml):

```yaml
database:
  driver: mysql
  host: localhost
  port: 3306
  username: modelcraft
  password: modelcraft123
  database: modelcraft
  migrate_on_startup: true  # Enable/disable startup migration
```

### How It Works

1. Application connects to MySQL server (without specifying database)
2. Creates the target database if it doesn't exist (`CREATE DATABASE IF NOT EXISTS`)
3. Loads all `*.sql` files from `db/schema/mysql/` in order
4. Executes each SQL file against the database
5. Logs progress and errors

### Use Cases

- Local development without Docker
- Production deployments where database init scripts are not preferred
- Environments where container recreation with `docker compose down -v` is not desired

### Important Notes

- SQL files use `CREATE TABLE IF NOT EXISTS` for idempotency
- Migration logs are visible in application console output
- Migration failures will cause the application to exit with an error

## Manual Migration with Atlas CLI (Advanced)

For advanced schema management, you can use Atlas CLI directly:

### Install Atlas

```bash
make install-atlas
```

### Apply Schema Manually

```bash
atlas schema apply \
  -u "mysql://root:password@127.0.0.1:3306/modelcraft" \
  --to file://db/schema/mysql/ \
  --auto-approve
```

### Generate Migration Diff

```bash
atlas migrate diff initial \
  --to file://db/schema/mysql/ \
  --dev-url "mysql://root:password@127.0.0.1:3306/modelcraft_dev"
```

## Reference Files

- `db/schema/mysql/*.sql` - SQL schema files (source for Docker init and app migration)
