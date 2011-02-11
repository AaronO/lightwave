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
  vec "container/vector"
  "os"
  "sync"
)

// ----------------------------------------------------
// Some constants

const (
  SessionDuration = 60 * 60 * 24
  ServerSecret = "MakeMeASecret"
)

// ----------------------------------------------------
// View

type View struct {
  session *Session
  id string
  queryId string
  revision int64
  keys vec.StringVector
}

func NewView(session *Session, viewId string, mapping DocumentMappingId, withTags []string, withoutTags []string) *View {
  r := &View{session:session, id:viewId, revision:0}
  r.queryId = session.Id + ":" + viewId
  session.sessionDatabase.server.Indexer.Query(r.queryId, r, mapping, withTags, withoutTags)
  return r
}

func (self *View) stop() {
  self.session.sessionDatabase.server.Indexer.StopQuery(self.queryId)
}

func (self *View) hasKey(key string) bool {
  for _, k := range self.keys {
    if k == key {
      return true
    }
  }
  return false
}

func (self *View) AddResult(queryId string, key string, value interface{}) {
  if self.hasKey(key) {
    self.updateResult(queryId, key, value)
    return
  }
  log.Println("VIEW ADD", key, value)
  mut := make(map[string]interface{})
  mut["_rev"] = float64(self.revision)
  self.revision++
  mut["_endRev"] = float64(self.revision)
  arraymut := make([]interface{}, 0, 2)
  digmut := make(map[string]interface{})
  jsonValue := make(map[string]interface{})
  json.Unmarshal([]byte(value.(string)), &jsonValue)
  digmut["value"] = jsonValue
  digmut["key"] = key
  arraymut = append(arraymut, digmut)
  if len(self.keys) > 0 {
    arraymut = append(arraymut, NewSkipMutation(len(self.keys)))
  }
  self.keys.Insert(0, key)
  datamut := NewObjectMutation()
  datamut["items"] = NewArrayMutation(arraymut)
  mut["_data"] = datamut
  self.send(mut)
}

func (self *View) updateResult(queryId string, key string, value interface{}) {
  log.Println("VIEW UPDATE", key, value)
  mut := make(map[string]interface{})
  mut["_rev"] = float64(self.revision)
  self.revision++
  mut["_endRev"] = float64(self.revision)
  arraymut := make([]interface{}, 0, 3)
  // What is the position of this key
  for i, k := range self.keys {
    if key != k {
      continue
    }
    if i > 0 {
      arraymut = append(arraymut, NewSkipMutation(i))
    }
    digmut := NewObjectMutation()
    jsonValue := make(map[string]interface{})
    json.Unmarshal([]byte(value.(string)), &jsonValue)
    digmut["value"] = jsonValue
    arraymut = append(arraymut, digmut)
    if len(self.keys) > i + 1 {
      arraymut = append(arraymut, NewSkipMutation(len(self.keys) - i - 1))
    }
    datamut := NewObjectMutation()
    datamut["items"] = NewArrayMutation(arraymut)
    mut["_data"] = datamut
    self.send(mut)
    return
  }
}

func (self *View) DeleteResult(queryId string, key string) {
  log.Println("VIEW DELETE", key)
  mut := make(map[string]interface{})
  mut["_rev"] = float64(self.revision)
  mut["_endRev"] = float64(self.revision + 1)
  arraymut := make([]interface{}, 0, 3)
  // What is the position of this key
  for i, k := range self.keys {
    if key != k {
      continue
    }
    if i > 0 {
      arraymut = append(arraymut, NewSkipMutation(i))
    }
    arraymut = append(arraymut, NewDeleteMutation(1))
    if len(self.keys) > i + 1 {
      arraymut = append(arraymut, NewSkipMutation(len(self.keys) - i - 1))
    }
    datamut := NewObjectMutation()
    datamut["items"] = NewArrayMutation(arraymut)
    mut["_data"] = datamut    
    self.revision++
    self.send(mut)
    return
  }
}

func (self *View) send(mut map[string]interface{}) {
  jsonmut, err := json.Marshal(mut)
  if err != nil {
    panic("Cannot serializa JSON")
  }
  log.Println("VIEW", "/_view/" + self.id, string(jsonmut))
  self.session.Enqueue( &UpdateMsg{URI:"/_view/" + self.id, Mutation: []string{string(jsonmut)}} )
}

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
// Requests

type SessionRequest struct {
  Response http.ResponseWriter
  // Send to this channel if the request has been completed.
  // Send true to indicate success and false to indicate that an error occurred.
  FinishSignal chan bool  
}

