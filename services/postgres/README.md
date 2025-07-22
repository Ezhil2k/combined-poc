# Go + PostgreSQL CRUD Example

## Prerequisites
- Go 1.20+
- PostgreSQL running on localhost:5433
- Database: `verity`, User: `gobackend_user`, Password: `strong_secure_password`

## Manual Table Creation
Run this SQL in your `verity` database before starting the app:
```sql
CREATE TABLE IF NOT EXISTS users (
  id SERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  email TEXT NOT NULL UNIQUE
);
```

## Run the App
```sh
go mod tidy
go run .
```

## Example API Usage
Create user:
```sh
curl -X POST -H "Content-Type: application/json" -d '{"name":"Alice","email":"alice@example.com"}' http://localhost:8081/users
```
Get all users:
```sh
curl http://localhost:8081/users
``` 