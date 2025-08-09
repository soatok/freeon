package main

// This file contains the basic help text for each subcommand

const usage = `FREON - FOSS Resists Executive Overreaching Nations

USAGE:
    freon [OPTIONS] <COMMAND>

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

Use 'freon <COMMAND> --help' for more information on a specific command.

EXAMPLES:
    freon keygen create -h coordinator:8080 -n 5 -t 3
    freon sign create -g abc123 message.txt
    freon sign join -c ceremony456

`

const keygenUsage = `
FREON KEYGEN - Distributed Key Generation

USAGE:
    freon keygen <SUBCOMMAND>

SUBCOMMANDS:
    create    Initialize a new DKG ceremony
    join      Join an existing DKG ceremony
    list      List local key shares and groups
    help      Print this message or the help of the given subcommand(s)
`

const keygenCreateUsage = `FREON KEYGEN CREATE - Initialize DKG ceremony

USAGE:
    freon keygen create [OPTIONS] -h <HOST> -n <PARTICIPANTS> -t <THRESHOLD>

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
    freon keygen create -h coord.example.com:8080 -n 7 -t 3
    freon keygen create -h 192.168.1.100:8080 -n 5 -t 3 -r age1abc...

`

const keygenJoinUsage = `
FREON KEYGEN JOIN - Join DKG ceremony

USAGE:
    freon keygen join [OPTIONS] -h <HOST> -g <GROUP_ID>

DESCRIPTION:
    Join an existing DKG ceremony using the Group ID from the creator.
    Maintains connection until all participants join and key is generated.

OPTIONS:
    -h, --host <HOST>         Coordinator hostname:port  
    -g, --group <GROUP_ID>    Group ID from ceremony creator
    -r, --recipient <PUBKEY>  Age/SSH public key to encrypt share
        --help                Print help information

EXAMPLES:
    freon keygen join -h coord.example.com:8080 -g grp_abc123def456
    freon keygen join -h coord.example.com:8080 -g grp_xyz789 -r ~/.ssh/id_ed25519.pub

`

const signUsage = `FREON SIGN - Signature Generation

USAGE:
    freon sign <SUBCOMMAND>

SUBCOMMANDS:
    create    Initialize a new signature ceremony
    join      Join an existing signature ceremony
    list      List active signature ceremonies
    help      Print this message or the help of the given subcommand(s)

`

const signCreateUsage = `FREON SIGN CREATE - Initialize signature ceremony

USAGE:
    freon sign create [OPTIONS] -g <GROUP_ID> [MESSAGE]

DESCRIPTION:
    Creates a new signature ceremony for the specified group and message.
    Returns a Ceremony ID that participants use to join. Does not require
    holding a key share (useful for CI/CD automation).

ARGUMENTS:
    [MESSAGE]    File containing message to sign (use '-' for stdin)

OPTIONS:
    -g, --group <GROUP_ID>    Group ID from DKG ceremony
    -h, --host <HOST>         Coordinator hostname:port
        --openssh             Return OpenSSH-compatible signature format
        --help                Print help information

EXAMPLES:
    freon sign create -g grp_abc123 message.txt
    echo "Hello World" | freon sign create -g grp_abc123 -
    freon sign create --openssh -g grp_abc123 release.tar.gz

`

const signJoinUsage = `FREON SIGN JOIN - Join signature ceremony  

USAGE:
    freon sign join [OPTIONS] -c <CEREMONY_ID> [MESSAGE]

DESCRIPTION:
    Join an existing signature ceremony. The message must match what was
    specified during ceremony creation (used for verification).

ARGUMENTS:
    [MESSAGE]    File containing message to sign (use '-' for stdin)

OPTIONS:
    -c, --ceremony <CEREMONY_ID>    Ceremony ID from sign create
    -h, --host <HOST>               Coordinator hostname:port
    -i, --identity <FILE>           Path to age secret keys file
        --auto-confirm              Skip message confirmation prompt
        --help                      Print help information

EXAMPLES:
    freon sign join -c cer_def456 message.txt
    echo "Hello World" | freon sign join -c cer_def456 -
    freon sign join -c cer_def456 -i ~/.age/keys.txt message.txt
    freon sign join -c cer_def456 --auto-confirm message.txt

`

const terminateUsage = `FREON TERMINATE - Terminate ceremonies

USAGE:
    freon terminate <CEREMONY_ID>

DESCRIPTION:
    Terminate an incomplete signature ceremony. Does not require
    privileged access - any user can terminate any ceremony.

ARGUMENTS:
    <CEREMONY_ID>    Ceremony ID to terminate

OPTIONS:
        --help    Print help information

EXAMPLES:
    freon terminate cer_def456

`
