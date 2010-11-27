package wave

import (
  . "lightwave"
  "strings"
  proto "goprotobuf.googlecode.com/hg/proto"  
)

// -------------------------------------------
// WaveletNode

type WaveletNode struct {
  NodeBase
  wavelet *Wavelet
  level int
  subscriptions map[string]Subscription
  // List of domains which participate in federating this document
  federatedDomains map[string]bool
  // TODO history *DocumentHistory
}

func NewWaveletNode(parent Node, name string, level int) *WaveletNode {
  namesplit := strings.Split(name, "$")
  if len(namesplit) != 3 {
	log.Println("Invalid name for a wavelet node ", name)
	return nil
  }
  d := &DocumentNode{children:make(map[string]Node), level:level, NodeBase:NodeBase{parent:parent, name:name,postChannel:make(chan *PostRequest), getChannel:make(chan *GetRequest), stopChannel:make(chan bool), pubSubChannel:make(chan *PubSubRequest)}}
  d.subscriptions = make(map[string]Subscription)  
  wurl := &WaveUrl{namesplit[1],namesplit[0], self.Host().Name(), namesplit[2]}
  d.wavelet = NewWavelet(wurl)
  d.federatedDomains = make(map[string]bool)
  // TODO d.history = NewDocumentHistory()
  return d
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
  if self.HashedVersion.Equals( delta.HashedVersion ) {
	err := self.wavelet.ApplyDelta(delta)
	if err != nil {
	  log.Println("Failed applying delta: ", err)
	  return false
	}
	self.wavelet.HashedVersion.Version += len(delta.Operation)
	// TODO: Calculate new hash

	// Send message to subscribers
	for _, s := range self.subscriptions {
	  // TODO: Send JSON encoding 
	  // s.Subscriber.Update( &UpdateMsg{self.URI(), m})
	}
  } else {
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
	  if u.Domain != server.manifest.Domain {
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
  return self.wavelet.HashedVersion.Version
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
	  switch req.MimeType {
		case "application/protobuf":	  
		  delta := &ProtocolWaveletDelta{}
		  if err := proto.Unmarshal(req.Data, delta); err != nil {
			makeErrorResponse(req.Response, "Cannot parse HTTP body. No valid ProtoBuf or wrong message type")
			req.FinishSignal <- false
			return
		  }
		  if !self.apply(delta) {
			makeErrorResponse(req.Response, "Could not apply document mutation")
			req.FinishSignal <- false
			return
		  }
		  req.FinishSignal <- true
		default:
		  makeErrorResponse(req.Response, "Data type " + req.MimeType + " not allowed for put/post")
		  req.FinishSignal <- false
	  }
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
	  req.Subscriber.Update( &UpdateMsg{self.URI(), cloneJsonObject(self.doc)} )
	case Unsubscribe:
	  log.Println("Unsubscribing node ", self.URI())
	  self.subscriptions[req.Filter.Id] = Subscription{}, false
	default:
	  panic("Unsupported PubSub action")
  }
}
