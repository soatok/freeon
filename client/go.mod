module github.com/soatok/freon/client

go 1.24

require (
	filippo.io/age v1.2.1
	github.com/taurusgroup/frost-ed25519 v0.0.0-00010101000000-000000000000
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	golang.org/x/crypto v0.24.0 // indirect
	golang.org/x/sys v0.21.0 // indirect
)

replace github.com/taurusgroup/frost-ed25519 => github.com/soatok/frost-ed25519 v0.0.0-20250805104728-ae78c7826e4b
