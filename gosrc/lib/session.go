package lightwave

import (
  "log"
  "json"
  "fmt"
  "http"
  "strings"
  "strconv"
  "time"
  hmac "crypto/hmac"
  "os"
)

// ------------------------------------------------------
// SessionData

type SessionData struct {
  Filters map[string]*NodeFilter
}

func (self *SessionData) contains(id string) bool {
  for _, f := range self.Filters {
	if f.Id == id {
	  return true
	}
  }
  return false
}

func (self *SessionData) find(id string) *NodeFilter {
  for _, f := range self.Filters {
	if f.Id == id {
	  return f
	}
  }
  return nil
}

// ------------------------------------------------------
// NodeFilter

// A NodeFilter selects a set of document nodes based on their URI, schema and mime type.
type NodeFilter struct {
  Id string "_id"
  Rev float64 "_rev"
  Prefix string
  // If true, all children of matching document nodes are considered.
  Recursive bool
  // Restrictions on the desired mime type
  MimeTypes []string
  // Restrictions on the desired schema
  Schemas []string
  // Determines whether the server might send a snapshot of the current state
  Snapshot bool
  // The user that creaeted this filter
  User string
}

func NewNodeFilter(prefix string, val interface{}) *NodeFilter {
  m, ok := val.(map[string]interface{})
  if !ok {
	return nil
  }
  f := &NodeFilter{}
  bytes, _ := json.Marshal(m)
  err := json.Unmarshal(bytes, f)
  if err != nil {
	log.Println(err)
	return nil
  }
  f.Prefix = prefix
  return f
}

// ------------------------------------------------------
// SessionNode

// Represents a session
type SessionNode struct {
  NodeBase
  // The user to which this session belongs
  user string
  // The session ID
  session string
  // The content of the session
  doc map[string]interface{}
  // The key is the URI of a node that has sent an update.
  // The value is the document describing this update, i.e. a serialized JSON document
  queue map[string][]string
  // This channel is used by document nodes to send updates to the session
  updateChannel chan *UpdateMsg
  poll *GetRequest
}

func NewSessionNode(parent *SessionRootNode, user string, session string) *SessionNode {
  s := &SessionNode{NodeBase:NodeBase{parent:parent, name:user + "/" + session, postChannel:make(chan *PostRequest), getChannel:make(chan *GetRequest), stopChannel:make(chan bool)}, user:user, session:session, doc:make(map[string]interface{})}
  s.queue = make(map[string][]string)
  s.updateChannel = make(chan *UpdateMsg)
  s.doc["_meta"] = make(map[string]interface{})
  s.doc["_data"] = make(map[string]interface{})
  s.doc["_rev"] = float64(0)
  return s
}

func (self *SessionNode) Server() *Server {
  return self.parent.(*SessionRootNode).parent.(*Server)
}

func (self *SessionNode) Run() {
  for {
	select {
	  case req := <-self.postChannel:
		self.post(req)
	  case req := <-self.getChannel:
		self.get(req)
	  case msg := <-self.updateChannel:
		self.update(msg)
	  case <-self.stopChannel:
		return
	}
  }
}

// Subscriber interface
func (self *SessionNode) Update(msg *UpdateMsg) {
  self.updateChannel <- msg
}

func (self *SessionNode) marshalQueue() []byte {
  str := "{"
  for name, mutlist := range self.queue {
	if str != "{" {
	  str += ","
	}
	q := "["
	for _, m := range mutlist {
	  if ( q != "[" ) {
		q += ","
	  }
	  q += m
	}
	q += "]"
	str += fmt.Sprintf(`"%v":%v`, name, q)
  }
  str += "}"
  return []byte(str)
}

func (self *SessionNode) update(msg *UpdateMsg) {
  log.Println("Update for session ", self.session, " from URI ", msg.URI)
  if self.queue == nil {
	self.queue = make(map[string][]string)
  }
  lst, ok := self.queue[msg.URI]
  if !ok {
	lst = msg.Mutation
	self.queue[msg.URI] = lst
  } else {
	self.queue[msg.URI] = append(lst, msg.Mutation...)
  }
  
  // Is somebody polling?
  if self.poll != nil {
	self.poll.Response.SetHeader("Content-Type", "application/json")
	_, err := self.poll.Response.Write( self.marshalQueue() )
	if err != nil {
	  self.poll = nil;
	  log.Println("Failed writing HTTP response")
	  self.poll.FinishSignal <- false
	  return
	}
	self.queue = nil
	self.poll.FinishSignal <- false
	self.poll = nil
  }
}

