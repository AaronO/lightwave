package lightwave

import (
  "regexp"
  "log"
  "json"
  "fmt"
  "os"
)

// -------------------------------------
// UserId

type UserId struct {
  Username string
  Domain string
}

func NewUserId(userid string) (result *UserId, err os.Error) {
  r := regexp.MustCompile("^([\\-_A-Za-z0-9.]+)@([\\-_A-Za-z0-9.]+)$")
  if submatches := r.FindStringSubmatch(userid); submatches != nil {
	return &UserId{submatches[1], submatches[2]}, nil
  }
  return nil, os.NewError("Malformed userid")
}

func (self *UserId) String() string {
  return self.Username + "@" + self.Domain
}

// -------------------------------------
// UserNode

type UserNode struct {
  NodeBase
  doc map[string]interface{}
  // The user to which this session belongs
  user string
  digestChannel chan *DigestMsg
  history *DocumentHistory
}

func NewUserNode(parent *UserRootNode, user string) *UserNode {
  u := &UserNode{NodeBase:NodeBase{parent:parent, name:user, postChannel:make(chan *PostRequest), getChannel:make(chan *GetRequest), stopChannel:make(chan bool)}, user:user, doc:make(map[string]interface{})}
  u.digestChannel = make(chan *DigestMsg)
  // Initialize the JSON document
  u.doc["_meta"] = make(map[string]interface{})
  u.doc["_data"] = make(map[string]interface{})
  u.doc["_rev"] = float64(0)
  // TODO
  u.doc["_hash"] = "TODOHASH"  
  u.history = NewDocumentHistory()
  return u
}

func (self *UserNode) Run() {
  for {
	select {
	  case req := <-self.postChannel:
		self.post(req)
	  case req := <-self.getChannel:
		self.get(req)
	  case msg := <-self.digestChannel:
		self.digest(msg)
	  case <-self.stopChannel:
		return
	}
  }
}

func (self *UserNode) Digest(msg *DigestMsg) {
  self.digestChannel <- msg
}

func (self *UserNode) digest(msg *DigestMsg) {
  log.Println("Update for user ", self.user, " inbox from URI ", msg.URI)
  // TODO
}

func (self *UserNode) apply( mutation map[string]interface{} ) bool {
  if !(IsDocumentMutation(mutation)) {
	log.Println("Not a document mutation")
	return false
  }
  m := DocumentMutation(mutation)  

  // TODO: Check that the delta has the right hash
  
  // Apply the mutation at the most recent version of the document?
  if m.AppliedAtRevision() == self.Revision() {
	if !m.Apply(self.doc, NoFlags) {
	  log.Println("Failed applying delta")
	  return false
	}	
  } else if m.AppliedAtRevision() > self.Revision() {
	// Delta from the future -> error
	log.Println("Seen delta from the future")
	return false
  } else {
	// OT is required
	var ot Transformer
	deltas := self.history.Tail( m.AppliedAtRevision() )
	for _, d := range deltas {
	  log.Println("Transforming ", d, m )
	  err := ot.Transform( d.Clone(), m )
	  if err != nil {
		log.Println("OT Error: ", err)
		return false
	  }
	}
  }
  self.doc["_rev"] = float64(self.Revision() + 1)
  log.Println("Resulting version is ", self.Revision())
  // TODO
  m["_endHash"] = "TODOHASH"
  m["_endRev"] =  float64(self.Revision())
  self.history.Append(m)
  return true
}

func (self *UserNode) Revision() int64 {
  return int64(self.doc["_rev"].(float64))
}

func (self *UserNode) post(req *PostRequest) {  
  switch req.MimeType {
	// Posting a json document or a document mutation?
	case "application/json", "application/x-www-form-urlencoded":
	  m := make(map[string]interface{})
	  if err := json.Unmarshal(req.Data, &m); err != nil {
		makeErrorResponse(req.Response, "Cannot parse HTTP body. No valid JSON")
		req.FinishSignal <- false
		return
	  }
	  // It is not allowed to modify the meta data
	  if _, ok := m["_meta"]; ok {
		makeErrorResponse(req.Response, "Attempt to modify meta data using POST")
		req.FinishSignal <- false
		return
	  }
	  // Try to apply the data
	  if !self.apply(m) {
		makeErrorResponse(req.Response, "Could not apply document mutation")
		req.FinishSignal <- false
		return
	  }
	  req.Response.SetHeader("Content-Type", "application/json")
	  if _, err := req.Response.Write( []byte(fmt.Sprintf("{\"ok\":true, \"version\":%d}", self.Revision())) ); err != nil {
		log.Println("Failed writing HTTP response")
		req.FinishSignal <- false
		return
	  }
	  req.FinishSignal <- true
	default:
	  makeErrorResponse(req.Response, "Data type not allowed for post")
	  req.FinishSignal <- false
  }
}

func (self *UserNode) get(req *GetRequest) {  
  json, err := json.Marshal(self.doc)
  if err != nil {
	panic("Failed marshaling to json")
  }
  req.Response.SetHeader("Content-Type", "application/json")
  _, err = req.Response.Write( json )
  if err != nil {
	log.Println("Failed writing HTTP response")
	req.FinishSignal <- false
	return
  }
  req.FinishSignal <- true
}

// -------------------------------------
// UserRootNode

type UserRootNode struct {
  NodeBase
  users map[string]*UserNode
  digestChannel chan *DigestMsg
}

func NewUserRootNode(server *Server) *UserRootNode {
  return &UserRootNode{NodeBase:NodeBase{parent:server, name:"_user", postChannel:make(chan *PostRequest), getChannel:make(chan *GetRequest), stopChannel:make(chan bool)}, users:make(map[string]*UserNode), digestChannel:make(chan *DigestMsg)}
}

func (self *UserRootNode) Run() {
  for {
	select {
	  case req := <-self.postChannel:
		self.post(req)
	  case req := <-self.getChannel:
		self.get(req)
	  case req := <-self.digestChannel:
		self.digest(req)
	  case <-self.stopChannel:
		return
	}
  }
}

func (self *UserRootNode) Digest(msg *DigestMsg) {
  self.digestChannel <- msg
}

func (self *UserRootNode) post(req *PostRequest) {
  uri := req.URI.(UserURI)
  // Check for the session node
  u, ok := self.users[uri.User]
  if ok {
	u.Post(req)
	return
  }
  u = NewUserNode(self, uri.User)
  oldfinish := req.FinishSignal
  finish := make(chan bool)
  req.FinishSignal = finish
  go u.Run()
  u.Post(req)
  ok = <-finish
  if ok {
	self.addChild(u)
  } else {
	u.Stop()
  }
  oldfinish <- ok
}

func (self *UserRootNode) get(req *GetRequest) {  
  uri := req.URI.(UserURI)
  // Check for the session node
  u, ok := self.users[uri.User]
  if !ok {
	makeErrorResponse(req.Response, "The user does not exist")
	req.FinishSignal <- false
	return
  }
  u.Get(req)
}

func (self *UserRootNode) digest(msg *DigestMsg) {  
  // Check for the session node
  u, ok := self.users[msg.User]
  if !ok {
	log.Println("The user does not exist")
	return
  }
  u.Digest(msg)
}

func (self *UserRootNode) addChild(child *UserNode) {
  self.users[ child.Name() ] = child
}
