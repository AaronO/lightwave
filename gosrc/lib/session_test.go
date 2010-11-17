package lightwave

import (
  "testing"
  "bytes"
  "os"
  "io"
  "bufio"
  "json"
)

func testPost(server *Server, uri string, jsonRequest string, jsonResponse string, t *testing.T) bool {
  // Send a POST message to create a document
  u, ok := NewURI(uri)
  if !ok {
	t.Errorf("Failed to parse URI")
	return false
  }  
  ch := make(chan bool)
  w := NewResponseWriter()
  req := &PostRequest{Request{w, ch}, u, []byte(jsonRequest), "application/json"}
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

func testGet(server *Server, uri string, jsonResponse string, t *testing.T) bool {
  // Send a POST message to create a document
  u, ok := NewURI(uri)
  if !ok {
	t.Errorf("Failed to parse URI")
	return false
  }  
  ch := make(chan bool)
  w := NewResponseWriter()
  req := &GetRequest{Request{w, ch}, u}
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

func TestSession(t *testing.T) {
  // Create a root node
  server := NewServer(&ServerManifest{Domain:"localhost", HostName:"localhost", Port:8080})
  go server.Run()

  // Stop the root node
  defer server.Stop()

  if !testPost(server, "/local/foo", `{"_rev":0, "_data":{"foo":"bar"}}`, `{"ok":true, "appliedAt":1}`, t) {
	return
  }
  if !testPost(server, "/local/foo", `{"_rev":1, "_data":{"$object":true, "hoo":"gar"}}`, `{"ok":true, "appliedAt":2}`, t) {
	return
  }
  if !testGet(server, "/local/foo", `{"_rev":2, "_data":{"foo":"bar", "hoo":"gar"}, "_meta":{}}`, t) {
	return
  }
  if !testPost(server, "/_session/weis/s1", `{"_rev":0, "_data":{"filters":[{"prefix":"/local/foo", "recursive":true, "mimeTypes":[], "schemas":[]}]}}`, `{"ok":true, "appliedAt":1}`, t) {
	return
  }
//  time.Sleep(10000)
//  if !testGet(root, "/_session/weis/s1/_update", `{"/local/foo":[{"_data":{"foo":"bar","hoo":"gar"},"_meta":{},"_rev":2}]}`, t) {
  if !testGet(server, "/_session/weis/s1/_poll", `{"/local/foo":[{"_data":{"foo":"bar","hoo":"gar"},"_meta":{},"_rev":2}]}`, t) {
	return
  }
}
