package main

// This file contains the basic help text for each subcommand

const usage = `FREEON - FOSS Resists Extralegal Executive Overreaching Nations

USAGE:
    freeon [OPTIONS] <COMMAND>

DESCRIPTION:
    Distributed Ed25519 signature generation using FROST (RFC 9591).
    Enables geographically distributed teams to collaboratively generate
    digital signatures without exposing private key material.

OPTIONS:
    -h, --help       Print help information
    -V, --version    Print version information
    -v, --verbose    Enable verbose output

COMMANDS:
    keygen       Distributed key generation ceremonies
    sign         Signature generation ceremonies  
    terminate    Terminate incomplete ceremonies
    help         Print this message or the help of the given subcommand(s)

Use 'freeon <COMMAND> --help' for more information on a specific command.

EXAMPLES:
    freeon keygen create -h coordinator:8080 -n 5 -t 3
    freeon sign create -g abc123 message.txt
    freeon sign join -c ceremony456

`

const keygenUsage = `
freeon KEYGEN - Distributed Key Generation

USAGE:
    freeon keygen <SUBCOMMAND>

SUBCOMMANDS:
    create    Initialize a new DKG ceremony
    join      Join an existing DKG ceremony
    list      List local key shares and groups
    help      Print this message or the help of the given subcommand(s)
`

const keygenCreateUsage = `freeon KEYGEN CREATE - Initialize DKG ceremony

USAGE:
    freeon keygen create [OPTIONS] -h <HOST> -n <PARTICIPANTS> -t <THRESHOLD>

DESCRIPTION:
    Creates a new distributed key generation ceremony. Returns a Group ID
    that other participants use to join the ceremony.

OPTIONS:
    -h, --host <HOST>              Coordinator hostname:port
    -n, --participants <NUM>       Total number of participants (2-255)
    -t, --threshold <NUM>          Minimum signatures required (1 to n)
    -r, --recipient <PUBKEY>       Age/SSH public key to encrypt share
        --help                     Print help information

EXAMPLES:
    freeon keygen create -h coord.example.com:8080 -n 7 -t 3
    freeon keygen create -h 192.168.1.100:8080 -n 5 -t 3 -r age1abc...

`

const keygenJoinUsage = `
freeon KEYGEN JOIN - Join DKG ceremony

USAGE:
    freeon keygen join [OPTIONS] -h <HOST> -g <GROUP_ID>

DESCRIPTION:
    Join an existing DKG ceremony using the Group ID from the creator.
    Maintains connection until all participants join and key is generated.

OPTIONS:
    -h, --host <HOST>         Coordinator hostname:port  
    -g, --group <GROUP_ID>    Group ID from ceremony creator
    -r, --recipient <PUBKEY>  Age/SSH public key to encrypt share
        --help                Print help information

EXAMPLES:
    freeon keygen join -h coord.example.com:8080 -g grp_abc123def456
    freeon keygen join -h coord.example.com:8080 -g grp_xyz789 -r ~/.ssh/id_ed25519.pub

`

const signUsage = `freeon SIGN - Signature Generation

USAGE:
    freeon sign <SUBCOMMAND>

SUBCOMMANDS:
    create    Initialize a new signature ceremony
    join      Join an existing signature ceremony
    list      List recent signing ceremonies
    help      Print this message or the help of the given subcommand(s)

`

const signCreateUsage = `freeon SIGN CREATE - Initialize signature ceremony

USAGE:
    freeon sign create [OPTIONS] -g <GROUP_ID> [MESSAGE]

DESCRIPTION:
    Creates a new signature ceremony for the specified group and message.
    Returns a Ceremony ID that participants use to join. Does not require
    holding a key share (useful for CI/CD automation).

ARGUMENTS:
    [MESSAGE]    File containing message to sign (use '-' for stdin)

OPTIONS:
    -g, --group <GROUP_ID>    Group ID from DKG ceremony
    -h, --host <HOST>         Coordinator hostname:port
        --help                Print help information
    --openssh                 Return an OpenSSH formatted signature
    --namespace <NAMESPACE>   Specify a namespace for OpenSSH (default: "file")

EXAMPLES:
    freeon sign create -g grp_abc123 message.txt
    echo "Hello World" | freeon sign create -g grp_abc123 -
    freeon sign create -g grp_abc123  --openssh --namespace git release.tar.gz

`

const signListUsage = `freeon SIGN LIST - List recent signing ceremonies

USAGE:
    freeon sign list [OPTIONS] -g <GROUP_ID>

DESCRIPTION:
    List recent signing ceremonies for the specified group. Results are
    paginated and can be filtered using limit and offset parameters.

OPTIONS:
    -g, --group <GROUP_ID>    Group ID to list ceremonies for
    -h, --host <HOST>         Coordinator hostname:port
        --limit <NUM>         Maximum number of ceremonies to return
        --offset <NUM>        Number of ceremonies to skip (for pagination)
        --help                Print help information

EXAMPLES:
    freeon sign list -g grp_abc123
    freeon sign list -g grp_abc123 --limit 10
    freeon sign list -g grp_abc123 --limit 20 --offset 40
    freeon sign list -g grp_abc123 -h coord.example.com:8080 --limit 5

`

const signJoinUsage = `freeon SIGN JOIN - Join signature ceremony  

USAGE:
    freeon sign join [OPTIONS] -c <CEREMONY_ID> [MESSAGE]

DESCRIPTION:
    Join an existing signature ceremony. The message must match what was
    specified during ceremony creation (used for verification).

ARGUMENTS:
    [MESSAGE]    File containing message to sign (use '-' for stdin)

OPTIONS:
    -c, --ceremony <CEREMONY_ID>    Ceremony ID from sign create
    -h, --host <HOST>               Coordinator hostname:port
    -i, --identity <FILE>           Path to age secret keys file
        --help                      Print help information

EXAMPLES:
    freeon sign join -c cer_def456 message.txt
    echo "Hello World" | freeon sign join -c cer_def456 -
    freeon sign join -c cer_def456 -i ~/.age/keys.txt message.txt

`

const signGetUsage = `freeon SIGN GET - Get signature from coordinator

USAGE:
    freeon sign get [OPTIONS] -c <CEREMONY_ID>

DESCRIPTION:
    Query the coordinator for the final signature for a concluded
    ceremony.

OPTIONS:
    -c, --ceremony <CEREMONY_ID>    Ceremony ID from sign create
    -h, --host <HOST>               Coordinator hostname:port
        --help                      Print help information

EXAMPLES:
    freeon sign get -c cer_def456
    freeon sign get -h coord.example.com:8080 -c cer_def456 message.txt

`

const terminateUsage = `freeon TERMINATE - Terminate ceremonies

USAGE:
    freeon terminate <CEREMONY_ID>

DESCRIPTION:
    Terminate an incomplete signature ceremony. Does not require
    privileged access - any user can terminate any ceremony.

ARGUMENTS:
    <CEREMONY_ID>    Ceremony ID to terminate

OPTIONS:
        --help    Print help information

EXAMPLES:
    freeon terminate cer_def456

`
