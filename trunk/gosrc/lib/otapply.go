package lightwave

import (
  vec "container/vector"
  "log"
  "fmt"
)

const (
  NoFlags uint32 = 0
  CreateIDs uint32 = 1<<iota
)

var idcounter uint64 = 0
func uniqueId() string {
  idcounter++
  return fmt.Sprintf("id%d", idcounter)
}

// ---------------------------------------------------------
// Helper functions

func cloneJsonArray(arr []interface{}) []interface{} {
  a := make([]interface{}, len(arr))
  for i, val := range arr {
	if obj, ok := val.(map[string]interface{}); ok {
	  a[i] = cloneJsonObject(obj)
	} else if arr, ok := val.([]interface{}); ok {
	  a[i] = cloneJsonArray(arr)
	} else {
	  a[i] = val
	}	
  }
  return a
}

func cloneJsonObject(obj map[string]interface{}) map[string]interface{} {
  m := make(map[string]interface{})
  for key, val := range obj {
	if obj, ok := val.(map[string]interface{}); ok {
	  m[key] = cloneJsonObject(obj)
	} else if arr, ok := val.([]interface{}); ok {
	  m[key] = cloneJsonArray(arr)
	} else {
	  m[key] = val
	}
  }
  return m
}

// ---------------------------------------------------------
// Document Mutation

// A single mutation of a document
type DocumentMutation map[string]interface{}

func (self DocumentMutation) DataMutation() (result interface{}, ok bool) {
  m, ok := self["_data"]
  return m, ok
}

func (self DocumentMutation) MetaMutation() (result interface{}, ok bool) {
  m, ok := self["_meta"]
  return m, ok
}

// Creates a deep copy. This is required because OT changes a DocumentMutation
func (self DocumentMutation) Clone() DocumentMutation {
  return DocumentMutation( cloneJsonObject(self) )
}

func (self DocumentMutation) AppliedAtRevision() int64 {
  return int64(self["_rev"].(float64))
}

// Base64 encoded string
func (self DocumentMutation) AppliedAtHash() string {
  return self["_hash"].(string)
}

func (self DocumentMutation) ResultingRevision() int64 {
  return int64(self["_endRev"].(float64))
}

// Base64 encoded string
func (self DocumentMutation) ResultingHash() string {
  return self["_endHash"].(string)
}

// Applies this document mutation to an object.
// Returns true on success.
// If the application fails then the object remains unchanged.
func (self DocumentMutation) Apply(obj map[string]interface{}, flags uint32) bool {
  data, okdata := self.DataMutation()
  if okdata {
	if !self.Check(obj["_data"], data) {
	  return false
	}
  }
  meta, okmeta := self.MetaMutation()
  if okmeta {
	if !self.Check(obj["_meta"], meta) {
	  return false
	}
  }
  if okdata {
	obj["_data"] = self.apply(obj["_data"], data, flags)
  }
  if okmeta {
	obj["_meta"] = self.apply(obj["_meta"], meta, flags)
  }
  return true
}

func (self DocumentMutation) Check(val interface{}, mutation interface{} ) bool {
  switch {
	case IsObjectMutation(mutation):
	  obj, test := val.(map[string]interface{})
	  if !test {
		log.Println("Expected object")
		return false
	  }	  
	  return self.checkObjectMutation(obj, toObjectMutation(mutation))
	case IsArrayMutation(mutation):
	  a, test := val.([]interface{})
	  if !test {
		log.Println("Expected array")
		return false
	  }
	  return self.checkArrayMutation( a, toArrayMutation(mutation) )
	case IsTextMutation(mutation):
	  s, test := val.(string)
	  if !test {
		log.Println("Expected string")
		return false
	  }	  
	  return self.checkTextMutation( s, toTextMutation(mutation) )
	case IsInsertMutation(mutation):
	  if !self.checkInsertMutation(mutation) {
		return false
	  }	  
	  return true
	default:
	  log.Println("Unknown or unexpected mutation: ", mutation)
	  return false
  }
  return false
}

func (self DocumentMutation) checkInsertMutation(mutation interface{}) bool {
  if obj, ok := mutation.(map[string]interface{}); ok {
	if !self.checkInsertObjectMutation(obj) {
	  return false
	}
  }
  if arr, ok := mutation.([]interface{}); ok {
	if !self.checkInsertArrayMutation(arr) {
	  return false
	}
  }
  return true
}

func (self DocumentMutation) checkInsertObjectMutation(obj map[string]interface{}) bool {
  for key, val := range obj {
	if key[0] == '$' {
	  log.Println("$ is not allowed in attribute names")
	  return false
	}
	if !self.checkInsertMutation(val) {
	  return false
	}
  }
  return true
}

