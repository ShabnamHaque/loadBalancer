package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

type SimpleServer struct {
	addr  string
	proxy *httputil.ReverseProxy
}

func handleErr(err error) {
	if err != nil {
		fmt.Printf("err: %v", err)
		os.Exit(1)
	}
}
func newSimpleServer(addr string) *SimpleServer {
	serverUrl, err := url.Parse(addr)
	handleErr(err)

	return &SimpleServer{
		addr:  addr,
		proxy: httputil.NewSingleHostReverseProxy(serverUrl),
	}
}

type LoadBalancer struct {
	port            string
	roundRobinCount int
	servers         []Server
}

type Server interface {
	Address() string
	IsAlive() bool
	Serve(rw http.ResponseWriter, r *http.Request)
}

func NewLoadBalancer(port string, servers []Server) *LoadBalancer {
	return &LoadBalancer{
		roundRobinCount: 0,
		port:            port,
		servers:         servers,
	}
}

func (lb *LoadBalancer) getNextAvailableServer() Server {
	server := lb.servers[lb.roundRobinCount%len(lb.servers)]
	//which server will it be directed to.

	for !server.IsAlive() {  //if the server is not alive,we move to the next server until we find a one that is alive and return set the server
		lb.roundRobinCount++
		server = lb.servers[lb.roundRobinCount%len(lb.servers)] //move to the next server.
	}
	lb.roundRobinCount++
	return server
}
func (lb *LoadBalancer) ServeProxy(rw http.ResponseWriter, req *http.Request) {
	targetServer := lb.getNextAvailableServer() //find the apporpriate server to direct to.
	fmt.Printf("Forwarding req to address %q\n", targetServer.Address())
	targetServer.Serve(rw, req) //serve the request to the targetServer
}
func (s *SimpleServer) Address() string {
	return s.addr
}
func (s *SimpleServer) IsAlive() bool {
	return true
}
func (s *SimpleServer) Serve(rw http.ResponseWriter, req *http.Request) {
	s.proxy.ServeHTTP(rw, req)
}
func main() {

	servers := []Server{
		newSimpleServer("https://www.instagram.com"),
		newSimpleServer("https://www.bing.com"),
		newSimpleServer("https://www.facebook.com"),
	}
	lb := NewLoadBalancer("8080", servers)
	handleRedirect := func(rw http.ResponseWriter, req *http.Request) {
		lb.ServeProxy(rw, req)
	}

	http.HandleFunc("/", handleRedirect)
	fmt.Printf("serving requests at 'localhost:%s'\n", lb.port)
	http.ListenAndServe(":"+lb.port, nil)
}
