package wave

import (
  . "lightwave"
  "strings"
  "log"
  "http"
  "strconv"
  "json"
  "bytes"
  proto "goprotobuf.googlecode.com/hg/proto"  
)

// --------------------------------------------------------
// Helper functions

// Creates a HTTP error response
func makeErrorResponse(res http.ResponseWriter, errorText string) {
  log.Println(errorText)
  m := make(map[string]interface{})
  m["ok"] = false
  m["error"] = errorText
  json, err := json.Marshal(m)
  if err != nil {
	panic("Failed marshaling to json")
  }
  res.SetHeader("Content-Type", "application/json")
  _, err = res.Write( json )
  if err != nil {
	log.Println("Failed writing HTTP response")
  }
}

func makeClientErrorResponse(res http.ResponseWriter, errorText string) {
  log.Println(errorText)
  response := &ClientSubmitResponse{}
  response.OperationsApplied = proto.Int( 0 )
  response.ErrorMessage = proto.String(errorText)
  var buffer bytes.Buffer
  err := Marshal(response, &buffer)
  if err != nil {
	panic("Failed marshaling")
  }
  res.SetHeader("Content-Type", "application/json")
  _, err = res.Write( buffer.Bytes() )
  if err != nil {
	log.Println("Failed writing HTTP response")
  }
}


// -------------------------------------------
// WaveletNode

type WaveletNode struct {
  parent Node
  name string
  postChannel chan *PostRequest
  getChannel chan *GetRequest
  stopChannel chan bool
  pubSubChannel chan *PubSubRequest
  wavelet *Wavelet
  level int
  subscriptions map[string]Subscription
  // List of domains which participate in federating this document
  federatedDomains map[string]bool
  // TODO history *DocumentHistory
}

func NewWaveletNode(parent Node, name string, level int) Node {
  namesplit := strings.Split(name, "$", -1)
  if len(namesplit) != 3 {
	log.Println("Invalid name for a wavelet node ", name)
	return nil
  }
  d := &WaveletNode{level:level, parent:parent, name:name,postChannel:make(chan *PostRequest), getChannel:make(chan *GetRequest), stopChannel:make(chan bool), pubSubChannel:make(chan *PubSubRequest)}
  d.subscriptions = make(map[string]Subscription)  
  wurl := &WaveUrl{WaveDomain:namesplit[0],WaveId:namesplit[1], WaveletDomain:d.Host().Name(), WaveletId:namesplit[2]}
  d.wavelet = NewWavelet(wurl)
  d.federatedDomains = make(map[string]bool)
  // TODO d.history = NewDocumentHistory()
  return d
}

func (self *WaveletNode) Server() *Server {
  if self.parent == nil {
	return nil
  }
  if s, ok := self.parent.(*Server); ok {
	return s
  }
  return self.parent.Server()
}

func (self *WaveletNode) Host() *HostNode {
  if self.parent == nil {
	return nil
  }
  if s, ok := self.parent.(*HostNode); ok {
	return s
  }
  return self.parent.Host()
}

func (self *WaveletNode) Post(req *PostRequest) {
  self.postChannel <- req
}

func (self *WaveletNode) Get(req *GetRequest) {
  self.getChannel <- req
}

func (self *WaveletNode) Stop() {
  self.stopChannel <- true
}

func (self *WaveletNode) PubSub(req *PubSubRequest) {
    self.pubSubChannel <- req
}

func (self *WaveletNode) Name() string {
  return self.name
}

func (self *WaveletNode) URI() string {
  if self.parent != nil {
	return self.parent.URI() + "/" + self.name	
  }
  return ""
}

func (self *WaveletNode) Run() {
  for {
	select {
	  case req := <-self.postChannel:
		self.post(req)
	  case req := <-self.getChannel:
		self.get(req)
	  case req := <-self.pubSubChannel:
		self.pubSub(req)		
	  case <-self.stopChannel:
		return
	}
  }
}

