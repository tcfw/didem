<center>
<h1>DIDEm</h1>

A simple distributed storage register and p2p messaging service for Decentralised Identities (DID) and Verifiable Claims (VC).
</center>

# State

Very much NOT production ready.

# Components

### Storage

The underlying storage mechanism for the blockchain data is [IPFS (Interplanetary File System)](https://ipfs.io/). IPFS has been chosen for its content-based addressing which is perfect for storing blockchain components.

### Node Communication

[libp2p](https://libp2p.io/) provides the multiplexing and p2p components for all internode communication. For the blockchain, gossipsub messaging relays the consensus and block information between each peer.

### Blockchain

The blockchain is loosely based on the TenderMint consensus algorithm to provide a simple and efficient agreement on top of permissioned nodes. The blockchain provides a simple storage agreement protocol for recording DID and VC's within IPFS as well as maintaining the active or revoked state of the records. 

Each consensus round only has 1 proposer. The round proposers are selected via an external randomness beacon provided by [drand](https://drand.love/).

Blocks contain transactions of an action to add, update or remove DIDs, VCs or nodes, all of which require a cryptographic signature, checked by a validator. 

### DID/VC & Messaging

The DID and VC data structures are directly taken from the w3c specifications for DIDComm, DID and VC's.

## Running

You may use DIDEm by running directly using `go` or run in a container runtime from the image `ghcr.io/tcfw/didem:main` (amd64 or arm64)

## Configuration 

TBC

## Contributing

TBC

## License

Please see [LICENSE](https://github.com/tcfw/didem/blob/main/LICENSE) file.