module github.com/soatok/freon/client

go 1.25

require (
	filippo.io/age v1.2.1
	github.com/taurusgroup/frost-ed25519 v0.0.0-20210707140332-5abc84a4dba7
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	golang.org/x/crypto v0.41.0 // indirect
	golang.org/x/sys v0.35.0 // indirect
)

replace github.com/taurusgroup/frost-ed25519 => github.com/soatok/frost-ed25519 v0.0.0-20250805104728-ae78c7826e4b
