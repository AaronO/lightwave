package wave

import (
  . "lightwave"
  "testing"
  "bytes"
  "os"
  "io"
  "bufio"
  "log"
  "json"
  proto "goprotobuf.googlecode.com/hg/proto"  
)

func compareJson(val1, val2 interface{}) bool {
  if obj1, ok := val1.(map[string]interface{}); ok {
	obj2, ok := val2.(map[string]interface{})
	if !ok {
	  return false
	}
	if !compareJsonObject(obj1, obj2) {
	  return false
	}
	return true
  }
  if a1, ok := val1.([]interface{}); ok {
	a2, ok := val2.([]interface{})
	if !ok {
	  return false
	}	  
	if !compareJsonArray(a1, a2) {
	  return false
	}
	return true
  }
  if s1, ok := val1.(string); ok {
	s2, ok := val2.(string)
	if !ok {
	  return false
	}
	if s1 != s2 {
	  return false
	}
	return true
  }
  if s1, ok := val1.(bool); ok {
	s2, ok := val2.(bool)
	if !ok {
	  return false
	}
	if s1 != s2 {
	  return false
	}
	return true
  }
  if s1, ok := val1.(float64); ok {
	s2, ok := val2.(float64)
	if !ok {
	  return false
	}
	if s1 != s2 {
	  return false
	}
	return true
  }
  return false
}

func compareJsonArray(arr1, arr2 []interface{}) bool {
  if len(arr1) != len(arr2) {
	return false
  }
  for i, val1 := range arr1 {
	val2 := arr2[i]
	if !compareJson(val1, val2 ) {
	  return false
	}
  }
  return true
}

func compareJsonObject(obj1, obj2 map[string]interface{}) bool {
  for key, val1 := range obj1 {
	val2, ok := obj2[key]
	if !ok {
	  return false
	}
	if !compareJson(val1, val2 ) {
	  return false
	}
  }
  for key, _ := range obj2 {
	_, ok := obj1[key]
	if !ok {
	  return false
	}
  }
  return true
}

func testJsonPost(server *Server, uri string, jsonRequest string, jsonResponse string, t *testing.T) bool {
  // Send a POST message to create a document
  u, ok := NewURI(uri)
  if !ok {
	t.Errorf("Failed to parse URI")
	return false
  }  
  ch := make(chan bool)
  w := NewResponseWriter()
  req := &PostRequest{Request{w, "", ch, ClientOrigin, u}, []byte(jsonRequest), "application/json"}
  server.Post( req )
  <- ch
  if w.StatusCode != 200 {
	t.Errorf("Expected status code 200")
	return false
  }  
  
  j1 := make(map[string]interface{})
  if err := json.Unmarshal( w.Buffer.Bytes(), &j1 ); err != nil {
    t.Errorf("Failed to parse JSON response: %v", err)
    return false
  }
  j2 := make(map[string]interface{})
  if err := json.Unmarshal( []byte(jsonResponse), &j2 ); err != nil {
    t.Errorf("Malformed JSON in test case: %v\n%v", jsonResponse, err)
    return false
  }
  if !compareJson(j1, j2) {
    t.Errorf("Expected\n\t%v\ninstead got\n\t%v\n", jsonResponse, string(w.Buffer.Bytes()))
    return false
  }
  return true
}

func testJsonGet(server *Server, uri string, jsonResponse string, t *testing.T) bool {
  // Send a POST message to create a document
  u, ok := NewURI(uri)
  if !ok {
	t.Errorf("Failed to parse URI")
	return false
  }  
  ch := make(chan bool)
  w := NewResponseWriter()
  req := &GetRequest{Request{w, "", ch, ClientOrigin, u}}
  server.Get( req )
  <- ch	
  if w.StatusCode != 200 {
	t.Errorf("Expected status code 200")
	return false
  }  
  
  j1 := make(map[string]interface{})
  if err := json.Unmarshal( w.Buffer.Bytes(), &j1 ); err != nil {
    t.Errorf("Failed to parse JSON response: %v", err)
    return false
  }
  j2 := make(map[string]interface{})
  if err := json.Unmarshal( []byte(jsonResponse), &j2 ); err != nil {
    t.Errorf("Malformed JSON in test case: %v\n", jsonResponse, err)
    return false
  }
  if !compareJson(j1, j2) {
    t.Errorf("Expected\n\t%v\ninstead got\n\t%v\n", jsonResponse, string(w.Buffer.Bytes()))
    return false
  }
  return true
}

func testPost(server *Server, uri string, jsonRequest []byte, t *testing.T) ([]byte, bool) {
  // Send a POST message to create a document
  u, ok := NewURI(uri)
  if !ok {
	t.Errorf("Failed to parse URI")
	return nil, false
  }  
  ch := make(chan bool)
  w := NewResponseWriter()
  req := &PostRequest{Request{w, "", ch, ClientOrigin, u}, jsonRequest, "application/json-wave"}
  server.Post( req )
  <- ch
  if w.StatusCode != 200 {
	t.Errorf("Expected status code 200")
	return nil, false
  }  
  return w.Buffer.Bytes(), true
}

