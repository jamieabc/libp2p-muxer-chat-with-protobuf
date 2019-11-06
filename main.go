package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"fmt"
	mrand "math/rand"
	"os"
	"os/signal"

	"github.com/golang/protobuf/proto"
	"github.com/jamieabc/libp2p-muxer-chat-with-protobuf/pb"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/multiformats/go-multiaddr"
)

const (
	chatProtocol = "/mychat/1.0.0"
	maxBytes     = 1000
)

func main() {
	if 1 == len(os.Args) {
		h := serverMode()
		h.SetStreamHandler(protocol.ID(chatProtocol), serverStreamHandler)
	} else {
		s := clientMode()
		rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
		go readFromStdinAndWrite(rw, s, "client")
		go readFromStream(rw, s)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}

func serverMode() host.Host {
	prvKey := generateEncryption()
	port := mrand.Intn(65535)
	maddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", port))
	if nil != err {
		panic(err)
	}
	h, err := libp2p.New(
		context.Background(),
		libp2p.ListenAddrs(maddr),
		libp2p.Identity(prvKey),
	)

	if nil != err {
		panic(err)
	}

	fmt.Printf("server multi address: /ip4/127.0.0.1/tcp/%d/p2p/%s\n", port, h.ID().Pretty())
	return h
}

func serverStreamHandler(s network.Stream) {
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	go readFromStream(rw, s)
	go readFromStdinAndWrite(rw, s, "server")
}

func readFromStdinAndWrite(rw *bufio.ReadWriter, s network.Stream, source string) {
	defer s.Close()

	stdin := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("> ")
		msg, err := stdin.ReadString('\n')
		if nil != err {
			fmt.Printf("read stream with error: %s\n", err)
			return
		}

		data := &pb.Message{
			Source: source,
			Msg:    msg,
		}
		bin, err := proto.Marshal(data)
		if nil != err {
			panic(err)
		}

		_, _ = rw.Write(bin)
		_ = rw.Flush()
	}
}

func readFromStream(rw *bufio.ReadWriter, s network.Stream) {
	defer s.Close()

	data := make([]byte, maxBytes)
	for {
		length, err := rw.Read(data)
		if nil != err {
			fmt.Printf("read stream with error: %s\n", err)
			return
		}
		if 0 == length {
			continue
		}
		out := pb.Message{}
		err = proto.Unmarshal(data[:length], &out)
		if nil != err {
			panic(err)
		}
		fmt.Printf("from %s: %s", out.Source, out.Msg)
	}
}

func generateEncryption() crypto.PrivKey {
	r := rand.Reader
	privateKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if nil != err {
		panic(err)
	}
	return privateKey
}

func clientMode() network.Stream {
	prvKey := generateEncryption()
	maddr, err := multiaddr.NewMultiaddr(os.Args[1])
	if nil != err {
		panic(err)
	}

	info, err := peer.AddrInfoFromP2pAddr(maddr)
	if nil != err {
		panic(err)
	}

	h, err := libp2p.New(
		context.Background(),
		libp2p.Identity(prvKey),
	)

	if nil == h {
		panic(fmt.Errorf("libp2p new empty host"))
	}

	h.Peerstore().AddAddrs(info.ID, info.Addrs, peerstore.PermanentAddrTTL)
	s, err := h.NewStream(context.Background(), info.ID, protocol.ID(chatProtocol))
	if nil != err {
		panic(err)
	}

	return s
}
