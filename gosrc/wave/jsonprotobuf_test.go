package wave

import (
  "reflect"
  "bytes"
  "testing"
)

type DataStruct struct {
  Key *string "PB(bytes,1,req,name=key)"
  Value *string "PB(bytes,2,req,name=value)"
  Version *int64 "PB(varint,3,opt,name=version)"
  Arr []string "PB(varint,4,opt,name=version)"
  BArr []byte "PB(bytes,5,opt,name=version)"
  Next *DataStruct "PB(bytes,6,opt,name=key)"
  XXX_unrecognized []byte
}

func Test_EncDec(t *testing.T) {
  x := &DataStruct{}
  str := "Hallo"
  x.Key = &str
  str2 := "Welt"
  x.Value = &str2
  x.Arr = [...]string{"Wow", "Ui"}[:]
  x.BArr = [...]byte{100,102,104,106}[:]
  i := int64(123)
  x.Version = &i 
  y := &DataStruct{}
  str3 := "XHallo"
  y.Key = &str3
  str4 := "XWelt"
  y.Value = &str4
  x.Next = y
  
  b := &bytes.Buffer{}
  Marshal(x, b)
  
  x2 := &DataStruct{}
  err := Unmarshal( b.Bytes(), x2 )
  if err != nil {
	t.Errorf("Failed to unmashal data")
	return
  }
  if !reflect.DeepEqual(x, x2) {
	t.Errorf("Marshaled and Unmarshaled data is not equivalent")
  }
}