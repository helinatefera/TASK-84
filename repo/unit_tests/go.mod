module github.com/localinsights/portal/unit_tests

go 1.25.0

require github.com/localinsights/portal v0.0.0

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/go-sql-driver/mysql v1.9.3 // indirect
	github.com/golang-jwt/jwt/v5 v5.3.1 // indirect
	github.com/jmoiron/sqlx v1.4.0 // indirect
	golang.org/x/crypto v0.50.0 // indirect
	golang.org/x/sys v0.43.0 // indirect
)

replace github.com/localinsights/portal => ../backend
