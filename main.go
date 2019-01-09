package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"reflect"
	"sort"
	"time"

	libp2p "github.com/libp2p/go-libp2p"
	crypto "github.com/libp2p/go-libp2p-crypto"
	ps "github.com/libp2p/go-libp2p-peerstore"
	ma "github.com/multiformats/go-multiaddr"
)

// helper method - create a lib-p2p host to listen on a port
func makeRandomNode(port int, done chan bool, profile []string, orbit []Body) *Node {
	// Ignoring most errors for brevity
	// See echo example for more details and better implementation
	priv, _, _ := crypto.GenerateKeyPair(crypto.Secp256k1, 256)
	listen, _ := ma.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", port))
	host, _ := libp2p.New(
		context.Background(),
		libp2p.ListenAddrs(listen),
		libp2p.Identity(priv),
		libp2p.DisableRelay(),
	)

	return NewNode(host, done, profile, []Body{})
}

type TestData struct {
	TestNetwork map[string][]string
	TestOrbit   []string
}

func testGravitation(fname string) bool {
	// Read test data from file
	testConfig := TestData{}
	b, err := ioutil.ReadFile(fname)
	if err != nil {
		log.Println("Error reading data from file. ")
	}
	err = json.Unmarshal(b, &testConfig)

	if err != nil {
		log.Println("Error loading data. ")
	}

	done := make(chan bool, 1)

	// Set up test network
	hostMap := make(map[string]*Node)
	orbitPeerIds := []string{}
	for k, v := range testConfig.TestNetwork {
		rand.Seed(666)
		port := rand.Intn(100) + 10000

		if _, exist := hostMap[k]; !exist {
			profile := time.Now().Format("20060102150405")
			hostMap[k] = makeRandomNode(port, done, []string{profile}, []Body{})
		}
		for _, peer := range v {

			if _, exist := hostMap[peer]; !exist {
				rand.Seed(666)
				newPort := port + 1

				testProfile := []string{"test"}
				inOrbit := false

				// if part of orbit
				for _, hostName := range testConfig.TestOrbit {
					if hostName == peer {
						inOrbit = true
						testProfile = hostMap[k].gravData.Profile
					}
				}

				// Creates host and creates peer relationship between it and the root peer
				hostMap[peer] = makeRandomNode(newPort, done, testProfile, []Body{})
				if inOrbit {
					orbitPeerIds = append(orbitPeerIds, hostMap[peer].ID().String())
				}
				hostMap[k].Peerstore().AddAddrs(hostMap[peer].ID(), hostMap[peer].Addrs(), ps.PermanentAddrTTL)
				hostMap[peer].Peerstore().AddAddrs(hostMap[k].ID(), hostMap[k].Addrs(), ps.PermanentAddrTTL)
				log.Printf("This is a conversation between %s and %s\n", hostMap[k].ID(), hostMap[peer].ID())

				// Perform gravitation
				hostMap[k].Gravitation(hostMap[peer].Host)

			}
		}
	}

	// Waits for process to finish
	time.Sleep(2 * time.Second)

	actualOrbitIds := []string{}
	for _, data := range hostMap["A"].gravData.Orbit {
		actualOrbitIds = append(actualOrbitIds, data.peerID)
	}

	sort.Strings(actualOrbitIds)
	sort.Strings(orbitPeerIds)

	for i := 0; i < 4; i++ {
		<-done
	}

	return reflect.DeepEqual(actualOrbitIds, orbitPeerIds)
}

const help = `
Creates Gravitation protocol instance. 
Usage: 
./GravitationProtocol 
  - runs default gravitation test 
./GravitationProtocol -t testfile 
  - runs a gravitation protocol test with given test file 
`

func main() {
	flag.Usage = func() {
		fmt.Println(help)
		flag.PrintDefaults()
	}

	// Parse some flags
	testFile := flag.String("t", "", "Test File")
	flag.Parse()

	if *testFile != "" {
		if testGravitation(*testFile) {
			log.Println("Test successful!")
		} else {
			log.Println("Test failed.")
		}
	} else {
		testGravitation("test.json")
	}
}
