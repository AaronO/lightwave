package lightwave

import (
  vec "container/vector"
  "log"
  "fmt"
  "os"
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

func min(a, b int) int {
  if a < b {
    return a
  }
  return b
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
func (self DocumentMutation) Apply(obj map[string]interface{}, flags uint32) os.Error {
  data, okdata := self.DataMutation()
  if okdata {
    if err := self.Check(obj["_data"], data); err != nil {
      return err
    }
  }
  meta, okmeta := self.MetaMutation()
  if okmeta {
    if err := self.Check(obj["_meta"], meta); err != nil {
      return err
    }
  }
  if okdata {
    obj["_data"] = self.apply(obj["_data"], data, flags)
  }
  if okmeta {
    obj["_meta"] = self.apply(obj["_meta"], meta, flags)
  }
  return nil
}

func (self DocumentMutation) Check(val interface{}, mutation interface{} ) os.Error {
  switch {
  case IsObjectMutation(mutation):
    obj, test := val.(map[string]interface{})
    if !test {
      obj = make(map[string]interface{})
    }      
    return self.checkObjectMutation(obj, toObjectMutation(mutation))
  case IsArrayMutation(mutation):
    a, test := val.([]interface{})
    if !test {
      a = make([]interface{},0)[:]
    }
    return self.checkArrayMutation( a, toArrayMutation(mutation) )
  case IsTextMutation(mutation):
    s, test := val.(string)
    if !test {
      s = ""
    }      
    return self.checkTextMutation( s, toTextMutation(mutation) )
  case IsRichTextMutation(mutation):
    obj, test := val.(map[string]interface{})
    if !test {
      obj = make(map[string]interface{})
      obj["text"] = make([]interface{},0)[:]
    }      
    return self.checkRichTextMutation( obj, toRichTextMutation(mutation) )
  case IsInsertMutation(mutation):
    return self.checkInsertMutation(mutation)
  }
  log.Println("Unknown or unexpected mutation: ", mutation)
  return os.NewError("Unknown or unexpected mutation")
}

func (self DocumentMutation) checkInsertMutation(mutation interface{}) os.Error {
  if obj, ok := mutation.(map[string]interface{}); ok {
    return self.checkInsertObjectMutation(obj)
  }
  if arr, ok := mutation.([]interface{}); ok {
    return self.checkInsertArrayMutation(arr)
  }
  return nil
}

func (self DocumentMutation) checkInsertObjectMutation(obj map[string]interface{}) os.Error {
  for key, val := range obj {
    if key[0] == '$' {
      log.Println("$ is not allowed in attribute names")
      return os.NewError("$ is not allowed in attribute names")
    }
    if err := self.checkInsertMutation(val); err != nil {
      return err
    }
  }
  return nil
}

func (self DocumentMutation) checkInsertArrayMutation(arr []interface{}) os.Error {
  for _, val := range arr {
    if err := self.checkInsertMutation(val); err != nil {
      return err
    }
  }
  return nil
}

func (self DocumentMutation) checkObjectMutation(obj map[string]interface{}, mutation ObjectMutation ) os.Error {
  for key, mut := range mutation {
    if key[0] == '$' {
      continue
    }
    if IsInsertMutation(mut) {
      if err := self.checkInsertMutation(mut); err != nil {
        return err
      }
      continue
    }
    val, ok := obj[key]
    if !ok {
      log.Println("Object does not contain key ", key)
      return os.NewError("Object does not contain key " + key)
    }
    if err := self.Check(val, mut); err != nil {
      return err
    }
  }
  return nil
}

func (self DocumentMutation) checkArrayMutation(arr []interface{}, mutation ArrayMutation ) os.Error {
  squeezes := make(map[string]bool)
  lifts := make(map[string]bool)
  
  index := 0
  for _, mut := range mutation.Array() {
    if IsInsertMutation(mut) {
      if err := self.checkInsertMutation(mut); err != nil {
        return err
      }
      continue
    }
    if IsSqueezeMutation(mut) {
      s := toSqueezeMutation(mut)
      if _, ok := squeezes[s.Id()]; ok {
        return os.NewError("Duplicate use of squeeze id " + s.Id())
      }
      squeezes[s.Id()] = true
      continue
    }
    
    if index >= len(arr) {
      log.Println("Array mutation is too long")
      return os.NewError("Array mutation is too long")
    }
    
    switch {
      case IsDeleteMutation(mut):
        count := toDeleteMutation(mut).Count()
        if count <= 0 {
          log.Println("Delete <0 not allowed")
          return os.NewError("Delete <0 not allowed")
        }
        index += count
      case IsSkipMutation(mut):
        count := toSkipMutation(mut).Count()
        if count <= 0 {
          log.Println("Skip <0 not allowed")
          return os.NewError("Skip <0 not allowed")
        }
        index += count
      case IsObjectMutation(mut):
        val, ok := arr[index].(map[string]interface{})
        if !ok {
          log.Println("Expected object")
          return os.NewError("Expected object")
        }
        self.checkObjectMutation( val, toObjectMutation(mut))
        index++
      case IsArrayMutation(mut):
        val, ok := arr[index].([]interface{})
        if !ok {
          log.Println("Expected array")
          return os.NewError("Expected array")
        }
        self.checkArrayMutation( val, toArrayMutation(mut))
        index++
      case IsTextMutation(mut):
        val, ok := arr[index].(string)
        if !ok {
          log.Println("Expected string")
          return os.NewError("Expected string")
        }
        self.checkTextMutation( val, toTextMutation(mut))
        index++
      case IsLiftMutation(mut):
        l := toLiftMutation(mut)
        if l.HasMutation() {
	if err := self.Check(arr[index], l.Mutation()); err != nil {
          return err
        }
      }
        if _, ok := lifts[l.Id()]; ok {
        return os.NewError("Lift ID " + l.Id() + " is used twice")
        }
        lifts[l.Id()] = true
        index++
      default:
        log.Println("Unknown or unexpected mutation: ", mut)
      return os.NewError("Unknown or unexpected mutation")
    }
  }

  if index < len(arr) {
    log.Println("Array mutation is too small")
    log.Println(index)
    log.Println(len(arr))
    log.Println(arr)
    return os.NewError("Array mutation is too small")
  }
  if len(lifts) != len(squeezes) {
    log.Println("Not every lift has a squeeze or the other way round")
    return os.NewError("Not every lift has a squeeze or the other way round")
  }
  for id, _ := range lifts {
    if _, ok := squeezes[id]; !ok {
      log.Println("No squeeze for ", id)
      return os.NewError("No squeeze for " + id)
    }
  }
  return nil
}

func (self DocumentMutation) checkTextMutation(str string, mutation TextMutation ) os.Error {
  index := 0
  for _, mut := range mutation.Array() {
    if _, ok := mut.(string); ok {
      continue
    }
    if index >= len(str) {
      return os.NewError("Mutation is longer than text")
    }
    switch {
      case IsDeleteMutation(mut):
        count := toDeleteMutation(mut).Count()
        if count <= 0 {
        return os.NewError("Delete with <= 0 is not allowed")
        }
        index += count
      case IsSkipMutation(mut):
        count := toSkipMutation(mut).Count()
        if count <= 0 {
        return os.NewError("Skip with <= 0 is not allowed")
        }
        index += count
      default:
      return os.NewError("Mutation ot allowed on strings")
    }
  }
  if index < len(str) {
    return os.NewError("Mutation is smaller than string")
  }
  return nil
}

func (self DocumentMutation) checkRichTextMutation(obj map[string]interface{}, mutation RichTextMutation ) os.Error {
  // Test that the object is really a rich text object
  tmp, ok := obj["text"]
  if !ok {
    return os.NewError("Richtext object has no text field")
  }
  _, ok = tmp.([]interface{})
  if !ok {
    return os.NewError("text field of the richtext object is not an array")
  }
  
  index := 0
  for _, mut := range mutation.TextArray() {
    if _, ok := mut.(string); ok {
      continue
    }
    switch {
    case IsDeleteMutation(mut):
      count := toDeleteMutation(mut).Count()
      if count <= 0 {
        return os.NewError("Delete with <= 0 is not allowed")
      }
      index += count
    case IsSkipMutation(mut):
      count := toSkipMutation(mut).Count()
      if count <= 0 {
        return os.NewError("Skip with <= 0 is not allowed")
      }
      index += count
    case IsObjectMutation(mut):
      o := getRichTextObject(obj, index)
      if o == nil {
	return os.NewError("No object at position in richtext object")
      }
      if err := self.Check(o, mut); err != nil {
        return err
      }
      index += 1
    case IsObjectInsertMutation(mut):
      continue
    default:
      return os.NewError("Mutation not allowed in richtext object")
    }
  }
  if index != getRichTextLength(obj) {
    return os.NewError("Length of mutation and richtext object are different")
  }
  return nil
}

func getRichTextLength(obj map[string]interface{}) int {
  arr := obj["text"].([]interface{})
  count := 0
  for i := 0; i < len(arr); i++ {
    if str, ok := (arr[i]).(string); ok {
      count += len(str)
    } else if _, ok := (arr[i]).(map[string]interface{}); ok {
      count++
    }
  }
  return count
}

func getRichTextObject(obj map[string]interface{}, index int) map[string]interface{} {
  arr := obj["text"].([]interface{})
  count := 0
  for i := 0; i < len(arr); i++ {
    if str, ok := arr[i].(string); ok {
      count += len(str)
    } else if o, ok := arr[i].(map[string]interface{}); ok {
      if count == index {
	return o
      }
      count++
    }
    if count > index {
      return nil
    }
  }
  return nil
}

func (self DocumentMutation) apply(val interface{}, mutation interface{}, flags uint32 ) interface{} {
  switch {
    case IsObjectMutation(mutation):
      o, test := val.(map[string]interface{})
      if !test {
        o = make(map[string]interface{})
      }  
      self.applyObjectMutation(o, toObjectMutation(mutation), flags)
      return o
    case IsArrayMutation(mutation):
      a, test := val.([]interface{})
      if !test {
        a = make([]interface{},0)[:]
      }
      return self.applyArrayMutation( a, toArrayMutation(mutation), flags )
    case IsTextMutation(mutation):
      s, test := val.(string)
      if !test {
        s = ""
      }      
      return self.applyTextMutation( s, toTextMutation(mutation), flags )
    case IsRichTextMutation(mutation):
      o, test := val.(map[string]interface{})
      if !test {
        o = make(map[string]interface{})
        o["text"] = make([]interface{},0)[:]
      }  
      self.applyRichTextMutation( o, toRichTextMutation(mutation), flags )
      return o;
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
        v[index] = self.apply( v[index], m, flags )
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

func (self DocumentMutation) applyRichTextMutation(obj map[string]interface{}, mutation RichTextMutation, flags uint32) {
  if flags & CreateIDs == CreateIDs {
    if _, ok := obj["_id"]; !ok {
      obj["_id"] = uniqueId()
    }
    obj["_rev"] = self.AppliedAtRevision()
  }

  var text vec.Vector = obj["text"].([]interface{})
  index := 0
  inside := 0
  for _, m := range mutation.TextArray() {
    switch {
    case IsDeleteMutation(m):
      count := toDeleteMutation(m).Count()
      for count > 0 {
	if str, ok := text[index].(string); ok {
	  l := min(len(str) - inside, count)
	  inside += l
	  str = str[inside:]
	  if len(str) == 0 {
	    text.Delete(index)
	    inside = 0
	  } else if inside == len(str) {
	    index++
	    inside = 0
	    text[index] = str
	  } else {
	    text[index] = str
	  }
	  count -= count
	} else {
	  inside = 0
	  text.Delete(index)
	  count--
	}
      }
      // ** Ended up in the middle between two strings? -> Join them **
      if inside == 0 && index < len(text) {
	if str1, ok := text[index-1].(string); ok {
	  if str2, ok := text[index].(string); ok {
	    text[index-1] = str1 + str2;
	    text.Delete(index)
	    index--
	    inside = len(str1)
	    continue
	  }
	}
      } else if index < len(text) - 1 {
	if str1, ok := text[index].(string); inside == len(str1) && ok {
	  if str2, ok := text[index + 1].(string); ok {
	    text[index] = str1 + str2
	    inside = len(str1)
	    text.Delete(index + 1)
	    continue
	  }
	}
      }	
    case IsSkipMutation(m):
      count := toSkipMutation(m).Count()
      for count > 0 {
	if str, ok := text[index].(string); ok {
	  l := min(len(str) - inside, count)
	  inside += l
	  count -= count
	  if inside == len(str) {
	    index++
	    inside = 0
	  }
	} else {
	  inside = 0
	  index++
	  count--
	}
      }
    default:
      // *** A string? ***
      if s, ok := m.(string); ok {
	// Appending to a string?
	if inside == 0 && index > 0 {
	  if str, ok := text[index-1].(string); ok {
	    text[index-1] = str + s
	    continue
	  }
	}
	// End of string?
	if index == len(text) {
	  text.Insert(index, s)
	  index++
	  continue
	}
	// Prepending/Inserting to an existing string?
	if str, ok := text[index].(string); ok {
	// Insert into an existing string ?
	  text[index] = str[0:inside] + s + str[inside:]
	  inside += len(str)
	  continue
	}
	// Insert in front of an object
	text.Insert(index, s)
	index++
	continue
      }
      // *** Insert an object ***
      if index < len(text) && inside > 0 && inside == len(text[index].(string)) {
	// End of string?
	inside = 0
	index++
      }
      // Middle of a string -> split it
      if inside > 0 {
	str := text[index].(string)
	text.Insert(index + 1, str[inside:])
	text[index] = str[:inside]
	text.Insert(index + 1, m)
	inside = 0
	index += 2
      } else {
	text.Insert(index, m)
	index++
      }
    }
  }
    
  obj["text"] = []interface{}(text)
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