func testGet(server *Server, uri string, t *testing.T) ([]byte, bool) {
  // Send a POST message to create a document
  u, ok := NewURI(uri)
  if !ok {
	t.Errorf("Failed to parse URI")
	return nil, false
  }  
  ch := make(chan bool)
  w := NewResponseWriter()
  req := &GetRequest{Request{w, "", ch, ClientOrigin, u}}
  server.Get( req )
  <- ch	
  if w.StatusCode != 200 {
	t.Errorf("Expected status code 200")
	return nil, false
  }  
  return w.Buffer.Bytes(), true
}

// ------------------------------------------------
// Fake response writer

type ResponseWriter struct {
  Header map[string]string
  Buffer bytes.Buffer
  StatusCode int
}

func NewResponseWriter() *ResponseWriter {
  r := &ResponseWriter{}
  r.Header = make(map[string]string)
  r.StatusCode = 200
  return r
}

func (self *ResponseWriter) RemoteAddr() string {
  return "remote"
}

func (self *ResponseWriter) UsingTLS() bool {
  return false
}

func (self *ResponseWriter) SetHeader(key string, value string) {
  self.Header[key] = value
}

func (self *ResponseWriter) WriteHeader(code int) {
  self.StatusCode = code
}

func (self *ResponseWriter) Write(data []byte) (int, os.Error) {
  self.Buffer.Write(data)
  return len(data), nil
}

func (self *ResponseWriter) Flush() {
  // Do nothing by intention
}

func (self *ResponseWriter) Hijack() (io.ReadWriteCloser, *bufio.ReadWriter, os.Error) {
  panic("Hijack is not implemented")
  return nil, nil, nil
}

func createSubmit(version int64, hash []byte, doc string, docop *ProtocolDocumentOperation) []byte {
  submit := &ClientSubmitRequest{}
  delta := &ProtocolWaveletDelta{}
  delta.Author = proto.String("torben")
  delta.HashedVersion = &ProtocolHashedVersion{Version:proto.Int64(version), HistoryHash: hash}
  
  mut := &ProtocolWaveletOperation_MutateDocument{DocumentId:proto.String(doc), DocumentOperation:docop}
  op := &ProtocolWaveletOperation{MutateDocument:mut}
  
  delta.Operation = [...]*ProtocolWaveletOperation{op}[:]
  submit.Delta = delta
  
  var buffer bytes.Buffer
  Marshal(submit, &buffer)
  log.Println( buffer.String() )
  return buffer.Bytes()
}

func checkResponse(t *testing.T, response []byte, version int64, ops_applied int32) bool {
  pr := &ClientSubmitResponse{}
  err := Unmarshal(response, pr)
  if err != nil {
	t.Errorf("Result is no json protobuf: %v", string(response))
	return false
  }
  
  if pr.ErrorMessage != nil {
	t.Errorf("Error from wave server: " + *pr.ErrorMessage)
	return false
  }
  if *pr.OperationsApplied != ops_applied {
	t.Errorf("Not all operations have been applied")
	return false
  }
  if *pr.HashedVersionAfterApplication.Version != version {
	t.Errorf("Version is wrong")
	return false
  }  
  return true
}

func TestSubmit(t *testing.T) {
  // Create a root node
  server := NewServer(&ServerManifest{Domain:"localhost", HostName:"localhost", Port:8080})
  RegisterNodeFactory("application/x-protobuf-wave", NewWaveletNode)
  RegisterNodeFactory("application/json-wave", NewWaveletNode)
  RegisterNodeFactory("application/json", NewDocumentNode)
  go server.Run()

  // Stop the root node
  defer server.Stop()

  // Create a mutation
  docop := &ProtocolDocumentOperation{}
  docop.Component = [...]*ProtocolDocumentOperation_Component{ opElementStart("p"), opCharacters("Hallo Welt"), opElementEnd() }[:]  
  msg := createSubmit(0, []byte("wave://localhost/w+abc/conv+root"), "b+1", docop)
  r, ok := testPost(server, "/localhost/localhost$w+abc$conv+root", msg, t)
  if !ok || !checkResponse( t, r, 1, 1 ) {
	t.Errorf("Post failed")
	return
  }

  // Create a session
  if !testJsonPost(server, "/_session/weis/s1", `{"_rev":0, "_data":{"filters":[{"prefix":"/localhost/localhost$w+abc$conv+root", "recursive":true, "mimeTypes":[], "schemas":[]}]}}`, `{"ok":true, "appliedAt":1}`, t) {
	return
  }
  
  // Another mutation
  docop.Component = [...]*ProtocolDocumentOperation_Component{ opRetain(1), opCharacters("Wow! "), opRetain(11) }[:]
  msg = createSubmit(1, []byte("wave://localhost/w+abc/conv+root"), "b+1", docop)
  r, ok = testPost(server, "/localhost/localhost$w+abc$conv+root", msg, t)
  if !ok || !checkResponse( t, r, 2, 1 ) {
	t.Errorf("Post failed")
	return
  }
  
  // Poll the session
  msg, ok = testGet(server, "/_session/weis/s1/_poll", t)
  if !ok {
	t.Errorf("Post failed")
	return
  }
  log.Println(string(msg))
}