// Create a SessionData object from the JSON representation
func (self *SessionNode) parseSessionData() *SessionData {
  s := new(SessionData)
  s.Filters = make(map[string]*NodeFilter)
  d, ok := self.doc["_data"]
  if !ok {
	return s
  }
  f, ok := d.(map[string]interface{})["filters"]
  if !ok {
	return s
  }
  filters, ok := f.(map[string]interface{})
  if !ok {
	return s
  }
  for prefix, val := range filters {
	filter := NewNodeFilter(prefix, val)
	if filter == nil {
	  continue
	}
	filter.User = self.user;
	s.Filters[prefix] = filter
  }
  return s
}

func (self *SessionNode) apply( mutation map[string]interface{} ) bool {
  if !(IsDocumentMutation(mutation)) {
	log.Println("Not a document mutation")
	return false
  }
  m := DocumentMutation(mutation)

  olddata := self.parseSessionData()
  if olddata == nil {
	panic("Cannot parse my own data")
  }
  
  if m.AppliedAtRevision() == self.Revision() {
	rev := float64(self.Revision() + 1)
	m["_rev"] = rev
    if err := m.Apply(self.doc, CreateIDs); err != nil {
	  log.Println("Failed applying delta")
	  return false
	}
	self.doc["_rev"] = rev
	
	newdata := self.parseSessionData()
	if newdata == nil {
	  log.Println("The last delta messed up the session. The session is now broken")
	  return false;
	  // TODO: unsubscribe everything and mark the session as dead
	  // TODO: This could be avoided with a schema checker
	}
	// Search for new filters
	for _, filter := range newdata.Filters {
	  if !olddata.contains(filter.Id) {
		log.Println("NEW FILTER ", filter)
		self.Server().PubSub( &PubSubRequest{Action:Subscribe, Filter:filter, Subscriber:self} )
	  } else if filter.Rev == rev {
		log.Println("MODIFIED FILTER ", filter)
		self.Server().PubSub( &PubSubRequest{Action:Unsubscribe, Filter:olddata.find(filter.Id), Subscriber:self} )
		self.Server().PubSub( &PubSubRequest{Action:Subscribe, Filter:filter, Subscriber:self} )
	  }
	}
	// Search for deleted filters
	for _, filter := range olddata.Filters {
	  if  !newdata.contains(filter.Id) {
		log.Println("OLD FILTER ", filter)
		self.Server().PubSub( &PubSubRequest{Action:Unsubscribe, Filter:filter, Subscriber:self} )
	  }
	}
  } else {
	// Sessions must be updates at the latest revision always. No OT is performed
	log.Println("Sessions must be updated at the latest revision always. No OT is performed")
	return false
  }
  return true
}

func (self *SessionNode) Revision() int64 {
  return int64(self.doc["_rev"].(float64))
}

