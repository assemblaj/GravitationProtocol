package main

import (
	"bufio"
	"context"
	"fmt"
	"log"

	uuid "github.com/google/uuid"
	p2p "github.com/libp2p/go-libp2p-examples/multipro/pb"
	host "github.com/libp2p/go-libp2p-host"
	inet "github.com/libp2p/go-libp2p-net"
	protobufCodec "github.com/multiformats/go-multicodec/protobuf"
)

// pattern: /protocol-name/request-or-response-message/version
const gravitationRequest = "/gravitation/gravitationreq/0.0.1"
const gravitationResponse = "/gravitation/gravitationresp/0.0.1"

// GravitationProtocol type
type GravitationProtocol struct {
	node     *Node                              // local host
	requests map[string]*p2p.GravitationRequest // used to access request data from response handlers
	done     chan bool                          // only for demo purposes to stop main from terminating
}

func NewGravitationProtocol(node *Node, done chan bool) *GravitationProtocol {
	p := &GravitationProtocol{node: node, requests: make(map[string]*p2p.GravitationRequest), done: done}
	node.SetStreamHandler(gravitationRequest, p.onGravitationRequest)
	node.SetStreamHandler(gravitationResponse, p.onGravitationResponse)
	return p
}

// remote peer requests handler
func (p *GravitationProtocol) onGravitationRequest(s inet.Stream) {

	// get request data
	data := &p2p.GravitationRequest{}
	decoder := protobufCodec.Multicodec(nil).Decoder(bufio.NewReader(s))
	err := decoder.Decode(data)
	if err != nil {
		log.Println(err)
		return
	}

	log.Printf("%s: Received gravitation request from %s. Message: %s", s.Conn().LocalPeer(), s.Conn().RemotePeer(), data.Message)

	valid := p.node.authenticateMessage(data, data.MessageData)

	if !valid {
		log.Println("Failed to authenticate message")
		return
	}

	// generate response message
	log.Printf("%s: Sending gravitation response to %s. Message id: %s...", s.Conn().LocalPeer(), s.Conn().RemotePeer(), data.MessageData.Id)

	resp := &p2p.GravitationResponse{MessageData: p.node.NewMessageData(data.MessageData.Id, false),
		Message: fmt.Sprintf("Gravitation response from %s", p.node.ID())}

	// sign the data
	signature, err := p.node.signProtoMessage(resp)
	if err != nil {
		log.Println("failed to sign response")
		return
	}

	// add the signature to the message
	resp.MessageData.Sign = signature

	// send the response
	s, respErr := p.node.NewStream(context.Background(), s.Conn().RemotePeer(), gravitationResponse)
	if respErr != nil {
		log.Println(respErr)
		return
	}

	ok := p.node.sendProtoMessage(resp, s)

	if ok {
		log.Printf("%s: Gravitation response to %s sent.", s.Conn().LocalPeer().String(), s.Conn().RemotePeer().String())
	}
}

// remote gravitation response handler
func (p *GravitationProtocol) onGravitationResponse(s inet.Stream) {
	data := &p2p.GravitationResponse{}
	decoder := protobufCodec.Multicodec(nil).Decoder(bufio.NewReader(s))
	err := decoder.Decode(data)
	if err != nil {
		return
	}

	valid := p.node.authenticateMessage(data, data.MessageData)

	if !valid {
		log.Println("Failed to authenticate message")
		return
	}

	// locate request data and remove it if found
	_, ok := p.requests[data.MessageData.Id]
	if ok {
		// remove request from map as we have processed it here
		delete(p.requests, data.MessageData.Id)
	} else {
		log.Println("Failed to locate request data boject for response")
		return
	}

	log.Printf("%s: Received gravitation response from %s. Message id:%s. Message: %s.", s.Conn().LocalPeer(), s.Conn().RemotePeer(), data.MessageData.Id, data.Message)
	p.done <- true
}

func (p *GravitationProtocol) Gravitation(host host.Host) bool {
	log.Printf("%s: Sending gravitation to: %s....", p.node.ID(), host.ID())

	// create message data
	req := &p2p.GravitationRequest{MessageData: p.node.NewMessageData(uuid.New().String(), false),
		Message: fmt.Sprintf("Gravitation from %s", p.node.ID())}

	// sign the data
	signature, err := p.node.signProtoMessage(req)
	if err != nil {
		log.Println("failed to sign pb data")
		return false
	}

	// add the signature to the message
	req.MessageData.Sign = signature

	s, err := p.node.NewStream(context.Background(), host.ID(), gravitationRequest)
	if err != nil {
		log.Println(err)
		return false
	}

	ok := p.node.sendProtoMessage(req, s)

	if !ok {
		return false
	}

	// store ref request so response handler has access to it
	p.requests[req.MessageData.Id] = req
	log.Printf("%s: Gravitation to: %s was sent. Message Id: %s, Message: %s", p.node.ID(), host.ID(), req.MessageData.Id, req.Message)
	return true
}
