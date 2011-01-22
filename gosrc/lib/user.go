package lightwave

import (
  "regexp"
  "log"
  "os"
  "strings"
  "http"
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
// Inbox

type Inbox struct {
  DocumentNode
  digestChannel chan *DigestMsg
}

func NewInbox(parent Node) *Inbox {
  i := &Inbox{DocumentNode:*NewDocumentNode(parent, "inbox", 3), digestChannel:make(chan *DigestMsg)}
  data := i.doc["_data"].(map[string]interface{})
  if _, ok := data["docs"]; !ok {
    mut := make(map[string]interface{})
    mut["_rev"] = float64(0)
    mut["_hash"] = "TODOHash"
    datamut := NewObjectMutation()
    datamut["docs"] = [0]interface{}{}[:]
    mut["_data"] = datamut
    i.apply(mut)
    // data["docs"] = [0]interface{}{}[:]
  }
  return i
}

func (self *Inbox) Run() {
  for {
    select {
      case req := <-self.postChannel:
        self.post(req)
      case req := <-self.getChannel:
        self.get(req)
      case req := <-self.pubSubChannel:
        self.pubSub(req)        
      case msg := <-self.digestChannel:
        self.digest(msg)
      case <-self.stopChannel:
        return
    }
  }
}

func (self *Inbox) Digest(msg *DigestMsg) {
  self.digestChannel <- msg;
}

func (self *Inbox) post(req *PostRequest) {  
  makeErrorResponse(req.Response, "Posting to the inbox is not allowed")
  req.FinishSignal <- false
}

func (self *Inbox) docs() []interface{} {
  return self.doc["_data"].(map[string]interface{})["docs"].([]interface{})
}

func (self *Inbox) digest(msg *DigestMsg) {
  log.Println("DIGEST: Got digest message for " + msg.URI + " authors " + msg.Authors)
  mut := make(map[string]interface{})
  mut["_rev"] = float64(self.Revision())
  arraymut := make([]interface{}, 0, 3)
  // Is the document already in the list?
  found := false
  lst := self.docs()  
  for i, doc := range lst {
    docmap := doc.(map[string]interface{})
    if msg.URI != docmap["uri"].(string) {
      continue
    }
    if i > 0 {
      arraymut = append(arraymut, NewSkipMutation(i))
    }
    digmut := NewObjectMutation()
    digmut["uri"] = msg.URI
    digmut["digest"] = msg.Digest
    digmut["authors"] = msg.Authors
    digmut["msgcount"] = msg.MessageCount
    arraymut = append(arraymut, digmut)
    if len(lst) > i + 1 {
      arraymut = append(arraymut, NewSkipMutation(len(lst) - i - 1))
    } 
    found = true
    break
  }
  if !found {
    digmut := make(map[string]interface{})
    digmut["uri"] = msg.URI
    digmut["digest"] = msg.Digest
    digmut["authors"] = msg.Authors
    digmut["msgcount"] = msg.MessageCount
    arraymut = append(arraymut, digmut)
    if len(lst) > 0 {
      arraymut = append(arraymut, NewSkipMutation(len(lst)))
    }
  }
  if !msg.IsSubscribed {
    // Subscribe to this document to receive further digest data
    self.Server().PubSub( &PubSubRequest{Action:SubscribeDigest, Filter:&NodeFilter{User:self.parent.Name(), Prefix:msg.URI, Recursive:false}} )
  }
  datamut := NewObjectMutation()
  datamut["docs"] = NewArrayMutation(arraymut)
  mut["_data"] = datamut
  ok := self.apply(mut)
  if !ok {
    panic("Failed to apply mutation for meta data of " + self.Name())
  }
}

// -------------------------------------
// UserNode

type UserNode struct {
  DocumentNode
  inbox *Inbox
}

func NewUserNode(parent *UserRootNode, user string) *UserNode {
  u := &UserNode{DocumentNode:*NewDocumentNode(parent, user, 2)}
  u.inbox = NewInbox(u)
  go u.inbox.Run()
  u.children[u.inbox.Name()] = u.inbox
  return u
}

func (self *UserNode) Digest(msg *DigestMsg) {
  log.Println("DIGEST")
  self.inbox.digestChannel <- msg
}

func (self *UserNode) Run() {
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

func (self *UserNode) get(req *GetRequest) {
  docuri := req.URI.(DocumentURI)
  
  // Request is aimed at this document?
  if req.RawQuery != "" && len(docuri.NameSeq) == self.level {
    // Parse the query
    query, err := http.ParseQuery(req.RawQuery)
    if err != nil {
      log.Println("Failed parsing query")
      makeErrorResponse(req.Response, "Failed parsing query")
      req.FinishSignal <- false
      return
    }
    kind, ok := query["kind"]
    if !ok || ( kind != nil && len(kind) != 1 ) {
      log.Println("Missing kind in query or malformed kind")
      makeErrorResponse(req.Response, "Failed parsing query")
      req.FinishSignal <- false
      return
    }
    switch kind[0] {
    case "friends":
      // TODO
      reply := []byte("{\"friends\":[{\"displayName\":\"Tux\", \"userid\":\"tux@localhost\"},{\"displayName\":\"Konqi\", \"userid\":\"konqi@localhost\"}]}")
      req.Response.SetHeader("Content-Type", "application/json")
      _, err = req.Response.Write( reply )
      if err != nil {
	log.Println("Failed writing HTTP response")
	req.FinishSignal <- false
	return
      }
    default:
      log.Println("Unknown kind in query or malformed kind")
      makeErrorResponse(req.Response, "Failed parsing query")
      req.FinishSignal <- false
      return
    }
    req.FinishSignal <- true
    return
  } else {
    // Forward request to the default implementation
    self.DocumentNode.get(req)
  }
}

// -------------------------------------
// UserRootNode

type UserRootNode struct {
  NodeBase
  users map[string]*UserNode
  digestChannel chan *DigestMsg
}

func NewUserRootNode(parent Node) *UserRootNode {
  return &UserRootNode{NodeBase:NodeBase{parent:parent, name:"_user", postChannel:make(chan *PostRequest), getChannel:make(chan *GetRequest), stopChannel:make(chan bool), pubSubChannel:make(chan *PubSubRequest)}, users:make(map[string]*UserNode), digestChannel:make(chan *DigestMsg)}
}

func (self *UserRootNode) Run() {
  for {
    select {
      case req := <-self.postChannel:
        self.post(req)
      case req := <-self.getChannel:
        self.get(req)
      case req := <-self.pubSubChannel:
        self.pubSub(req)                
      case req := <-self.digestChannel:
        self.digest(req)
      case <-self.stopChannel:
        return
    }
  }
}

func (self *UserRootNode) CreateUser(name string) *UserNode {
  if _, ok := self.users[name]; ok {
    log.Println("User already exists")
    return nil
  }
  u := NewUserNode(self, name)
  go u.Run()
  self.addChild(u)
  return u
}

func (self *UserRootNode) Digest(msg *DigestMsg) {
  self.digestChannel <- msg
}

func (self *UserRootNode) post(req *PostRequest) {
  uri := req.URI.(DocumentURI)
  if len(uri.NameSeq) < 2 {
    makeErrorResponse(req.Response, "Cannot get _user")
    req.FinishSignal <- false
    return
  }
  // Check for the user node
  u, ok := self.users[uri.NameSeq[1]]
  if ok {
    u.Post(req)
    return
  }
  u = NewUserNode(self, uri.NameSeq[1])
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
  uri := req.URI.(DocumentURI)
  if len(uri.NameSeq) < 2 {
    makeErrorResponse(req.Response, "Cannot post _user")
    req.FinishSignal <- false
    return
  }
  // Check for the user node
  u, ok := self.users[uri.NameSeq[1]]
  if !ok {
    makeErrorResponse(req.Response, "The user does not exist")
    req.FinishSignal <- false
    return
  }
  u.Get(req)
}

func (self *UserRootNode) pubSub(req* PubSubRequest) {
  seq := strings.Split(req.Filter.Prefix[1:], "/", -1)
  if len(seq) < 3 {
    log.Println("Malformed NodeFilter prefix")
    return
  }
  // Send to all host nodes
  node, ok := self.users[seq[2]]
  if ok {
    node.PubSub(req)
  }  
}

func (self *UserRootNode) digest(msg *DigestMsg) {  
  // Check for the user node
  u, ok := self.users[msg.User]
  if !ok {
    log.Println("The user ", msg.User, " does not exist")
    return
  }
  u.Digest(msg)
}

func (self *UserRootNode) addChild(child *UserNode) {
  self.users[ child.Name() ] = child
}

// -------------------------------------
// UserDB

type UserAccount struct {
  EMail string
  // The name of the user without the domain postfix 
  Username string
  Password string
  DisplayName string
}

type UserAccountDB struct {
  server *Server
  users map[string]*UserAccount
}

func NewUserAccountDB(server *Server) *UserAccountDB {
  return &UserAccountDB{server:server, users: make(map[string]*UserAccount)}
}

func (self *UserAccountDB) FindUser(username string) (*UserAccount, os.Error) {
  user, ok := self.users[username]
  if !ok {
    return nil, os.NewError("Unknown user")
  }
  return user, nil
}

func (self *UserAccountDB) SignUpUser(email string, displayname string, username string, password string) (*UserAccount, os.Error) {
  _, ok := self.users[username]
  if ok {
    return nil, os.NewError("User already exists");
  }
  user := &UserAccount{EMail:email, Username:username, Password:password, DisplayName:displayname}
  self.users[username] = user
  return user, nil
}

func (self *UserAccountDB) CheckCredentials(username string, password string) os.Error {
  user, ok := self.users[username]
  if !ok {
    return os.NewError("User does not exists");
  }
  if user.Password != password {
    return os.NewError("Wrong password");
  }
  return nil
}
