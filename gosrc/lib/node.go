package lightwave

import (
  "fmt"
  "json"
  "log"
  "http"
  "strings"
  "sync"
  "strconv"
  "os"
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

// --------------------------------------------------------
// Requests

// A base class for incoming requests, i.e. HTTP POST or HTTP GET
type Request struct {
  Response http.ResponseWriter
  // The URL query string (if any)
  RawQuery string
  // Send to this channel if the request has been completed.
  // Send true to indicate success and false to indicate that an error occurred.
  FinishSignal chan bool
  // Either FederationOrigin or ClientOrigin to indicate the origin of the request.
  Origin int32
  // A unique identifier of the destination of this request, i.e. it identifies a Node
  URI URI
  SessionPtr *Session
}

const (
  FederationOrigin = iota
  ClientOrigin
)

// Call this after the request has been completed
func (self Request) SendFinishSignal(result bool) {
  self.FinishSignal <- result
}

func (self Request) GetResponseWriter() http.ResponseWriter {
  return self.Response
}

type PostRequest struct {
  Request
  Data []byte
  MimeType string
}

type GetRequest struct {
  Request
}

const (
  Subscribe = iota
  Unsubscribe
  SubscribeIndexer
  UnsubscribeIndexer
)

type PubSubRequest struct {
  // Id of the subscription. Required to remove a subscription later on.
  Id string
  // Used for Subscribe
  Snapshot bool
  // Subscribe or Unsubscribe
  Action uint32
  // Used in SubscribeIndexer
  Mapping DocumentMappingId
  DocumentURI string
  Subscriber Subscriber
  FinishSignal chan bool
}

// Update messages are sent to subscribers
type UpdateMsg struct {
  // The document URI to which this update belongs
  URI string
  // JSON encoded mutation
  Mutation []string
}

type Subscriber interface {
  Enqueue(msg interface{})
}

// -------------------------------------------
// Node interface

// All nodes in the document tree must implement this interface
type Node interface {
  Name() string
  URI() string
  Run()
  Post(req *PostRequest)
  Get(req *GetRequest)
  PubSub(req *PubSubRequest)
  Stop()
  Server() *Server
  Host() *HostNode
}

// -------------------------------------------
// NodeBase

// A partial implementation of the Node interface
type NodeBase struct {
  parent Node
  name string
  postChannel chan *PostRequest
  getChannel chan *GetRequest
  stopChannel chan bool
  pubSubChannel chan *PubSubRequest
}

func (self *NodeBase) Server() *Server {
  if self.parent == nil {
    return nil
  }
  if s, ok := self.parent.(*Server); ok {
    return s
  }
  return self.parent.Server()
}

func (self *NodeBase) Host() *HostNode {
  if self.parent == nil {
    return nil
  }
  if s, ok := self.parent.(*HostNode); ok {
    return s
  }
  return self.parent.Host()
}

func (self *NodeBase) Post(req *PostRequest) {
  self.postChannel <- req
}

func (self *NodeBase) Get(req *GetRequest) {
  self.getChannel <- req
}

func (self *NodeBase) Stop() {
  self.stopChannel <- true
}

func (self *NodeBase) PubSub(req *PubSubRequest) {
    self.pubSubChannel <- req
}

func (self *NodeBase) Name() string {
  return self.name
}

func (self *NodeBase) URI() string {
  if self.parent != nil {
    return self.parent.URI() + "/" + self.name    
  }
  return ""
}

// -------------------------------------------
// NodeFactory

type NodeFactory func(parent Node, name string, level int) Node

var factories map[string]NodeFactory = make(map[string]NodeFactory)

func RegisterNodeFactory(mimeType string, factory NodeFactory) {
  factories[mimeType] = factory;
}

func CreateNode(parent Node, name string, level int, mimeType string) (node Node, err os.Error) {
  fac, ok := factories[mimeType]
  if !ok {
    return nil, os.NewError("Unsupported mimeType: " + mimeType)
  }
  return fac(parent, name, level), nil
}

// -------------------------------------------
// DocumentNode

type DocumentNode struct {
  NodeBase
  children map[string]Node
  level int
  doc map[string]interface{}
  subscriptions map[string]Subscriber
  // List of domains which participate in federating this document
  federatedDomains map[string]bool
  history *DocumentHistory
  mappings map[DocumentMappingId]interface{}
  tags []string
}

func DocumentNodeFactory(parent Node, name string, level int) Node {
  return NewDocumentNode(parent, name, level)
}

func NewDocumentNode(parent Node, name string, level int) *DocumentNode {
  d := &DocumentNode{children:make(map[string]Node), level:level, NodeBase:NodeBase{parent:parent, name:name,postChannel:make(chan *PostRequest), getChannel:make(chan *GetRequest), stopChannel:make(chan bool), pubSubChannel:make(chan *PubSubRequest)}}
  d.subscriptions = make(map[string]Subscriber)
  d.federatedDomains = make(map[string]bool)
  d.history = NewDocumentHistory(d)
  if d.history.broken {
    log.Println("Failed to read: ", d.URI())
  }
  d.tags = []string{}
  d.mappings = make(map[DocumentMappingId]interface{})
  return d
}

func (self *DocumentNode) Run() {
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

func (self *DocumentNode) Schema() string {  
  return getString(getObject(self.doc["_meta"])["schema"])
}

func (self *DocumentNode) Participants() []*UserId {  
  meta, ok := self.doc["_meta"]
  if !ok {
    return nil;
  }
  metamap := meta.(map[string]interface{})
  particiants, ok := metamap["participants"]
  if !ok {
    return nil;
  }
  arr, ok := particiants.([]interface{})
  if !ok {
    return nil;
  }
  result := make([]*UserId, 0, len(arr))
  for _, p := range arr {
    if user, ok := p.(map[string]interface{}); ok {
      if d, ok := user["userid"]; ok {
	if s, ok := d.(string); ok {
	  u, err := NewUserId(s)
	  if err == nil {
            result = append( result, u )
	  }
	}
      }
    }
  }
  return result
}


func (self *DocumentNode) apply( mutation map[string]interface{} ) bool {
  if !(IsDocumentMutation(mutation)) {
    log.Println("Not a document mutation")
    return false
  }
  m := DocumentMutation(mutation)

  // TODO: Check that the delta has the right hash
  
  // Apply the mutation at the most recent version of the document?
  if m.AppliedAtRevision() == self.Revision() {
    if err := m.Apply(self.doc, NoFlags); err != nil {
      log.Println("Failed applying delta: ", err)
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
  // Send message to subscribers
  for _, s := range self.subscriptions {
    jsonmsg, _ := json.Marshal(m)
    s.Enqueue( &UpdateMsg{self.URI(), []string{string(jsonmsg)}})
  }
  // At this point we know that the mutation is applied.
  
  // Compute tags and digest for the indexer
  server := self.Server()
  self.tags = GetTags(self)
  for mapping, _ := range self.mappings {
    self.mappings[mapping] = self.view(mapping)
  }
  if len(self.mappings) > 0 || len(self.tags) > 0 {
    server.Indexer.Put( self.URI(), self.mappings, self.tags )
  }
  
  // If this is just a remote server for this document, we are done
  if !self.Host().IsLocal() {
    return true
  }

  // Find out whether there are new domains involved. In this case we
  // have to send the mutation as a wavelet update to all the others
  // This code assumes that the meta data can be arbitrarily malformed. This is perhaps overly defensive
  users := self.Participants()
  self.federatedDomains = make(map[string]bool)  
  for _, u := range users {
    // If this is a remote user, we must federate
    if u.Domain != server.manifest.Domain {
      self.federatedDomains[u.Domain] = true
    }
  }
  
  // Forward the mutation to all federated servers
  if len(self.federatedDomains) > 0 {
    msg, err := json.Marshal(mutation)
    if err != nil {
      panic("Cannot encode my own data")
    }
    for domain, _ := range self.federatedDomains {
      server.Gateway(domain).WaveletUpdate(self.URI(), msg)
    }
  }
  
  return true
}

func (self *DocumentNode) Revision() int64 {
  return int64(self.doc["_rev"].(float64))
}

func (self *DocumentNode) addChild(child Node) {
  self.children[child.Name()] = child
  // Add meta data1
  mut := make(map[string]interface{})
  mut["_rev"] = float64(self.Revision())
  var childmut interface{}
  childlist, ok := self.doc["_meta"].(map[string]interface{})["children"]
  if !ok {
    childmut = []interface{}{ child.Name() }    
  } else {
    skip := make(map[string]interface{})
    skip["$skip"] = float64(len(childlist.([]interface{})))
    m := make(map[string]interface{})
    m["$array"] = []interface{}{ skip, child.Name() }
    childmut = m
  }
  metamut := make(map[string]interface{})
  metamut["$object"] = true
  metamut["children"] = childmut
  mut["_meta"] = metamut
  ok = self.apply(mut)
  if !ok {
    panic("Failed to apply mutation for meta data of " + self.Name())
  }
}
 
func (self *DocumentNode) post(req *PostRequest) {  
  docuri := req.URI.(DocumentURI)
  
  // Request is aimed at this document?
  if len(docuri.NameSeq) == self.level {
    log.Println("Document is putting itself: ", req.URI)
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
        // if _, ok := m["_meta"]; ok {
        //  makeErrorResponse(req.Response, "Attempt to modify meta data using POST")
        //  req.FinishSignal <- false
        //  return
        // }
        // Try to apply the data
        if !self.apply(m) {
          makeErrorResponse(req.Response, "Could not apply document mutation" + string(req.Data))
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
        makeErrorResponse(req.Response, "Data type not allowed for put")
        req.FinishSignal <- false
    }
    return
  }
  
  // Request is aimed at a child document
  doc, ok := self.children[docuri.NameSeq[self.level]]
  if !ok {
    doc = self.loadDocument(docuri.NameSeq[self.level])
  }
  if doc != nil {
    doc.Post(req)
    return
  }
  // Create a document?
  if len(docuri.NameSeq) != self.level + 1 {
    makeErrorResponse(req.Response, "Cannot create an inner node and a child node in one request")
    req.FinishSignal <- false
    return
  }
  // Try to initialize the document. Discard it if initialization fails
  doc = NewDocumentNode(self, docuri.NameSeq[self.level], self.level + 1)
  go doc.Run()
  oldfinish := req.FinishSignal
  finish := make(chan bool)
  req.FinishSignal = finish
  doc.Post(req) 
  ok = <- finish
  if ok {
    self.addChild(doc)
  } else {
    doc.Stop()
  }
  oldfinish <- ok
}

func (self *DocumentNode) get(req *GetRequest) {
  docuri := req.URI.(DocumentURI)
  
  // Request is aimed at this document?
  if len(docuri.NameSeq) == self.level {
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
      result, err := self.history.Range(int64(v1), v1hash[0], int64(v2), v2hash[0], int64(limit))
      if err != nil {
        log.Println("Failed retrieving history ", err)
        makeErrorResponse(req.Response, "Failed retrieving history")
        req.FinishSignal <- false
        return
      }
      // Send the result by HTTP
      req.Response.SetHeader("Content-Type", "application/json")
      _, err = req.Response.Write( []byte(result) )
      if err != nil {
        log.Println("Failed writing HTTP response")
        req.FinishSignal <- false
        return
      }
      req.FinishSignal <- true
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
    fmt.Println("Document is getting itself with req %v", req)
    req.FinishSignal <- true
    return
  }
  
  // Request is aimed at a child document
  doc, ok := self.children[docuri.NameSeq[self.level]]
  if !ok {
    doc = self.loadDocument(docuri.NameSeq[self.level])
  }
  if doc == nil {
    makeErrorResponse( req.Response, "Document " + docuri.NameSeq[self.level] + " does not exist" )
    req.FinishSignal <- false
    return
  }
  doc.Get(req)
}

func (self *DocumentNode) pubSub(req *PubSubRequest) {
  uri := self.URI()
  seq := strings.Split(req.DocumentURI[1:], "/", -1)
  if len(seq) < 1 + self.level {
    panic("The subscription must not reach this node")
    return
  }
  if len(seq) == 1 + self.level && uri != req.DocumentURI {
    panic("The subscription must not reach this node")
    return
  }
    
  if uri == req.DocumentURI {
    // Subscribe/Unsubscribe
    switch req.Action {
    case Subscribe:
      log.Println("Subscribing node ", uri)
      if req.FinishSignal != nil {
	req.FinishSignal <- true
      }
      self.subscriptions[req.Id] = req.Subscriber
      if req.Snapshot {
        // Send a snapshot to the subscriber
        clone := cloneJsonObject(self.doc)
        clone["_endRev"] = clone["_rev"]
        // clone["_endHash"] = clone["_hash"]
        clone["_rev"] = 0
        // clone["_hash"] = "TODOHASH"
        clone["_data"].(map[string]interface{})["$object"] = true;
        clone["_meta"].(map[string]interface{})["$object"] = true;
        m, _ := json.Marshal(clone)
        req.Subscriber.Enqueue( &UpdateMsg{uri, []string{string(m)}} )
      } else {
        // Send delta history to subscriber
        lst := make([]string, self.Revision())
        tail := self.history.Tail(0)
        for i, d := range tail {
          m, _ := json.Marshal(d)
          lst[i] = string(m)
        }
        req.Subscriber.Enqueue( &UpdateMsg{uri, lst} )
      }
    case Unsubscribe:
      log.Println("Unsubscribing node ", uri)
      self.subscriptions[req.Id] = nil, false
      if req.FinishSignal != nil {
	req.FinishSignal <- true
      }
    case SubscribeIndexer:
      self.mappings[req.Mapping] = self.view(req.Mapping)
      self.Server().Indexer.Put(req.DocumentURI, self.mappings, self.tags)
    case UnsubscribeIndexer:
      self.mappings[req.Mapping] = "", false
    default:
      panic("Unsupported PubSub action")
    }
  } else {
    node, ok := self.children[seq[self.level + 1]]
    if !ok {
      node = self.loadDocument(seq[self.level + 1])
    }
    if node == nil {
      log.Println("Document ", seq[self.level + 1], " does not exist")
      return
    }
    node.PubSub(req)  
  }    
}

func (self *DocumentNode) view(mapping DocumentMappingId) string {
  result, _ := json.Marshal(GetDigest(self, mapping))
  return string(result)
}

func (self *DocumentNode) loadDocument(name string) Node {
  if !isDocumentPersisted(self.Server(), self.URI() + "/" + name) {
    return nil
  }
  doc := NewDocumentNode(self, name, self.level + 1)
  go doc.Run()
  self.addChild(doc)
  return doc
}

// -----------------------------------------------
// HostNode

type HostNode struct {
  NodeBase
  children map[string]Node
  proxy *FederationProxy
  users *UserRootNode
}

func NewHostNode(parent Node, host string) *HostNode {
  h := &HostNode{children:make(map[string]Node), NodeBase:NodeBase{parent:parent, name:host, postChannel:make(chan *PostRequest), getChannel:make(chan *GetRequest), stopChannel:make(chan bool), pubSubChannel:make(chan *PubSubRequest)}}
  h.users = NewUserRootNode(h)
  go h.users.Run()
  h.children[h.users.Name()] = h.users
  return h
}

func (self *HostNode) Run() {
  if !self.IsLocal() {
    self.proxy = NewFederationProxy(self.name)
    go self.proxy.Run()
  }
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

func (self *HostNode) Users() *UserRootNode {
  return self.users
}

func (self *HostNode) IsLocal() bool {
  return self.parent.(*Server).manifest.Domain == self.name
}

func (self *HostNode) addChild(child Node) {
  self.children[child.Name()] = child
}

func (self *HostNode) post(req *PostRequest) {
  // Send it via the federation host, i.e. this is only the remote server for this document??
  if !self.IsLocal() && req.Origin == ClientOrigin {
    self.proxy.Submit(req)
    return
  }
  
  docuri := req.URI.(DocumentURI)
  // A document of the desired name exists?
  existing_doc, ok := self.children[docuri.NameSeq[0]]
  if ok {
    existing_doc.Post(req)
    return
  }
  // Create a document? In this case the destination document must be a direct
  // child of the current node.
  if len(docuri.NameSeq) != 1 {
    makeErrorResponse(req.Response, "ERROR" )
    req.FinishSignal <- false
    return
  }
  
  doc, err := CreateNode(self, docuri.NameSeq[0], 1, req.MimeType)
  if err != nil {
    makeErrorResponse(req.Response, err.String())
    req.FinishSignal <- false
    return
  }  

  oldfinish := req.FinishSignal
  finish := make(chan bool)
  req.FinishSignal = finish

  go doc.Run()
  doc.Post(req)
  ok = <-finish
  if ok {
    self.addChild(doc)
  } else {
    doc.Stop()
  }
  oldfinish <- ok
}

func (self *HostNode) get(req *GetRequest) {
  docuri := req.URI.(DocumentURI)
  // Does a document of the desired name exist?
  doc, ok := self.children[docuri.NameSeq[0]]
  // Document does not exist? -> Error
  if !ok {
    doc = self.loadDocument(docuri.NameSeq[0])
  }
  if doc == nil {
    makeErrorResponse(req.Response, "Document does not exist")
    req.FinishSignal <- false
    return
  }
  // Forward the request to the document
  doc.Get(req)
}

func (self *HostNode) pubSub(req* PubSubRequest) {
  seq := strings.Split(req.DocumentURI[1:], "/", -1)
  if len(seq) < 2 {
    log.Println("Malformed NodeFilter prefix")
    return
  }
  node, ok := self.children[seq[1]]
  if !ok {
    node = self.loadDocument(seq[1])
  }
  if node == nil {
    log.Println("Document ", seq[1], " does not exist")
    return
  }
  node.PubSub(req)  
}

func (self *HostNode) loadDocument(name string) Node {
  if !isDocumentPersisted(self.Server(), "/" + self.name + "/" + name) {
    return nil
  }
  doc, err := CreateNode(self, name, 1, "application/json")
  if err != nil {
    return nil
  }
  go doc.Run()
  self.addChild(doc)
  return doc
}

// --------------------------------------------------------------
// ServerManifest

type ServerManifest struct {
  Domain string "domain"
  ProtocolVersions []int32 "protocolVersions"
  Features map[string]interface{} "features"
  HostName string "host"
  Port uint16 "port"
}

// --------------------------------------------------------------
// Server

type Server struct {
  NodeBase
  Config *ServerConfig
  manifest *ServerManifest
  children map[string]Node
  SessionDatabase *SessionDB
  UserAccountDatabase *UserAccountDB
  Indexer *MemoryIndexer
  gateways map[string]*FederationGateway
  gatewaysMutex sync.Mutex
}

func NewServer(config *ServerConfig) *Server {
  r := &Server{Config:config, children:make(map[string]Node), NodeBase:NodeBase{parent:nil, name:config.Domain, postChannel:make(chan *PostRequest), getChannel:make(chan *GetRequest), stopChannel:make(chan bool), pubSubChannel:make(chan *PubSubRequest)}}
  r.manifest = &ServerManifest{Domain:config.Domain, Port:config.MainConfig.Port, HostName:config.Hostname};
  r.gateways = make(map[string]*FederationGateway)
  r.SessionDatabase = NewSessionDB(r)
  r.UserAccountDatabase = NewUserAccountDB(r)
  di := NewDiskIndexer(config.IndexDB)
  r.Indexer = NewMemoryIndexer(r, di)
  go r.Indexer.Run()
  return r
}

func (self *Server) Capabilities() *ServerManifest {
  return self.manifest
}

func (self *Server) Gateway(domain string) *FederationGateway {
  self.gatewaysMutex.Lock()
  defer self.gatewaysMutex.Unlock()
  if g, ok := self.gateways[domain]; ok {
    return g
  }
  g := NewFederationGateway(self, domain)
  go g.Run()
  self.gateways[domain] = g;
  return g;
}

func (self *Server) LocalHost() *HostNode {
  h, ok := self.children[self.manifest.Domain].(*HostNode)
  if ok {
    return h
  }
  h = NewHostNode(self, self.manifest.Domain)
  go h.Run()
  self.AddChild(h)
  return h
}

func (self *Server) Run() {
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

// The child being added must already feature an active run method, i.e. it
// is expected to handle events
func (self *Server) AddChild(child Node) {
  self.children[child.Name()] = child
}

// TODO: The root should not create host nodes upon everybody's request
func (self *Server) post(req *PostRequest) {
  switch req.URI.(type) {
    case DocumentURI:
      docuri := req.URI.(DocumentURI)
      // Check for the host node
      h, ok := self.children[docuri.Host]
      if ok {
        h.Post(req)
        return
      }
      h = NewHostNode(self, docuri.Host)
      oldfinish := req.FinishSignal
      finish := make(chan bool)
      req.FinishSignal = finish
      go h.Run()
      h.Post(req)
      ok = <-finish
      if ok {
        self.AddChild(h)
      } else {
        h.Stop()
      }
      oldfinish <- ok
    case ViewURI:
      panic("TODO")
    case ManifestURI:
      makeErrorResponse(req.Response, "Posting to a manifest is not allowed")
      req.FinishSignal <- false
      return
  }
}

func (self *Server) get(req *GetRequest) {
  switch req.URI.(type) {
    case DocumentURI:
      docuri := req.URI.(DocumentURI)
      // Check for the host node
      h, ok := self.children[docuri.Host]
      if !ok {
        h = self.loadHost(docuri.Host)
      }
      if h == nil {
        makeErrorResponse(req.Response, "No documents hosted from this server")
        req.FinishSignal <- false
        return
      }
      h.Get(req)
      return
    case StaticURI:
      _, ok := req.URI.(*StaticURI)
      // Check for the static node
      n, ok := self.children["_static"]
      if !ok {
        makeErrorResponse(req.Response, "No static documents on this server")
        req.FinishSignal <- false
        return
      }
      n.Get(req)
      return
    case ViewURI:
      panic("TODO")
    case ManifestURI:
      log.Println("Asking for manifest")
      json, err := json.Marshal(self.manifest)
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
      return
    default:
      makeErrorResponse(req.Response, "Unsupported URI type")
      req.FinishSignal <- false
      return      
  }
}

func (self *Server) pubSub(req* PubSubRequest) {
  seq := strings.Split(req.DocumentURI[1:], "/", -1)
  if len(seq) < 1 {
    log.Println("Malformed URI")
    return
  }
  node, ok := self.children[seq[0]]
  if !ok {
    host := self.loadHost(seq[0])
    if host == nil {
      log.Println("Host ", seq[0], " does not exist")
      return
    }
    node = host
  }
  node.PubSub(req)
}

func (self *Server) loadHost(hostName string) *HostNode {
  if !isHostPersisted(self, "/" + hostName) {
    return nil
  }
  h := NewHostNode(self, hostName)
  go h.Run()
  self.AddChild(h)
  return h
}

func (self *Server) Map(uri string, mapping DocumentMappingId) {
  self.pubSub( &PubSubRequest{DocumentURI:uri, Mapping:mapping, Action:SubscribeIndexer} )
}