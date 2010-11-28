package wave

import (
  "reflect"
  "bytes"
  "strings"
  "fmt"
  "strconv"
  "os"
  "json"
)

func Marshal(val interface{}, buffer *bytes.Buffer) os.Error {
  if val == nil {
	buffer.WriteString("null")
	return nil
  }
  v := reflect.Indirect(reflect.NewValue(val)).(*reflect.StructValue)
  typ := v.Type().(*reflect.StructType)
  return enc_struct(v, typ, buffer)
}
  
func enc(val reflect.Value, typ reflect.Type, buffer *bytes.Buffer, req bool) (encoded bool, err os.Error) {  
  switch typ.(type) {
	case *reflect.PtrType:
	  // log.Println("pointer")
	  pt := typ.(*reflect.PtrType).Elem()
	  pv := reflect.Indirect(val)
	  // log.Println("Type is: ", pt.Name())
	  if pv == nil {
		if !req {
		  return false, nil
		} else {
		  return false, os.NewError("Required field is nil")
		}
	  }
	  return enc(pv, pt, buffer, req)
	case *reflect.BoolType:
	  if val.(*reflect.BoolValue).Get() {
		buffer.WriteString("true")
	  } else {
		buffer.WriteString( "false" )
	  }
	case *reflect.IntType:
	  buffer.WriteString( strconv.Itoa64( val.(*reflect.IntValue).Get()) )
	case *reflect.UintType:
	  buffer.WriteString( strconv.Uitoa64( val.(*reflect.UintValue).Get()) )
	case *reflect.StringType:
	  buffer.WriteString("\"")
	  // TODO: encode
	  buffer.WriteString( val.(*reflect.StringValue).Get() )
	  buffer.WriteString("\"")
	case *reflect.FloatType:
	  buffer.WriteString( strconv.Ftoa64( val.(*reflect.FloatValue).Get(), 'e', -1 ) )
	case *reflect.StructType:
	  enc_struct( val.(*reflect.StructValue), typ.(*reflect.StructType), buffer)
	case *reflect.SliceType:
	  st := typ.(*reflect.SliceType).Elem()
	  sv := val.(*reflect.SliceValue)
	  if sv.IsNil() {
		if req {
		  return false, os.NewError("Required field is nil")
		} else {
		  return false, nil
		}
	  }
	  buffer.WriteString("[")
	  for i := 0; i < sv.Len(); i++ {
		if i > 0 {
		  buffer.WriteString(",")
		}
		_, err := enc( sv.Elem(i), st, buffer, true )
		if err != nil {
		  return false, err
		}
	  }
	  buffer.WriteString("]")
	default:
	  return false, os.NewError("Unsupported type")
  }
  return true, nil
}

func enc_struct(val *reflect.StructValue, typ *reflect.StructType, buffer *bytes.Buffer) os.Error {  
  buffer.WriteString("{")
  first := true
  for i := 0; i < typ.NumField(); i++ {
	f := typ.Field(i)
	if f.Tag == "" {
	  continue
	}
	tags := strings.Split(f.Tag[3:len(f.Tag)-1], ",", -1)
	l := buffer.Len()
	if first {
	  buffer.WriteString( fmt.Sprintf("\"%v\":", tags[1] ) )
	} else {
	  buffer.WriteString( fmt.Sprintf(",\"%v\":", tags[1] ) )
	}
	written, err := enc( val.Field(i), f.Type, buffer, tags[2] == "req" )
	if err != nil {
	  return err
	}
	if !written {
	  buffer.Truncate(l)
	} else {
	  first = false
	}
  }
  buffer.WriteString("}")
  return nil
}

func Unmarshal(data []byte, val interface{}) os.Error {  
  if val == nil {
	return os.NewError("Object to unmarshal to must not be nil")
  }
  v, ok := reflect.Indirect(reflect.NewValue(val)).(*reflect.StructValue)
  if !ok {
	return os.NewError("Object to unmarshal to must be a struct type")
  }
  j := make(map[string]interface{})
  if err := json.Unmarshal(data, &j); err != nil {
	return err
  }
  typ := v.Type().(*reflect.StructType)
  return dec_struct(v, typ, j)
}

func dec_struct(val *reflect.StructValue, typ *reflect.StructType, json map[string]interface{}) os.Error {
  for i := 0; i < typ.NumField(); i++ {
	f := typ.Field(i)
	if f.Tag == "" {
	  continue
	}
	tags := strings.Split(f.Tag[3:len(f.Tag)-1], ",", -1)
	j, ok := json[tags[1]]
	if !ok {
	  if tags[2] == "req" {
		return os.NewError("Field " + f.Name + " is missing")
	  }
	  continue
	}
	err := dec( val.Field(i), f.Type, j )
	if err != nil {
	  return err
	}
  }
  return nil  
}

