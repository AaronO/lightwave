package lightwave

import (
  "regexp"
  "log"
  "os"
  "strings"
  "http"
  sqlite "gosqlite.googlecode.com/hg/sqlite"
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
  DocumentNode
}

func NewUserNode(parent *UserRootNode, user string) *UserNode {
  u := &UserNode{DocumentNode:*NewDocumentNode(parent, user, 2)}
  return u
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
      reply := []byte("{\"friends\":[{\"displayName\":\"Torben\", \"userid\":\"weis@localhost\"}, {\"displayName\":\"Tux\", \"userid\":\"tux@localhost\"},{\"displayName\":\"Konqi\", \"userid\":\"konqi@localhost\"}]}")
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
}

func NewUserRootNode(parent Node) *UserRootNode {
  return &UserRootNode{NodeBase:NodeBase{parent:parent, name:"_user", postChannel:make(chan *PostRequest), getChannel:make(chan *GetRequest), stopChannel:make(chan bool), pubSubChannel:make(chan *PubSubRequest)}, users:make(map[string]*UserNode)}
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
      case <-self.stopChannel:
        return
    }
  }
}

func (self *UserRootNode) CreateUser(name, displayName, email string) *UserNode {
  if _, ok := self.users[name]; ok {
    log.Println("User already exists")
    return nil
  }
  u := NewUserNode(self, name)
  
  mut := make(map[string]interface{})
  mut["_rev"] = float64(u.Revision())
  datamut := NewObjectMutation()
  datamut["userid"] = name + "@" + self.Server().Capabilities().Domain
  datamut["displayName"] = displayName
  datamut["email"] = email
  datamut["image"] = "/images/unknown.png"
  // HACK
  datamut["friends"] = []interface{}{"tux@localhost", "konqi@localhost"}
  mut["_data"] = datamut
  metamut := NewObjectMutation()
  metamut["schema"] = "//lightwave/user"
  mut["_meta"] = metamut
  u.apply(mut)

  go u.Run()
  self.addChild(u)
  return u
}

func (self *UserRootNode) post(req *PostRequest) {
  uri := req.URI.(DocumentURI)
  if len(uri.NameSeq) < 2 {
    makeErrorResponse(req.Response, "Cannot post to _user")
    req.FinishSignal <- false
    return
  }
  // Check for the user node
  var u Node
  u, ok := self.users[uri.NameSeq[1]]
  if !ok {
    u = self.loadDocument(uri.NameSeq[1])
  }
  if u == nil {
    makeErrorResponse(req.Response, "No such user")
    req.FinishSignal <- false
  }
  u.Post(req)
}

func (self *UserRootNode) get(req *GetRequest) {  
  uri := req.URI.(DocumentURI)
  if len(uri.NameSeq) < 2 {
    makeErrorResponse(req.Response, "Cannot get _user")
    req.FinishSignal <- false
    return
  }
  // Check for the user node
  var u Node
  u, ok := self.users[uri.NameSeq[1]]
  if !ok {
    u = self.loadDocument(uri.NameSeq[1])
  }
  if u == nil {
    makeErrorResponse(req.Response, "The user does not exist")
    req.FinishSignal <- false
    return
  }
  u.Get(req)
}

func (self *UserRootNode) pubSub(req* PubSubRequest) {
  seq := strings.Split(req.DocumentURI[1:], "/", -1)
  if len(seq) < 3 {
    log.Println("Malformed NodeFilter prefix")
    return
  }
  // Send to all host nodes
  var node Node
  node, ok := self.users[seq[2]]
  if !ok {
    node = self.loadDocument(seq[2])
  }
  if node != nil {
    node.PubSub(req)
  }  
}

func (self *UserRootNode) loadDocument(name string) Node {
  if !isDocumentPersisted(self.Server(), self.URI() + "/" + name) {
    return nil
  }
  doc := NewUserNode(self, name)
  go doc.Run()
  self.addChild(doc)
  return doc
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
  dbcon *sqlite.Conn
}

func NewUserAccountDB(server *Server) *UserAccountDB {
  db := &UserAccountDB{server:server}
  dbcon, err := sqlite.Open(server.Capabilities().Domain + "_user.db")
  if err != nil {
    panic("Cannot access user database")
  }
  stmnt, err := dbcon.Prepare("CREATE TABLE IF NOT EXISTS Accounts ( id VARCHAR(20) PRIMARY KEY, email VARCHAR(30), nick VARCHAR(30), passwd VARCHAR(16) )")
  if err != nil {
    panic("Cannot prepare stmnt for create account table")
  }
  err = stmnt.Exec()
  if err != nil {
    panic("Cannot create account table")
  }
  stmnt.Next()
  db.dbcon = dbcon
  return db
}

func (self *UserAccountDB) FindUser(username string) (*UserAccount, os.Error) {
  // user, ok := self.users[username]
  // if !ok {
  //   return nil, os.NewError("Unknown user")
  // }
  stmnt, err := self.dbcon.Prepare("SELECT * FROM Accounts WHERE id = ?1")
  if err != nil {
    log.Println(err)
    return nil, err
  }
  err = stmnt.Exec(username)
  if err != nil {
    return nil, err
  }
  if !stmnt.Next() {
    return nil, os.NewError("Unknown user")
  }
  var user UserAccount
  err = stmnt.Scan(&user.Username, &user.EMail, &user.DisplayName, &user.Password)
  if err != nil {
    return nil, err
  }  
  return &user, nil
}

func (self *UserAccountDB) SignUpUser(email string, displayname string, username string, password string) (*UserAccount, os.Error) {
  _, err := self.FindUser(username)
  if err == nil {
    return nil, os.NewError("User already exists");
  }
  user := &UserAccount{EMail:email, Username:username, Password:password, DisplayName:displayname}
  stmnt, err := self.dbcon.Prepare("INSERT INTO Accounts VALUES ( ?1, ?2, ?3, ?4 )")
  if err != nil {
    return nil, err
  }
  err = stmnt.Exec(username, email, displayname, password)
  if err != nil {
    return nil, err
  }  
  stmnt.Next()
  
  // Create a document for this user
  self.server.LocalHost().Users().CreateUser(username, displayname, email)
  return user, nil
}

func (self *UserAccountDB) CheckCredentials(username string, password string) os.Error {
  user, err := self.FindUser(username)
  if err != nil {
    return os.NewError("User does not exists");
  }
  if user.Password != password {
    return os.NewError("Wrong password");
  }
  return nil
}
