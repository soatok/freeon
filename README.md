# FREON

> FOSS Resists Executive Overreaching Nations

[![Build Status](https://github.com/soatok/freon/actions/workflows/ci.yml/badge.svg)](https://github.com/soatok/freon/actions/workflows/ci.yml)

FREON implements FROST ([RFC 9591](https://www.rfc-editor.org/rfc/rfc9591.html)) to allow geographically distributed teams produce digital signatures.

Each share of the signing key is encrypted locally using [age](https://github.com/FiloSottile/age).

> [!WARNING]
> This software is a minimum viable product (MVP) and is only in the **alpha** stage of development. It has not been audited. Do not use this in production yet!

## Installation

### Freon Clients

> [!WARNING]
> This command currently does not work. See [below](#temporary-workaround-for-replace-directive-error).

```terminal
go install github.com/soatok/freon/client@latest
```

#### For Developers

You can also clone the repository to install the client locally:

```terminal
git clone https://github.com/soatok/freon.git
cd freon/client
go install
```

### Freon Coordinators

> [!WARNING]
> The coordinator is expected to run on a private network, such as [Tailscale](https://tailscale.com),
> [ZeroTier](https://www.zerotier.com), or within a Virtual Private Cloud (VPC) from AWS, Azure, or GCP.
>
> In the future, we will develop a standalone coordinator that can safely be deployed on the public-facing
> Internet, but this was cut from the alpha release due to time constraints.

```terminal
git clone https://github.com/soatok/freon.git
cd freon
go build -o coordinator ./coordinator
./coordinator
```

## Usage

The order of operations is as followed:

1. Perform the **Distributed Key Generation** ceremony once, for each Ed25519 keypair.
   1. One client tell the coordinator to initiate a new DKG. The number of parties and threshold are required at this step.
   2. The coordinator sets up a session that other clients can connect to, and gives the DKG Group ID to the client. This is to be shared with other users. The Group ID is not sensitive; it only serves to allow multiple keys be managed by one Freon coordinator.
   3. Once every participating user enrolls in the DKG ceremony, the final public key is calculated and shared with each participant.
   4. Each client encrypts their Shamir Share, along with their Party ID and the group identifier, locally.
2. For each message to be signed:
   1. One client proposes a message to be signed to the coordinator.
   2. Each participating client connects and commits to the message (FROST Round 1).
   3. Each participating client connects and publishes their share of the final signature (FROST Round 2).
   4. The coordinator aggregates the signature and releases it to each connected client.

We will document each command in tandem with this order of operations.

### Distributed Key Generation

To initiate a new DKG group, one of the clients with access to the coordinator will run the following command (replace 7 and 3 with
the number of participants and threshold needed to perform a signature operation, respectively):

```terminal
freon keygen create -h hostname:port -n 7 -t 3
```

Upon success, a Group ID will be returned. This is to be shared with the other participants, who will pass it as an extra argument:

```terminal
freon keygen join -h hostname:port -g [group-id-goes-here]
```

This will maintain a connection with the coordinator until all `n` participants have connected. Afterwards, a copy of the public key will be returned to each client.

#### Optional Arguments

If you provide a public key (`-r [RECIPIENT]`) as an optional argument, the Freon client will use [age](https://age-encryption.org) to encrypt the share locally. This public key can be an age public key or an OpenSSH public key.

### Signature Generation

#### Initiate Signature Ceremony

To initiate a key ceremony, the following information is needed:

1. The Group ID used during Key Generation.
2. The message intended to be signed.

The message can be passed as a file name or via `STDIN`, like so:

```terminal
freon sign create -g [group-id-goes-here] file-with-message.txt
echo -n "MESSAGE TO BE SIGNED" | freon sign create -g [group-id-goes-here]
```

This will initialize a signature in progress and return a Ceremony ID.

> [!NOTE]
> You do not need a key share to initiate a signature proposal. This is an intentional design feature to allow
> CI/CD pipelines queue up a Signature Ceremony that clients can then opt into participating in.

##### Optional Arguments

By default, the final signature will be returned as a 128-character hex-encoded string (encoded as `R || z`).
You can pass an optional `--openssh` flag to return an OpenSSH-compatible signature.

```terminal
freon sign create --openssh -g [group-id-goes-here] file-with-message.txt
echo -n "MESSAGE TO BE SIGNED" | freon sign create --openssh -g [group-id-goes-here]
```

##### Terminating Incomplete Ceremonies

You can run this command to flush any incomplete ceremonies.

```terminal
freon terminate [ceremony-id-goes-here]
```

> [!WARNING]
> You do not need privileged access to perform this step.

#### Participate In Signature Ceremony

Each client will need to run this command to participate in the ceremony.

```terminal
freon sign join -c [ceremony-id] file-with-message.txt
echo -n "MESSAGE TO BE SIGNED" | freon sign join -c [ceremony-id]

# Identical:
freon sign join --ceremony [ceremony-id] file-with-message.txt
echo -n "MESSAGE TO BE SIGNED" | freon sign join --ceremony [ceremony-id]
```

##### Optional Arguments

You can furthermore pass the `-i` or `--identity` flag to specify the file path for your age secret keys.

```terminal
freon sign join -c [ceremony-id] -i /path/to/age.keys file-with-message.txt
echo -n "MESSAGE TO BE SIGNED" | freon sign join -i /path/to/age.keys -c [ceremony-id]

# Identical
freon sign join --ceremony [ceremony-id] --identity /path/to/age.keys file-with-message.txt
echo -n "MESSAGE TO BE SIGNED" | freon sign join --identity /path/to/age.keys --ceremony [ceremony-id]
```
