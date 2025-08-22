module github.com/soatok/freon/coordinator

go 1.25

require github.com/taurusgroup/frost-ed25519 v0.0.0-20210707140332-5abc84a4dba7

require github.com/alexedwards/scs/v2 v2.9.0

require (
	github.com/ncruces/go-sqlite3 v0.28.0
	github.com/stretchr/testify v1.10.0
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/ncruces/julianday v1.0.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/tetratelabs/wazero v1.9.0 // indirect
	golang.org/x/sys v0.35.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/taurusgroup/frost-ed25519 => github.com/soatok/frost-ed25519 v0.0.0-20250805104728-ae78c7826e4b
