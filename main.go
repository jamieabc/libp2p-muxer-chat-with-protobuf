package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"fmt"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/multiformats/go-multiaddr"
	mrand "math/rand"
	"os"
	"os/signal"
)

const (
	chatProtocol = "/mychat/1.0.0"
)

func main() {
	if 1 == len(os.Args) {
		h := serverMode()
		h.SetStreamHandler(protocol.ID(chatProtocol), serverStreamHandler)
	} else {
		s := clientMode()
		rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
		go readFromStdinAndWrite(rw)
		go readFromStream(rw)
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

	go readFromStream(rw)
	go readFromStdinAndWrite(rw)
}

func readFromStdinAndWrite(rw *bufio.ReadWriter) {
	stdin := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("> ")
		msg, err := stdin.ReadString('\n')
		if nil != err {
			panic(err)
		}

		_, _ = rw.WriteString(msg)
		_ = rw.Flush()
	}
}

func readFromStream(rw *bufio.ReadWriter) {
	for {
		msg, _ := rw.ReadString('\n')
		if "" == msg {
			return
		}
		if "\n" != msg {
			fmt.Println("receive message: ", msg)
		}
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
