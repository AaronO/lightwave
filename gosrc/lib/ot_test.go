package lightwave

import (
  "testing"
  "log"
  "json"
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

func TestOT(t *testing.T) {
  text1 := "{\"_rev\":0, \"_data\":{\"$object\":true, \"abc\":\"xyz\", \"num\": 123, \"foo\":{\"$object\":true, \"a\":1,\"b\":3}, \"arr\":{\"$array\":[1,2,3]}}}"
  dest1 := make(map[string]interface{})
  err := json.Unmarshal([]byte(text1), &dest1)
  if ( err != nil ) {
    t.Errorf("Cannot parse: %v", err)
    return
  }

  text2 := "{\"_rev\":0, \"_data\":{\"$object\":true, \"torben\":\"weis\", \"num\": 321, \"foo\":{\"$object\":true, \"a\":6,\"c\":33}, \"arr\":{\"$array\":[4,5,6]}}}"
  dest2 := make(map[string]interface{})
  err = json.Unmarshal([]byte(text2), &dest2)
  if ( err != nil ) {
    t.Errorf("Cannot parse: %v", err)
    return
  }

  m1 := DocumentMutation(dest1)
  m2 := DocumentMutation(dest2)
  var ot Transformer
  err = ot.Transform( m1, m2 )
  if ( err != nil ) {
    t.Errorf("Cannot transform: %v", err)
    return
  }
  
  enc1, _ := json.Marshal(dest1);
  log.Println("t1=", string(enc1))
  enc2, _ := json.Marshal(dest2);
  log.Println("t2=", string(enc2))
}

func TestRichTextOT(t *testing.T) {
  text1 := "{\"_rev\":0, \"_data\":{\"$object\":true, \"richtext\":{\"$rtf\":true, \"text\":[{\"$skip\":5},{\"_type\":\"parag\", \"italic\":true},\"Hallo Welt\"]}}}"
  dest1 := make(map[string]interface{})
  err := json.Unmarshal([]byte(text1), &dest1)
  if ( err != nil ) {
    t.Errorf("Cannot parse: %v", err)
    return
  }
  dest1Original := make(map[string]interface{})
  err = json.Unmarshal([]byte(text1), &dest1Original)
  if ( err != nil ) {
    t.Errorf("Cannot parse: %v", err)
    return
  }

  text2 := "{\"_rev\":0, \"_data\":{\"$object\":true, \"richtext\":{\"$rtf\":true, \"text\":[{\"$skip\":5},{\"_type\":\"parag\", \"bold\":true},\"Wahnsinn\"]}}}"
  dest2 := make(map[string]interface{})
  err = json.Unmarshal([]byte(text2), &dest2)
  if ( err != nil ) {
    t.Errorf("Cannot parse: %v", err)
    return
  }
  dest2Original := make(map[string]interface{})
  err = json.Unmarshal([]byte(text2), &dest2Original)
  if ( err != nil ) {
    t.Errorf("Cannot parse: %v", err)
    return
  }

  m1Original := DocumentMutation(dest1Original)
  m2Original := DocumentMutation(dest2Original)

  m1 := DocumentMutation(dest1)
  m2 := DocumentMutation(dest2)
  var ot Transformer
  ot.Transform( m1, m2 )
  
  enc1, _ := json.Marshal(dest1)
  log.Println("t1=", string(enc1))
  enc2, _ := json.Marshal(dest2)
  log.Println("t2=", string(enc2))
  
  obj := make(map[string]interface{})
  data := make(map[string]interface{})
  meta := make(map[string]interface{})
  richtext := make(map[string]interface{})
  text := make([]interface{},2)
  line := make(map[string]interface{})
  line["_type"] = "parag"
  line["font-size"] = "12px"
  text[0] = line
  text[1] = "Huhu"
  richtext["text"] = text
  data["richtext"] = richtext
  obj["_data"] = data
  obj["_meta"] = meta
  obj["_rev"] = float64(0)
  
  obj2 := cloneJsonObject(obj)
  
  if err := m1Original.Apply(obj, 0); err != nil {
    t.Errorf("Failed applying m1Original: %v", err)
    return
  }
  enc, _ := json.Marshal(obj)
  log.Println("obj=", string(enc))
  if err := m2.Apply(obj, 0); err != nil {
    t.Errorf("Failed applying m2Original: %v", err)
    return
  }
  enc, _ = json.Marshal(obj)
  log.Println("obj=", string(enc))


  if err := m2Original.Apply(obj2, 0); err != nil {
    t.Errorf("Failed applying m1Original: %v", err)
    return
  }
  enc, _ = json.Marshal(obj2)
  log.Println("obj2=", string(enc))
  if err := m1.Apply(obj2, 0); err != nil {
    t.Errorf("Failed applying m2Original: %v", err)
    return
  }
  enc, _ = json.Marshal(obj2)
  log.Println("obj2=", string(enc))
  
  if !compareJson(obj, obj2 ) {
    t.Errorf("Different  OT paths yield a different result")
    return
  }
}