func (self *WaveletNode) apply( delta *ProtocolWaveletDelta ) bool {
  if self.wavelet.HashedVersion.Equals( delta.HashedVersion ) {
	err := self.wavelet.ApplyDelta(delta)
	if err != nil {
	  log.Println("Failed applying delta: ", err)
	  return false
	}
	*self.wavelet.HashedVersion.Version += int64(len(delta.Operation))
	// TODO: Calculate new hash

	// Send message to subscribers
	if len(self.subscriptions) > 0 {
	  update := &ClientWaveletUpdate{}
	  update.WaveletName = proto.String(self.wavelet.Url.String())
	  update.Deltas = [...]*ProtocolWaveletDelta{delta}[:]
	  // Send JSON encoding 
	  var buffer bytes.Buffer
	  Marshal( update, &buffer )
	  for _, s := range self.subscriptions {
		s.Subscriber.Update( &UpdateMsg{self.URI(), buffer.String()})
	  }
	}
  } else {
	log.Println(string(self.wavelet.HashedVersion.HistoryHash))
	log.Println(string(delta.HashedVersion.HistoryHash))
	panic("Not implemented yet")
  }
  // TODO history self.history.Append(m)
  // At this point we know that the mutation is applied.

  // If this is just a remote server for this document, we are done
  if !self.Host().IsLocal() {
	return true
  }
  
  server := self.Server()
  // Find out whether there are new domains involved. In this case we
  // have to send the mutation as a wavelet update to all the others
  // This code assumes that the meta data can be arbitrarily malformed. This is perhaps overly defensive
  self.federatedDomains = make(map[string]bool)
  for _, s := range self.wavelet.Participants {
	u := NewUserId(s)
	if u != nil {
	  // If this is a remote user, we must federate
	  if u.Domain != server.Capabilities().Domain {
		self.federatedDomains[u.Domain] = true
	  }
	}
  }
  
  // Forward the mutation to all federated servers
  if len(self.federatedDomains) > 0 {
	/* TODO: Encode the delta and send it to the federated servers
	msg, err := json.Marshal(mutation)
	if err != nil {
	  panic("Cannot encode my own data")
	}
	for domain, _ := range self.federatedDomains {
	  self.Server().Gateway(domain).WaveletUpdate(self.URI(), msg)
	}
	*/
  }
  
  return true
}

func (self *WaveletNode) Revision() int64 {
  return *self.wavelet.HashedVersion.Version
}
 
func (self *WaveletNode) IsLocal() bool {
  return self.Host().IsLocal()
}

func (self *WaveletNode) post(req *PostRequest) {  
  docuri := req.URI.(DocumentURI)
  
  // Request is not aimed at this document?
  if len(docuri.NameSeq) != self.level {
	makeErrorResponse( req.Response, "Document " + docuri.NameSeq[self.level] + " does not exist" )
	req.FinishSignal <- false
	return	
  }
  log.Println("Wavelet is handling a post: ", req.URI)
  
  switch req.Origin {
	// A message from the client. It must be a ProtocolWaveletDelta
	case ClientOrigin:
	  if req.MimeType != "application/json-wave" {
		makeClientErrorResponse(req.Response, "Data type " + req.MimeType + " not allowed for put/post")
		req.FinishSignal <- false
	  }
	  submit := &ClientSubmitRequest{}
	  if err := Unmarshal(req.Data, submit); err != nil {
		makeClientErrorResponse(req.Response, "Cannot parse HTTP body. No valid ProtoBuf or wrong message type: " + err.String())
		req.FinishSignal <- false
		return
	  }
	  if !self.apply(submit.Delta) {
		makeClientErrorResponse(req.Response, "Could not apply document mutation")
		req.FinishSignal <- false
		return
	  }
	  response := &ClientSubmitResponse{}
	  response.OperationsApplied = proto.Int( len(submit.Delta.Operation) )
	  // TODO: time stamp
	  response.HashedVersionAfterApplication = &self.wavelet.HashedVersion
	  var buffer bytes.Buffer
	  Marshal(response, &buffer)
	  req.Response.SetHeader("Content-Type", "application/json")
	  if _, err := req.Response.Write( buffer.Bytes() ); err != nil {
		log.Println("Failed writing HTTP response")
		req.FinishSignal <- false
		return
	  }
	  req.FinishSignal <- true
	// A message via federation
	case FederationOrigin:
	  if self.IsLocal() {
		// This server is the hosting server?
		// It must be a ProtocolSignedDelta
		// TODO
	  } else {
		// This server is the remote server?
		// It must be a ProtocolAppliedWaveletDelta
		// TODO
	  }
  }
}

