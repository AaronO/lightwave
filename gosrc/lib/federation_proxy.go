package lightwave

import (
  vec "container/vector"
  "http"
  "net"
  "log"
  "bytes"
  "io"
  "os"
  "fmt"
)

//--------------------------------------------------
// Helper class

type RequestBody struct {
  bytes.Buffer
}

func NewRequestBody(data []byte) *RequestBody {
  b := new(RequestBody)
  b.Write(data)
  return b
}

func (self *RequestBody) Close() os.Error {
  return nil
}

// ---------------------------------------------
// FederationRequest

type FederationRequest interface {
  SendFinishSignal(result bool)
  GetResponseWriter() http.ResponseWriter
}

// ---------------------------------------------
// FederationProxy

type FederationProxy struct {
  domain string
  manifest *ServerManifest
  queue vec.Vector
  connection *http.ClientConn
  idle bool
  reqChannel chan FederationRequest
  stopChannel chan bool
}

func NewFederationProxy(domain string) *FederationProxy {
  return &FederationProxy{domain:domain, idle:true, reqChannel:make(chan FederationRequest, 1000), stopChannel:make(chan bool)}
}

func (self *FederationProxy) Submit(req FederationRequest) {
  self.reqChannel <- req
}

func (self *FederationProxy) Run() {
  self.manifest = Discover(self.domain)
  log.Println("Discovery finished")
  for {
	select {
	  case req := <-self.reqChannel:
		self.enqueue(req)
	  case <-self.stopChannel:
		// TODO: cleanup eisting http connections
		return
	}
  }
}

func (self *FederationProxy) enqueue(req FederationRequest) {
  log.Println("Enqueueing a message for federation")
  self.queue.Push(req)
  if len(self.queue) == 1 && self.idle {
	self.sendAllFromQueue()
  }
}

func (self *FederationProxy) sendAllFromQueue() {
  for len(self.queue) > 0 {
	self.sendFromQueue()
	// TODO: If sending fails, this loop will try for ever to repeat the sending
  }
}

func (self *FederationProxy) sendFromQueue() {
  if len(self.queue) == 0 {
	return
  }
  
  log.Println("Sending message from queue")

  // No HTTP connection open yet?
  if self.connection == nil {
	con, err := net.Dial("tcp", "", fmt.Sprintf("%v:%v", self.manifest.HostName, self.manifest.Port))
	if err != nil {
	  // TODO: Good error handling
	  log.Println("Failed connecting to ", self.manifest, err)
	  return;
	}
	self.connection = http.NewClientConn(con, nil)
  }
  
  // Dequeue a message
  req := self.queue.At(0).(FederationRequest)
  // Build the HTTP request
  var hreq http.Request
  hreq.Host = self.manifest.Domain
  hreq.Header = make(map[string]string)
  switch req.(type) {
	case *PostRequest:
	  preq := req.(*PostRequest)
	  hreq.RawURL = fmt.Sprintf("http://%v:%v/fed/%v", self.manifest.HostName, self.manifest.Port, preq.URI.String())
	  hreq.Method = "POST"
	  hreq.Body = NewRequestBody(preq.Data)
	  hreq.ContentLength = int64(len(preq.Data))
	  hreq.Header["Content-Type"] = preq.MimeType
	case *GetRequest:
	  greq := req.(*GetRequest)
	  hreq.Method = "GET"  
	  hreq.RawURL = fmt.Sprintf("http://%v:%v/fed/%v", self.manifest.HostName, self.manifest.Port, greq.URI.String())
	default:
	  log.Println(req)
	  panic("Unsupported kind of message forwarded internally to the federation proxy")
  }
  log.Println("Sending request to ", hreq.RawURL)
  
  // Send the HTTP request
  self.connection.Write(&hreq)
  // Read the HTTP response
  hres, err := self.connection.Read()
  if err != nil {
	log.Println("Error reading HTTP response from ", self.manifest, err)
	// TODO: Better error handling
	self.connection.Close()
	self.connection = nil;
	return
  }
  
  // Success. Remove the request from the queue
  self.queue.Delete(0)
  
  // Send the result back to the client
  _, err = io.Copy(req.GetResponseWriter(), hres.Body)
  hres.Body.Close()
  if err != nil {
	log.Println("Error sending result of federated message back to the client")
	req.SendFinishSignal( false )
  }
  req.SendFinishSignal(true)
}
