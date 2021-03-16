package node_tcp

import (
	"net"
	"pandora-pay/blockchain"
	"pandora-pay/config/globals"
	"pandora-pay/gui"
	"pandora-pay/mempool"
	"pandora-pay/node/api"
	node_http "pandora-pay/node/node-http"
	"pandora-pay/settings"
)

// ControllerAddr is the Tor controller interface address
// Note:
type TcpServer struct {
	Address string
	Port    string

	settings *settings.Settings
	chain    *blockchain.Blockchain

	tcpListener net.Listener
	api         *api.API
	HttpServer  *node_http.HttpServer
}

func (server *TcpServer) initialize() {

	// Create local listener on next available port

	port := "8080"
	if globals.Arguments["--tcp-server-port"] != nil {
		port = globals.Arguments["--tcp-server-port"].(string)
	}

	var address string
	if globals.Arguments["--tor-onion"] != nil {
		address = globals.Arguments["--tor-onion"].(string)
	}
	if globals.Arguments["--tcp-server-address"] != nil {
		address = globals.Arguments["--tcp-server-address"].(string)
	}

	if address == "" {
		conn, err := net.Dial("udp", "8.8.8.8:80")
		if err != nil {
			gui.Error("Error dialing dns to discover my own ip")
			panic(err)
		}
		address = conn.LocalAddr().(*net.UDPAddr).IP.String()
		if err = conn.Close(); err != nil {
			gui.Error("Error closing the connection")
		}
	}
	server.Address = address
	server.Port = port

	var err error
	server.tcpListener, err = net.Listen("tcp", "127.0.0.1:"+port)
	if err != nil {
		gui.Error("Error creating TcpServer")
		panic(err)
	}

	gui.InfoUpdate("TCP", address+":"+port)

	server.HttpServer = node_http.CreateHttpServer(server.tcpListener, server.chain, server.api)
}

func CreateTcpServer(settings *settings.Settings, chain *blockchain.Blockchain, mempool *mempool.Mempool) *TcpServer {

	server := &TcpServer{
		settings: settings,
		chain:    chain,
		api:      api.CreateAPI(chain, mempool),
	}
	server.initialize()

	return server
}
