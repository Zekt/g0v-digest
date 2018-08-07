## Dependency
Go, PostgreSQL

## Build
`go build main.go Routers.go Models.go Helpers.go Structs.go`

`psql -U <username> -d <dbname> < scheme.sql`

## Config
change file name of `config.json.example` to `config.json`

## API Endpoints
PUT `/medium` - Check g0v.news feed for update and write to database.

GET `/api/line` - XML RSS for LINE Today.

PUT `/api/line/tick` - Update the UUID of XML.

PUT `/mailchimp` - Set up a new campaign with the newest post in database.
