# visa
token issue

## database migration
brew install golang-migrate

### create migration
migrate create -ext sql -dir infra/pg/migrations -seq init_tables
