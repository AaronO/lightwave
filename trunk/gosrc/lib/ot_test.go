package lightwave

import (
  "testing"
  "log"
  "json"
)

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
  ot.Transform( m1, m2 )
  
  enc1, _ := json.Marshal(dest1);
  log.Println("t1=", string(enc1))
  enc2, _ := json.Marshal(dest2);
  log.Println("t2=", string(enc2))
}