func (self DocumentMutation) checkInsertArrayMutation(arr []interface{}) bool {
  for _, val := range arr {
	if !self.checkInsertMutation(val) {
	  return false
	}
  }
  return true
}

func (self DocumentMutation) checkObjectMutation(obj map[string]interface{}, mutation ObjectMutation ) bool {
  for key, mut := range mutation {
	if key[0] == '$' {
	  continue
	}
	if IsInsertMutation(mut) {
	  if !self.checkInsertMutation(mut) {
		return false
	  }
	  continue
	}
	val, ok := obj[key]
	if !ok {
	  log.Println("Object does not contain key ", key)
	  return false
	}
	if !self.Check(val, mut) {
	  return false
	}
  }
  return true
}

func (self DocumentMutation) checkArrayMutation(arr []interface{}, mutation ArrayMutation ) bool {
  squeezes := make(map[string]bool)
  lifts := make(map[string]bool)
  
  index := 0
  for _, mut := range mutation.Array() {
	if IsInsertMutation(mut) {
	  if !self.checkInsertMutation(mut) {
		return false
	  }
	  continue
	}
	if IsSqueezeMutation(mut) {
	  s := toSqueezeMutation(mut)
	  if _, ok := squeezes[s.Id()]; ok {
		return false
	  }
	  squeezes[s.Id()] = true
	  continue
	}
	
	if index >= len(arr) {
	  log.Println("Array mutation is too long")
	  return false
	}
	
	switch {
	  case IsDeleteMutation(mut):
		count := toDeleteMutation(mut).Count()
		if count <= 0 {
		  log.Println("Delete <0 not allowed")
		  return false
		}
		index += count
	  case IsSkipMutation(mut):
		count := toSkipMutation(mut).Count()
		if count <= 0 {
		  log.Println("Skip <0 not allowed")
		  return false
		}
		index += count
	  case IsObjectMutation(mut):
		val, ok := arr[index].(map[string]interface{})
		if !ok {
		  log.Println("Expected object")
		  return false
		}
		self.checkObjectMutation( val, toObjectMutation(mut))
		index++
	  case IsArrayMutation(mut):
		val, ok := arr[index].([]interface{})
		if !ok {
		  log.Println("Expected array")
		  return false
		}
		self.checkArrayMutation( val, toArrayMutation(mut))
		index++
	  case IsTextMutation(mut):
		val, ok := arr[index].(string)
		if !ok {
		   log.Println("Expected string")
		  return false
		}
		self.checkTextMutation( val, toTextMutation(mut))
		index++
	  case IsLiftMutation(mut):
		l := toLiftMutation(mut)
		if l.HasMutation() && !self.Check(arr[index], l.Mutation()) {
		  return false
		}
		if _, ok := lifts[l.Id()]; ok {
		  return false
		}
		lifts[l.Id()] = true
		index++
	  default:
		 log.Println("Unknown or unexpected mutation: ", mut)
		return false
	}
  }

  if index < len(arr) {
	log.Println("Array mutation is too small")
	return false
  }
  if len(lifts) != len(squeezes) {
	log.Println("Not every lift has a squeeze or the other way round")
	return false
  }
  for id, _ := range lifts {
	if _, ok := squeezes[id]; !ok {
	  log.Println("No squeeze for ", id)
	  return false
	}
  }
  return true
}

func (self DocumentMutation) checkTextMutation(str string, mutation TextMutation ) bool {
  index := 0
  for _, mut := range mutation.Array() {
	if _, ok := mut.(string); ok {
	  continue
	}
	if index >= len(str) {
	  return false
	}
	switch {
	  case IsDeleteMutation(mut):
		count := toDeleteMutation(mut).Count()
		if count <= 0 {
		  return false
		}
		index += count
	  case IsSkipMutation(mut):
		count := toSkipMutation(mut).Count()
		if count <= 0 {
		  return false
		}
		index += count
	  default:
		return false
	}
  }
  if index < len(str) {
	return false
  }
  return true
}

