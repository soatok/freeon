module github.com/soatok/freon/coordinator

go 1.25

require github.com/taurusgroup/frost-ed25519 v0.0.0-20210707140332-5abc84a4dba7

require github.com/alexedwards/scs/v2 v2.9.0

require github.com/mattn/go-sqlite3 v1.14.32

require filippo.io/edwards25519 v1.1.0 // indirect

replace github.com/taurusgroup/frost-ed25519 => github.com/soatok/frost-ed25519 v0.0.0-20250805104728-ae78c7826e4b
