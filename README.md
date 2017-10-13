# lseq

This intends to become a backend for a distributed, peer-to-peer,
collaborative/concurrent text editor (or, possibly, wiki).


## Progress

A document model is mostly implemented as `lseq.Document`.


## Building blocks / proposal

### Document model

We base this on a conflict-free replicated data structure ([CRDT][3]), to allow
for concurrent, asynchronous edits.
Academic papers include [LSEQ][1] and [Logoot-Undo][2]

[1]: https://hal.archives-ouvertes.fr/hal-00921633/
[2]: https://hal.archives-ouvertes.fr/hal-00450416/
[3]: https://en.wikipedia.org/wiki/Conflict-free_replicated_data_type


### Edit protocol

The CRDTs largely dictates the abstract protocols (patches as lists of
insertions and deletions).

To complement this, we need to model at the protocol level:

- edit lists
- per-peer cursors and catch-ups
- bootstrap / full-document transfer

And to implement at least:

- patch/edit list serialization
- document serialisation (for catchup)

We propose to use gRPC for peer-to-peer communications.


### NAT-proof transport

We want the system to operate peer-to-peer, which mostly precludes using TCP.
Unfortunately [grpc-go][5] only documents TCP usage.

[kcp-go][4] seems promising; it exposes a `net.TCPConn`-compatible API, for
reliable semantics over UDP.

[4]: https://github.com/xtaci/kcp-go
[5]: https://github.com/grpc/grpc-go

We will need to use one of the many publicly-available STUN servers during ICE.


### Peer discovery mechanism

The Kademlia DHT is tried and tested; we can use it to locate a list of peers for
a given document identifier.

A couple Golang implementations exist,
[shiyanhui/dht](https://github.com/shiyanhui/dht) and
[nictuku/dht](https://github.com/nictuku/dht).

This may require running at least one bootstrap node.


### Peer sampling protocol

When large swarms edit the same document, gossiping edits to all is
unreasonable.

A peer sampling protocol like [Spray][6] will keep gossiping and connection
counts reasonable.

[6]: https://hal.inria.fr/hal-01203363


### User interface

Browsers do not support true peer-to-peer as WebRTC does not support serverless
connection establishment. Golang does not have good WebRTC bindings yet, either.

For prototyping purposes, we propose using an Electron-style approach with
[murlokswarm/app](https://github.com/murlokswarm/app).