func (self *SessionNode) post(req *PostRequest) {  
  uri := req.URI.(SessionURI)
  if uri.Special != "" {
	makeErrorResponse(req.Response, "Not allowed to post to this URL")
	req.FinishSignal <- false
	return
  }
  
  switch req.MimeType {
	// Posting a json document or a document mutation?
	case "application/json", "application/x-www-form-urlencoded":
	  m := make(map[string]interface{})
	  if err := json.Unmarshal(req.Data, &m); err != nil {
		makeErrorResponse(req.Response, "Cannot parse HTTP body. No valid JSON")
		req.FinishSignal <- false
		return
	  }
	  // It is not allowed to modify the meta data1
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

func (self *SessionNode) get(req *GetRequest) {  
  uri := req.URI.(SessionURI)
  if uri.Special != "" {
	switch uri.Special {
	  case "_update":
		req.Response.SetHeader("Content-Type", "application/json")
		_, err := req.Response.Write( self.marshalQueue() )
		if err != nil {
		  log.Println("Failed writing HTTP response")
		  req.FinishSignal <- false
		  return
		}
		self.queue = nil
		req.FinishSignal <- true
	  case "_poll":
		log.Println("Polling")
		if self.queue != nil && len(self.queue) != 0 {
		  log.Println("Poll response")
		  req.Response.SetHeader("Content-Type", "application/json")
		  _, err := req.Response.Write( self.marshalQueue() )
		  if err != nil {
			log.Println("Failed writing HTTP response")
			req.FinishSignal <- false
			return
		  }
		  self.queue = nil
		  req.FinishSignal <- true
		} else {
		  log.Println("Poll wait")
		  if self.poll != nil {
			self.poll.Response.SetHeader("Content-Type", "application/json")
			self.poll.Response.Write( []byte("{}") )
			self.poll.FinishSignal <- true
		  }
		  self.poll = req
		}
	  default:
		// TODO: Return a 404 instead
		makeErrorResponse(req.Response, "Unknown URL")
		req.FinishSignal <- false
	}
	return
  }
  
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

// ------------------------------------------------------
// SessionRootNode

type SessionRootNode struct {
  NodeBase
  sessions map[string]*SessionNode
}

func NewSessionRootNode(server *Server) *SessionRootNode {
  return &SessionRootNode{NodeBase:NodeBase{parent:server, name:"_session", postChannel:make(chan *PostRequest), getChannel:make(chan *GetRequest), stopChannel:make(chan bool)}, sessions:make(map[string]*SessionNode)}
}

func (self *SessionRootNode) Run() {
  for {
	select {
	  case req := <-self.postChannel:
		self.post(req)
	  case req := <-self.getChannel:
		self.get(req)
	  case <-self.stopChannel:
		return
	}
  }
}

func (self *SessionRootNode) post(req *PostRequest) {
  uri := req.URI.(SessionURI)
  name := uri.User + "/" + uri.Name  
  // Check for the session node
  s, ok := self.sessions[name]
  if ok {
	s.Post(req)
	return
  }
  s = NewSessionNode(self, uri.User, uri.Name)
  oldfinish := req.FinishSignal
  finish := make(chan bool)
  req.FinishSignal = finish
  go s.Run()
  s.Post(req)
  ok = <-finish
  if ok {
	self.addChild(s)
  } else {
	s.Stop()
  }
  oldfinish <- ok
}

func (self *SessionRootNode) get(req *GetRequest) {  
  uri := req.URI.(SessionURI)
  name := uri.User + "/" + uri.Name  
  // Check for the session node
  s, ok := self.sessions[name]
  if !ok {
	makeErrorResponse(req.Response, "The session does not exist")
	req.FinishSignal <- false
	return
  }
  s.Get(req)
}

func (self *SessionRootNode) addChild(child *SessionNode) {
  self.sessions[ child.Name() ] = child
}

// ----------------------------------------------------
// Some constants

const (
  SessionDuration = 60 * 60 * 24
  ServerSecret = "MakeMeASecret"
)

// ----------------------------------------------------
// Helper function, thanks to web.go

/*
func getCookieSig(val []byte, timestamp string) string {
  hm := hmac.NewSHA1( []byte(ServerSecret) )
  hm.Write(val)
  hm.Write([]byte(timestamp))

  hex := fmt.Sprintf("%02x", hm.Sum())
  return hex
}

func (ctx *Context) encodeSecureCookie(val string, creationTime int64) {
  var buf bytes.Buffer
  encoder := base64.NewEncoder(base64.StdEncoding, &buf)
  encoder.Write([]byte(val))
  encoder.Close()
  timestamp := strconv.Itoa64(creationTime)
  sig := getCookieSig(buf.Bytes(), timestamp)
  return strings.Join([]string{buf.String(), timestamp, sig}, "|")
}

func decodeSecureCookie(string value, int64 maxAge) (string, os.Error) {
  parts := strings.Split(value, "|", 3)
  val := parts[0]
  timestamp := parts[1]
  sig := parts[2]
  // Check signature
  if getCookieSig(ServerSecret, []byte(val), timestamp) != sig {
    return "", os.NewError("Signature error, cookie is invalid")
  }
  // Check time stamp
  ts, _ := strconv.Atoi64(timestamp)
  if ts + maxAge < time.UTC().Seconds() {
    return "", os.NewError("Cookie is outdated")
  }

  buf := bytes.NewBufferString(val)
  encoder := base64.NewDecoder(base64.StdEncoding, buf)
  res, _ := ioutil.ReadAll(encoder)
  return string(res), nil
}
*/

func parseCookies(req *http.Request) (map[string]string, os.Error) {
  result := make(map[string]string)

  if v, ok := req.Header["Cookie"]; ok {
    cookies := strings.Split(v, ";", -1)
    for _, cookie := range cookies {
      cookie = strings.TrimSpace(cookie)
      parts := strings.Split(cookie, "=", 2)
      if len(parts) != 2 {
        continue
      }
      result[parts[0]] = parts[1]
    }
  }

  return result, nil
}

func webTime(t *time.Time) string {
  ftime := t.Format(time.RFC1123)
  if strings.HasSuffix(ftime, "UTC") {
    ftime = ftime[0:len(ftime)-3] + "GMT"
  }
  return ftime
}

var sessionIdCounter int64 = 1

func createSessionId(username string) string {
  hm := hmac.NewSHA1( []byte(ServerSecret) )
  hm.Write([]byte(username))
  hm.Write([]byte(strconv.Itoa64(sessionIdCounter)))
  hm.Write([]byte(strconv.Itoa64(time.Seconds())))
  sessionIdCounter++
  hex := fmt.Sprintf("%02x", hm.Sum())
  return hex
}

// ------------------------------------------------------
// Session

type Session struct {
  // The name of the user without the domain postfix 
  Id string
  Username string
  Cookie string
  CreationTime int64
  sessionDatabase *SessionDB
}

func NewSession(sessionDatabase *SessionDB, username string) *Session {
  s := &Session{Username:username}
  s.Id = createSessionId(username)
  s.CreationTime = time.UTC().Seconds()
  s.Cookie = s.Id
  s.sessionDatabase = sessionDatabase
  return s
}

func (self *Session) ExpirationTime() *time.Time {
  return time.SecondsToUTC(self.CreationTime + SessionDuration)
}

func (self *Session) SetCookie(writer http.ResponseWriter) {
  cookie := fmt.Sprintf("sid=%s; expires=%s", self.Cookie, webTime(self.ExpirationTime()))
  writer.SetHeader("Set-Cookie", cookie)
}

/*
 * @return information about this session
 */
func (self *Session) InfoHandler(w http.ResponseWriter, r *http.Request) os.Error {
  // Find the user of this session
  user, err := self.sessionDatabase.server.UserAccountDatabase.FindUser(self.Username)
  if err != nil {
    makeErrorResponse(w, err.String())
    return err
  }
  
  result := make(map[string]string)
  result["sid"] = self.Id;
  result["user"] = self.Username
  result["displayName"] = user.DisplayName
  result["domain"] = self.sessionDatabase.server.Capabilities().Domain
  json, err := json.Marshal(result)
  if err != nil {
    log.Println("Failed marshaling to json")
    makeErrorResponse(w, "Failed marshaling to json")
    return err
  }
  w.SetHeader("Content-Type", "application/json")
  _, err = w.Write( json )
  if err != nil {
    log.Println("Failed writing HTTP response")
    makeErrorResponse(w, "Failed writing HTTP response")
    return err
  }
  return nil
}

// ------------------------------------------------------
// SessionDB

type SessionDB struct {
  sessions map[string]*Session
  server *Server
}

func NewSessionDB(server *Server) *SessionDB {
  return &SessionDB{server:server, sessions:make(map[string]*Session)}
}

func (self *SessionDB) CreateSession(username string) (*Session, os.Error) {
  // TODO avoid that one user is creating too many concurrent sessions
  s := NewSession(self, username)
  self.sessions[s.Id] = s;
  return s, nil
}

func (self *SessionDB) FindSession(req *http.Request) (*Session, os.Error) {
  cookies, err := parseCookies(req)
  if err != nil {
    return nil, err
  }
  sid, ok := cookies["sid"]
  if !ok {
    return nil, os.NewError("No SID cookie")
  }
  return self.findSession(sid)
}

func (self *SessionDB) DeleteSession(session *Session) {
  log.Println("Deleting session ", session.Id)
  self.sessions[session.Id] = nil, false
}

func (self *SessionDB) findSession(cookie string) (*Session, os.Error) {
  session, ok := self.sessions[cookie]
  if !ok {
    return nil, os.NewError("No such session")
  }
  return session, nil
}
