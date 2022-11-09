package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"time"

	token "github.com/jskoven/mandatory_handin_4_dissys/grpc"
	"google.golang.org/grpc"
)

func main() {
	arg1, _ := strconv.ParseInt(os.Args[1], 10, 32)
	ownPort := int32(arg1) + 5000
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	p := &peer{
		id:            ownPort,
		index:         int32(arg1),
		amountOfPings: make(map[int32]int32),
		clients:       make(map[int32]token.PingClient),
		ctx:           ctx,
		token:         false,
	}
	if p.id == 5000 {
		p.token = true
		p.random = 1
	} else if p.id == 5001 {
		p.random = 2
	} else if p.id == 5002 {
		p.random = 3
	}

	// Create listener tcp on port ownPort
	list, err := net.Listen("tcp", fmt.Sprintf(":%v", ownPort))
	if err != nil {
		log.Fatalf("Failed to listen on port: %v", err)
	}
	grpcServer := grpc.NewServer()
	token.RegisterPingServer(grpcServer, p)

	go func() {
		if err := grpcServer.Serve(list); err != nil {
			log.Fatalf("failed to server %v", err)
		}
	}()

	for i := 0; i < 3; i++ {
		port := int32(5000) + int32(i)

		if port == ownPort {
			continue
		}

		var conn *grpc.ClientConn
		fmt.Printf("Trying to dial: %v\n", port)
		conn, err := grpc.Dial(fmt.Sprintf(":%v", port), grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			log.Fatalf("Could not connect: %s", err)
		}
		defer conn.Close()
		c := token.NewPingClient(conn)
		p.clients[port] = c
	}
	rand.Seed(int64(p.random))
	go p.hasToken()
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {

		/*Make this better lmao*/
		indextouse := int(p.id) + 1
		if indextouse == 5003 {
			indextouse = 5000
		}
		p.sendTokenToNext(indextouse)
	}
}

type peer struct {
	token.UnimplementedPingServer
	id            int32
	index         int32
	amountOfPings map[int32]int32
	clients       map[int32]token.PingClient
	ctx           context.Context
	token         bool
	random        int32
}

func (p *peer) Ping(ctx context.Context, req *token.Request) (*token.Reply, error) {
	id := req.Id
	p.amountOfPings[id] += 1

	rep := &token.Reply{Amount: p.amountOfPings[id]}
	return rep, nil
}

func (p *peer) sendPingToAll() {
	request := &token.Request{Id: p.id}

	for id, client := range p.clients {
		reply, err := client.Ping(p.ctx, request)
		emptybox := token.Empty{}
		reply2, err2 := client.GiveToken(p.ctx, &emptybox)
		if err != nil {
			fmt.Println("something went wrong")
		}
		if err2 != nil {
			fmt.Println("token not sent correctly")
		}
		_ = reply2
		fmt.Printf("Got reply from id %v: %v\n", id, reply.Amount)
	}
}

func (p *peer) sendTokenToNext(index int) {
	if p.token {
		emptybox := token.Empty{}
		p.clients[int32(index)].GiveToken(p.ctx, &emptybox)
		p.token = false
	}

}

func (p *peer) hasToken() {
	for {
		if p.token {
			fmt.Printf("%d has the token bro", p.id)
			fmt.Println()
			p.random = rand.Int31n(6)
			fmt.Printf("%d is the random number", p.random)
			fmt.Println()
			if p.random == 1 {
				time.Sleep(3 * time.Second)
				fmt.Printf("%d did their thing", p.id)
				fmt.Println()
			}
			indextouse := int(p.id) + 1
			if indextouse == 5003 {
				indextouse = 5000
			}
			time.Sleep(500 * time.Millisecond)
			p.sendTokenToNext(indextouse)
		}

	}
}

func (p *peer) GiveToken(context.Context, *token.Empty) (*token.Empty, error) {
	p.token = true
	emptybox := &token.Empty{}
	return emptybox, nil
}