func (self DocumentMutation) apply(val interface{}, mutation interface{}, flags uint32 ) interface{} {
  switch {
	case IsObjectMutation(mutation):
	  o, test := val.(map[string]interface{})
	  if !test {
		panic("That should have been caught before")
	  }  
	  self.applyObjectMutation(o, toObjectMutation(mutation), flags)
	  return o
	case IsArrayMutation(mutation):
	  a, test := val.([]interface{})
	  if !test {
		panic("That should have been caught before")
	  }
	  return self.applyArrayMutation( a, toArrayMutation(mutation), flags )
	case IsTextMutation(mutation):
	  s, test := val.(string)
	  if !test {
		panic("That should have been caught before")
	  }	  
	  return self.applyTextMutation( s, toTextMutation(mutation), flags )
	case IsInsertMutation(mutation):
	  return self.applyInsertMutation(mutation, flags)
	default:
	  return false
  }
  return false
}

func (self DocumentMutation) applyObjectMutation(obj map[string]interface{}, mutation ObjectMutation, flags uint32 ) {
  for key, val := range mutation {
	if key[0] == '$' {
	  continue
	}
	if flags & CreateIDs == CreateIDs {
	  if _, ok := obj["_id"]; !ok {
		obj["_id"] = uniqueId()
	  }
	  obj["_rev"] = self.AppliedAtRevision()
	}
	if val == nil {
	  obj[key] = nil, false
	} else {
	  obj[key] = self.apply(obj[key], val, flags)
	}  
  }
}

func (self DocumentMutation) applyArrayMutation(arr []interface{}, mutation ArrayMutation, flags uint32) []interface{} {
  index := 0
  var v vec.Vector = arr  

  // Find the lifts
  lifts := make(map[string]interface{})
  for _, mut := range mutation.Array() {
	switch {
	  case IsInsertMutation(mut), IsSqueezeMutation(mut):
		continue;
	  case IsDeleteMutation(mut):
		index += toDeleteMutation(mut).Count()
	  case IsSkipMutation(mut):
		index += toSkipMutation(mut).Count()
	  case IsObjectMutation(mut), IsArrayMutation(mut), IsTextMutation(mut):
		index++
	  case IsLiftMutation(mut):
		l := toLiftMutation(mut)
		val := v[index]
		if l.HasMutation() {
		  val = self.apply(val, l.Mutation(), flags)
		}
		lifts[l.Id()] = val
		index++
	  default:
		panic("Should never happen")
	}
  }

  index = 0
  for _, m := range mutation.Array() {
	switch {
	  case IsDeleteMutation(m):
		count := toDeleteMutation(m).Count()
		for i := 0; i < count; i++ {
		  v.Delete(i)
        }
	  case IsSkipMutation(m):
		index += toSkipMutation(m).Count()
	  case IsLiftMutation(m):
		v.Delete(index)
	  case IsSqueezeMutation(m):
		v.Insert(index, lifts[toSqueezeMutation(m).Id()])
		index++
	  case IsObjectMutation(m), IsArrayMutation(m), IsTextMutation(m):
		v.Insert(index, self.apply( v[index], m, flags ))
		index++		
	  default:
		// Must be an insert mutation
		v.Insert(index, self.apply( nil, m, flags ))
		index++
    }
  }
  
  return v
}

func (self DocumentMutation) applyTextMutation(str string, mutation TextMutation, flags uint32) string {
  index := 0
  for _, m := range mutation.Array() {
	switch {
	  case IsDeleteMutation(m):
		count := toDeleteMutation(m).Count()
		str = str[:index] + str[index + count:]
	  case IsSkipMutation(m):
		index += toSkipMutation(m).Count()
	  default:
		// Must be an insert mutation, e.g. a string
		s := m.(string)
		str = str[:index] + s + str[index:]
		index += len(s)
    }
  }
  
  return str
}

func (self DocumentMutation) applyInsertMutation(mutation interface{}, flags uint32) interface{} {
  if obj, ok := mutation.(map[string]interface{}); ok {
	return self.applyInsertObjectMutation(obj, flags)
  }
  if arr, ok := mutation.([]interface{}); ok {
	return self.applyInsertArrayMutation(arr, flags)
  }
  return mutation
}

func (self DocumentMutation) applyInsertObjectMutation(obj map[string]interface{}, flags uint32) map[string]interface{} {
  m := make(map[string]interface{})
  for key, val := range obj {
	m[key] = self.applyInsertMutation(val, flags)
  }
  if flags & CreateIDs == CreateIDs {
	if _, ok := m["_id"]; !ok {
	  m["_id"] = uniqueId()
	}
	m["_rev"] = self.AppliedAtRevision()
  }
  return m
}

func (self DocumentMutation) applyInsertArrayMutation(arr []interface{}, flags uint32) []interface{} {
  a := make([]interface{}, len(arr))
  for i, val := range arr {
	a[i] = self.applyInsertMutation(val, flags)
  }
  return a
}
