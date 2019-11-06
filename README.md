This repository uses libp2p for a local p2p network.

Usage

1. run program first time without any parameter, log shows host connection information as
`server multi address: /ip4/127.0.0.1/tcp/53126/p2p/QmVzJW5owZC9a4Miq6vRPtjuoML3anbnymjebwjEQxReXB`

  copy the whole string from /ip4.... to ReXB, noted everytime program runs, this string will be different, so it needs to use correct information listed on screen.
  
2. run program second time with parameter copied, e.g. `libp2p-muxer-chat-with-protobuf /ip4/127.0.0.1/tcp/53126/p2p/QmVzJW5owZC9a4Miq6vRPtjuoML3anbnymjebwjEQxReXB`

3. start chatting