func (self *WaveletNode) get(req *GetRequest) {
  docuri := req.URI.(DocumentURI)

  // Request is not aimed at this document?
  if len(docuri.NameSeq) != self.level {
	makeErrorResponse( req.Response, "Document " + docuri.NameSeq[self.level] + " does not exist" )
	req.FinishSignal <- false
	return	
  }
  log.Println("Wavelet is handling a get: ", req.URI)

  // Is a special history requested?
  if req.RawQuery != "" {
	// Parse the query
	query, err := http.ParseQuery(req.RawQuery)
	if err != nil {
	  log.Println("Failed parsing query")
	  makeErrorResponse(req.Response, "Failed parsing query")
	  req.FinishSignal <- false
	  return
	}
	_v1, ok := query["v1"]
	if !ok || len(_v1) != 1 {
	  log.Println("Expected v1 in query string")
	  makeErrorResponse(req.Response, "Expected v1 in query string")
	  req.FinishSignal <- false
	  return
	}
	v1, err := strconv.Atoi(_v1[0])
	if err != nil {
	  log.Println("Malformed query")
	  makeErrorResponse(req.Response, "Expected v1 in query string")
	  req.FinishSignal <- false
	  return
	}
	v1hash, ok := query["v1hash"]
	if !ok || len(v1hash) != 1 {
	  log.Println("Expected v1hash in query string")
	  makeErrorResponse(req.Response, "Expected v1hash in query string")
	  req.FinishSignal <- false
	  return
	}
	_v2, ok := query["v2"]
	if !ok || len(_v2) != 1 {
	  log.Println("Expected v2 in query string")
	  makeErrorResponse(req.Response, "Expected v2 in query string")
	  req.FinishSignal <- false
	  return
	}
	v2, err := strconv.Atoi(_v2[0])
	if err != nil {
	  log.Println("Malformed query")
	  makeErrorResponse(req.Response, "Expected v1 in query string")
	  req.FinishSignal <- false
	  return
	}
	v2hash, ok := query["v1hash"]
	if !ok || len(v2hash) != 1 {
	  log.Println("Expected v2hash in query string")
	  makeErrorResponse(req.Response, "Expected v2hash in query string")
	  req.FinishSignal <- false
	  return
	}
	_limit, ok := query["limit"]
	if ok && len(_limit) != 1 {
	  log.Println("Double limit in query string")
	  makeErrorResponse(req.Response, "Double limit in query string")
	  req.FinishSignal <- false
	  return
	}
	limit, err := strconv.Atoi(_limit[0])
	if err != nil {
	  log.Println("Malformed query")
	  makeErrorResponse(req.Response, "Expected v1 in query string")
	  req.FinishSignal <- false
	  return
	}	  
	  // Retrieve the history
	  log.Println(v1,v2,limit,v1hash,v2hash)
	  /* TODO
	  result, err := self.history.Range(int64(v1), v1hash[0], int64(v2), v2hash[0], int64(limit))
	  if err != nil {
		log.Println("Failed retrieving history ", err)
		makeErrorResponse(req.Response, "Failed retrieving history")
		req.FinishSignal <- false
		return
	  }
	  // Send the result by HTTP
	  req.Response.SetHeader("Content-Type", "application/protobuf")
	  _, err = req.Response.Write( []byte(result) )
	  if err != nil {
		log.Println("Failed writing HTTP response")
		req.FinishSignal <- false
		return
	  } */
	req.FinishSignal <- true
	return
  }
	
  makeErrorResponse( req.Response, "Snapshots are not implemented" )
  req.FinishSignal <- false
}

func (self *WaveletNode) pubSub(req* PubSubRequest) {
  seq := strings.Split(req.Filter.Prefix[1:], "/", -1)
  if len(seq) < 1 + self.level && !req.Filter.Recursive {
	panic("The subscription must not reach this node")
	return
  }
  if len(seq) != self.level + 1 {
	return
  }
  // Subscribe/Unsubscribe
  switch req.Action {
	case Subscribe:
	  log.Println("Subscribing node ", self.URI())
	  self.subscriptions[req.Filter.Id] = Subscription{req.Subscriber, req.Filter}
	  // Send a snapshot to the subscriber
	  // TODO req.Subscriber.Update( &UpdateMsg{self.URI(), cloneJsonObject(self.doc)} )
	case Unsubscribe:
	  log.Println("Unsubscribing node ", self.URI())
	  self.subscriptions[req.Filter.Id] = Subscription{}, false
	default:
	  panic("Unsupported PubSub action")
  }
}