func dec(val reflect.Value, typ reflect.Type, json interface{}) os.Error {
  switch typ.(type) {
	case *reflect.PtrType:
	  // log.Println("pointer")
	  pt := typ.(*reflect.PtrType).Elem()
	  switch pt.(type) {
		case *reflect.BoolType:
		  b, ok := json.(bool)
		  if !ok {
			return os.NewError("Expected a boolean value")
		  }
		  val.(*reflect.PtrValue).PointTo( reflect.NewValue(b) )
		case *reflect.IntType:
		  f, ok := json.(float64)
		  if !ok {
			return os.NewError("Expected a numerical value")
		  }
		  v := reflect.MakeZero(pt).(*reflect.IntValue)
		  if v.Overflow(int64(f)) {
			return os.NewError("int overflow")
		  }
		  v.Set( int64(f) )
		  val.(*reflect.PtrValue).PointTo( v )
		case *reflect.UintType:
		  f, ok := json.(float64)
		  if !ok {
			return os.NewError("Expected a numerical value")
		  }
		  v := reflect.MakeZero(pt).(*reflect.UintValue)
		  if v.Overflow(uint64(f)) {
			return os.NewError("uint overflow")
		  }
		  v.Set( uint64(f) )
		  val.(*reflect.PtrValue).PointTo( v )
		case *reflect.StringType:
		  str, ok := json.(string)
		  if !ok {
			return os.NewError("Expected a string value")
		  }
		  val.(*reflect.PtrValue).PointTo( reflect.NewValue(str) )
		case *reflect.FloatType:
		  f, ok := json.(float64)
		  if !ok {
			return os.NewError("Expected a numerical value")
		  }
		  v := reflect.MakeZero(pt).(*reflect.FloatValue)
		  if v.Overflow(f) {
			return os.NewError("float overflow")
		  }
		  v.Set( f )
		  val.(*reflect.PtrValue).PointTo( v )
		case *reflect.StructType:
		  v := reflect.MakeZero(pt).(*reflect.StructValue)
		  jobj, ok := json.(map[string]interface{})
		  if !ok {
			return os.NewError("Expected JSON object")
		  }
		  err := dec_struct(v, pt.(*reflect.StructType), jobj)
		  if err != nil {
			return err
		  }
		  val.(*reflect.PtrValue).PointTo( v )
		default:
		  return os.NewError("Unsupported type")
	  }
	case *reflect.SliceType:
	  jarr, ok := json.([]interface{})
	  if !ok {
		return os.NewError("Expected JSON array")
	  }
	  l := len(jarr)
	  v := reflect.MakeSlice( typ.(*reflect.SliceType), l, l )
	  st := typ.(*reflect.SliceType).Elem()
	  for i := 0; i < l; i++ {
		dec( v.Elem(i), st, jarr[i] )
	  }
	  val.(*reflect.SliceValue).Set(v)	  
	case *reflect.BoolType:
	  b, ok := json.(bool)
	  if !ok {
		return os.NewError("Expected a boolean value")
	  }
	  val.(*reflect.BoolValue).Set(b)
	case *reflect.IntType:
	  f, ok := json.(float64)
	  if !ok {
		return os.NewError("Expected a numerical value")
	  }
	  if val.(*reflect.IntValue).Overflow(int64(f)) {
		return os.NewError("int overflow")
	  }
	  val.(*reflect.IntValue).Set( int64(f) )
	case *reflect.UintType:
	  f, ok := json.(float64)
	  if !ok {
		return os.NewError("Expected a numerical value")
	  }
	  if val.(*reflect.UintValue).Overflow(uint64(f)) {
		return os.NewError("uint overflow")
	  }
	  val.(*reflect.UintValue).Set( uint64(f) )
	case *reflect.StringType:
	  str, ok := json.(string)
	  if !ok {
		return os.NewError("Expected a string value")
	  }
	  val.(*reflect.StringValue).Set(str)
	case *reflect.FloatType:
	  f, ok := json.(float64)
	  if !ok {
		return os.NewError("Expected a numerical value")
	  }
	  if val.(*reflect.FloatValue).Overflow(f) {
		return os.NewError("int overflow")
	  }
	  val.(*reflect.FloatValue).Set( f )
	default:
	  return os.NewError("Unsupported type")
  }
  return nil 
}