type SessionGetRequest struct {
  SessionRequest
}

type SessionPollRequest struct {
  SessionRequest
}

type SessionOpenDocRequest struct {
  SessionRequest
  DocumentURI string
  Snapshot bool
}

type SessionCloseDocRequest struct {
  SessionRequest
  DocumentURI string
}

type SessionOpenViewRequest struct {
  SessionRequest
  ViewId string
  Mapping DocumentMappingId
  WithTags []string
  WithoutTags []string
}

type SessionCloseViewRequest struct {
  SessionRequest
  ViewId string
}

type SessionCloseMsg struct {
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
  channel chan interface{}
  pollRequest *SessionPollRequest
  // The key is the URI of a node that has sent an update.
  // The value is the document describing this update, i.e. a serialized JSON document
  queue map[string][]string
  openDocs map[string]bool
  // The key is the view ID transmitted by the client.
  // The value is the queryId that has been passed to the indexer.
  openViews map[string]*View
}

func newSession(sessionDatabase *SessionDB, username string) *Session {
  s := &Session{Username:username}
  s.Id = createSessionId(username)
  s.CreationTime = time.UTC().Seconds()
  s.Cookie = s.Id
  s.sessionDatabase = sessionDatabase
  s.channel = make(chan interface{}, 10)
  s.queue = make(map[string][]string)
  s.openDocs = make(map[string]bool)
  s.openViews = make(map[string]*View)
  return s
}

func (self *Session) Enqueue( msg interface{} ) {
  self.channel <- msg
}

func (self *Session) Run() {
  for {
    select {
    case msg := <-self.channel:
      switch msg.(type) {
      case *UpdateMsg:
	self.update(msg.(*UpdateMsg))
      case *SessionGetRequest:
	self.get(msg.(*SessionGetRequest))
      case *SessionPollRequest:
	self.poll(msg.(*SessionPollRequest))
      case *SessionOpenDocRequest:
	self.openDoc(msg.(*SessionOpenDocRequest))
      case *SessionCloseDocRequest:
	self.closeDoc(msg.(*SessionCloseDocRequest))
      case *SessionOpenViewRequest:
	self.openView(msg.(*SessionOpenViewRequest))
      case *SessionCloseViewRequest:
	self.closeView(msg.(*SessionCloseViewRequest))
      case *SessionCloseMsg:
	self.closeSession()
	return
      default:
	panic("Unknown message type received in session")
      }
    }
  }
}

func (self *Session) openView(msg *SessionOpenViewRequest) {
  msg.Response.SetHeader("Content-Type", "application/json")

  if _, ok := self.openViews[msg.ViewId]; ok {
    _, err := msg.Response.Write( []byte("{\"ok\":false, \"error\":\"View ID is already in use.\"}") )
    if err != nil {
      log.Println("Failed writing HTTP response")
      msg.FinishSignal <- false
    }
    msg.FinishSignal <- true
    return
  }
  view := NewView(self, msg.ViewId, msg.Mapping, msg.WithTags, msg.WithoutTags)
  self.openViews[msg.ViewId] = view  

  _, err := msg.Response.Write( []byte("{\"ok\":true}") )
  if err != nil {
    log.Println("Failed writing HTTP response")
    msg.FinishSignal <- false
  }
  msg.FinishSignal <- true
}

func (self *Session) closeView(msg *SessionCloseViewRequest) {
  msg.Response.SetHeader("Content-Type", "application/json")

  view, ok := self.openViews[msg.ViewId]
  if !ok {
    _, err := msg.Response.Write( []byte("{\"ok\":false, \"error\":\"Unknown view ID\"}") )
    if err != nil {
      log.Println("Failed writing HTTP response")
      msg.FinishSignal <- false
    }
    msg.FinishSignal <- true
    return
  }
  
  view.stop()
  self.openViews[msg.ViewId] = nil, false
  
  _, err := msg.Response.Write( []byte("{\"ok\":true}") )
  if err != nil {
    log.Println("Failed writing HTTP response")
    msg.FinishSignal <- false
  }
  msg.FinishSignal <- true
}

