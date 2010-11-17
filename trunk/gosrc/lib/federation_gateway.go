package lightwave

import (
  vec "container/vector"
  "http"
  "net"
  "log"
  "fmt"
)

var capabilitiesStore *CapabilitiesStore = NewCapabilitiesStore()

type WaveletUpdateRequest struct {
  uri string
  Data []byte
  MimeType string
}

type FederationGateway struct {
  // The domain name of the remote server. This is required while we do not have a manifest
  domain string
  // Reference to the local server instance
  server *Server
  // Manuscript of the remote server instance (not-nil after discovery)
  manifest *ServerManifest
  // Queued messages
  queue vec.Vector
  // The HTTP connection used to transmit messages
  connection *http.ClientConn
  // Currently trying to send a message to the remote server? If not -> idle == true
  idle bool
  reqChannel chan *WaveletUpdateRequest
  // If a message has been transfered to the remote server, this signal is triggered to send the next one
  nextChannel chan bool
  stopChannel chan bool
}

func NewFederationGateway(server *Server, domain string) *FederationGateway {
  return &FederationGateway{domain:domain, server:server, idle:true, reqChannel:make(chan *WaveletUpdateRequest, 1000), stopChannel:make(chan bool), nextChannel:make(chan bool)}
}

func (self *FederationGateway) WaveletUpdate(uri string, data []byte) {
  req := &WaveletUpdateRequest{uri, data, "application/json"}
  self.reqChannel <- req
}

func (self *FederationGateway) Run() {
  self.manifest = Discover(self.domain)
  if self.manifest == nil {
	log.Println("Discovery FAILED")
	return
  }
  log.Println("Discovery of remote server finished")

  // Enter the main event loop
  for {
	select {
	  case result := <-self.nextChannel:
		self.idle = true
		if !result {
		  log.Println("Failed talking to remote server ", self.domain)
		  // TODO: Better error management
		} else {
		  // The first request in the list has been successfully transmitted
		  self.queue.Delete(0)
		  // If there are more requests, try with the next one
		  if len(self.queue) > 0 {
			go self.sendFromQueue(self.queue[0].(*WaveletUpdateRequest))
		  }
		}
	  case req := <-self.reqChannel:
		// Enqueue the request
		self.queue.Push(req)
		// Can we send it right now?
		if self.idle {
		  self.idle = false
		  go self.sendFromQueue(self.queue[0].(*WaveletUpdateRequest))
		}
	  case <-self.stopChannel:
		// TODO: cleanup existing http connections
		return
	}
  }
}

// This function is launched async as a go routine. It tries to send a message
// to a remote server and sends a bool to nextChannel to indicate whether this
// has succeeded or not. It is not allowed to run this function multiple times in parallel
// for the same FederationGateway.
func (self *FederationGateway) sendFromQueue(req *WaveletUpdateRequest) {  
  // No HTTP connection open yet?
  if self.connection == nil {
	con, err := net.Dial("tcp", "", fmt.Sprintf("%v:%v", self.manifest.HostName, self.manifest.Port))
	if err != nil {
	  // TODO: Good error handling
	  log.Println("Failed connecting to ", self.manifest, err)
	  self.nextChannel <- false
	  return
	}
	self.connection = http.NewClientConn(con, nil)
  }
  
  // Build the HTTP request
  var hreq http.Request
  hreq.Host = self.manifest.Domain
  hreq.Header = make(map[string]string)
  hreq.RawURL = fmt.Sprintf("http://%v:%v/fed%v", self.manifest.HostName, self.manifest.Port, req.uri)
  hreq.Method = "PUT"
  hreq.Body = NewRequestBody(req.Data)
  hreq.ContentLength = int64(len(req.Data))
  hreq.Header["Content-Type"] = req.MimeType
  log.Println("Sending WaveletUpdate to remote server ", hreq.RawURL)
  
  // Send the HTTP request
  self.connection.Write(&hreq)
  // Read the HTTP response
  hres, err := self.connection.Read()
  if err != nil {
	log.Println("Error reading HTTP response from ", self.manifest, err)
	// TODO: Better error handling
	self.connection.Close()
	self.connection = nil
	self.nextChannel <- false
	return
  }

  log.Println("After sending WaveletUpdate, status code is ", hres.StatusCode)
  // Success in sending the wavelet update?
  if hres.StatusCode == 200 {
	self.nextChannel <- true
	return
  }
  // Sending the wavelet update failed
  self.nextChannel <- false
}