func (self *Session) openDoc(msg *SessionOpenDocRequest) {
  success := true
  defer func() {
    msg.Response.SetHeader("Content-Type", "application/json")
    _, err := msg.Response.Write( []byte(fmt.Sprintf("{\"ok\":%v}", success)) )
    if err != nil {
      log.Println("Failed writing HTTP response")
      msg.FinishSignal <- false
    }
    msg.FinishSignal <- true
  } ()
  if _, ok := self.openDocs[msg.DocumentURI]; !ok {
    response := make(chan bool)
    self.sessionDatabase.server.PubSub( &PubSubRequest{Action:Subscribe, DocumentURI:msg.DocumentURI, Subscriber:self, FinishSignal:response, Snapshot:msg.Snapshot, Id:self.Id} )
    // Wait for the result
    success = <-response
    if success {
      self.openDocs[msg.DocumentURI] = true
    }
  }
}

func (self *Session) closeDoc(msg *SessionCloseDocRequest) {
  success := true
  defer func() {
    msg.Response.SetHeader("Content-Type", "application/json")
    _, err := msg.Response.Write( []byte(fmt.Sprintf("{\"ok\":%v}", success)) )
    if err != nil {
      log.Println("Failed writing HTTP response")
      msg.FinishSignal <- false
    }
    msg.FinishSignal <- true
  } ()
  if _, ok := self.openDocs[msg.DocumentURI]; !ok {
    success = false
  } else {
    self.sessionDatabase.server.PubSub( &PubSubRequest{Action:Unsubscribe, DocumentURI:msg.DocumentURI, Subscriber:self, FinishSignal:nil, Id:self.Id} )
    self.openDocs[msg.DocumentURI] = false, false
  }  
}

func (self *Session) closeSession() {
  for uri, _ := range self.openDocs {
    self.sessionDatabase.server.PubSub( &PubSubRequest{Action:Unsubscribe, DocumentURI:uri, Subscriber:self, FinishSignal:nil} )
  }
  if self.pollRequest != nil {
    self.pollRequest.Response.SetHeader("Content-Type", "application/json")
    _, err := self.pollRequest.Response.Write( []byte("{\"ok\":false, \"error\":\"Session has been closed\"}") )
    if err != nil {
      log.Println("Failed writing HTTP response")
      self.pollRequest.FinishSignal <- false
    }
    self.pollRequest.FinishSignal <- true
  }
}

// For Subscriber interface
//func (self *Session) Update(msg *UpdateMsg) {
//  self.channel <- msg
//}

func (self *Session) update(msg *UpdateMsg) {
  log.Println("Update for session ", self.Id, " from URI ", msg.URI)
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
  if self.pollRequest != nil {
    self.pollRequest.Response.SetHeader("Content-Type", "application/json")
    _, err := self.pollRequest.Response.Write( self.marshalQueue() )
    if err != nil {
      self.pollRequest = nil;
      log.Println("Failed writing HTTP response")
      self.pollRequest.FinishSignal <- false
      return
    }
    self.queue = nil
    self.pollRequest.FinishSignal <- false
    self.pollRequest = nil
  }
}

func (self *Session) marshalQueue() []byte {
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

func (self *Session) get(req *SessionGetRequest) {
  req.Response.SetHeader("Content-Type", "application/json")
  _, err := req.Response.Write( self.marshalQueue() )
  if err != nil {
    log.Println("Failed writing HTTP response")
    req.FinishSignal <- false
    return
  }
  self.queue = nil
  req.FinishSignal <- true
}

func (self *Session) poll(req *SessionPollRequest) {
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
    if self.pollRequest != nil {
      self.pollRequest.Response.SetHeader("Content-Type", "application/json")
      self.pollRequest.Response.Write( []byte("{}") )
      self.pollRequest.FinishSignal <- true
    }
    self.pollRequest = req
  }
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
  lock sync.Mutex
  sessions map[string]*Session
  server *Server
}

func NewSessionDB(server *Server) *SessionDB {
  return &SessionDB{server:server, sessions:make(map[string]*Session)}
}

func (self *SessionDB) CreateSession(username string) (*Session, os.Error) {
  self.lock.Lock()
  defer self.lock.Unlock()
  // TODO avoid that one user is creating too many concurrent sessions
  s := newSession(self, username)
  self.sessions[s.Id] = s
  go s.Run()
  return s, nil
}

func (self *SessionDB) FindSession(req *http.Request) (*Session, os.Error) {
  self.lock.Lock()
  defer self.lock.Unlock()
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
  self.lock.Lock()
  defer self.lock.Unlock()
  log.Println("Deleting session ", session.Id)
  session.channel <- &SessionCloseMsg{}
  self.sessions[session.Id] = nil, false
}

func (self *SessionDB) findSession(cookie string) (*Session, os.Error) {
  session, ok := self.sessions[cookie]
  if !ok {
    return nil, os.NewError("No such session")
  }
  return session, nil
}
