/*
  The main package is here to stay
*/
package lightwave

import (
  "os"
  "math"
  vec "container/vector"
)

// ------------------------------------------------------------------------------
// Helper

func IntMin(j, i int) int {
  if j < i {
	return j
  }
  return i
}

// ------------------------------------------------------------------------------
// Json

type JsonObject map[string]interface{}

type JsonArray []interface{}

func ToJsonObject(value interface{}) (result JsonObject, ok bool) {
  m, ok := value.(map[string]interface{})
  return JsonObject(m), ok
}

func ToJsonArray(value interface{}) (result JsonArray, ok bool) {
  m, ok := value.([]interface{})
  return JsonArray(m), ok
}

func (self JsonObject) Clone() JsonObject {
  m := make(map[string]interface{})
  for key, val := range self {
	if obj, ok := ToJsonObject(val); ok {
	  m[key] = obj.Clone()
	} else if arr, ok := ToJsonArray(val); ok {
	  m[key] = arr.Clone()
	} else {
	  m[key] = val
	}
  }
  return JsonObject(m)
}

func (self JsonArray) Clone() JsonArray {
  a := make([]interface{}, len(self))
  for i, val := range self {
	if obj, ok := ToJsonObject(val); ok {
	  a[i] = obj.Clone()
	} else if arr, ok := ToJsonArray(val); ok {
	  a[i] = arr.Clone()
	} else {
	  a[i] = val
	}	
  }
  return JsonArray(a)
}

func (self JsonObject) GetAttribute(name string) (attr interface{}, err bool) {
  obj, e := self[name]
  return obj, e
}

func (self JsonObject) GetString(name string) (attr string, err bool) {
  obj, e := self[name]
  if !e {
	return "", false
  }
  str, e := obj.(string)
  if !e {
	return "", false
  }
  return str, true
}

func (self JsonObject) GetBool(name string) (attr bool, err bool) {
  obj, e := self[name]
  if !e {
	return false, false
  }
  b, e := obj.(bool)
  if !e {
	return false, false
  }
  return b, true
}

func (self JsonObject) GetInt(name string) (attr int, err bool) {
  obj, e := self[name]
  if !e {
	return 0, false
  }
  n, e := obj.(float64)
  if !e {
	return 0, false
  }
  return int(n), true
}

func (self JsonObject) GetFloat(name string) (attr float64, err bool) {
  obj, e := self[name]
  if !e {
	return 0, false
  }
  n, e := obj.(float64)
  if !e {
	return 0, false
  }
  return n, true
}

func (self JsonObject) GetArray(name string) (attr JsonArray, err bool) {
  obj, e := self[name]
  if !e {
	return nil, false
  }
  a, e := obj.([]interface{})
  if !e {
	return nil, false
  }
  return JsonArray(a), true
}

func (self JsonObject) GetObject(name string) (attr JsonObject, err bool) {
  obj, e := self[name]
  if !e {
	return nil, false
  }
  o, e := obj.(map[string]interface{})
  if !e {
	return nil, false
  }
  return JsonObject(o), true
}

func (self JsonArray) GetAttribute(index int, name string) (attr interface{}, err bool) {
  if index < 0 || index >= len(self) {
	return nil, false
  }
  obj := self[index]
  return obj, true
}

func (self JsonArray) GetObject(index int, name string) (attr JsonObject, err bool) {
  if index < 0 || index >= len(self) {
	return nil, false
  }
  obj := self[index]
  o, e := obj.(map[string]interface{})
  if !e {
	return nil, false
  }
  return JsonObject(o), true
}

func (self JsonArray) GetBool(index int, name string) (attr bool, err bool) {
  if index < 0 || index >= len(self) {
	return false, false
  }
  obj := self[index]
  b, e := obj.(bool)
  if !e {
	return false, false
  }
  return b, true
}

func (self JsonArray) GetString(index int, name string) (attr string, err bool) {
  if index < 0 || index >= len(self) {
	return "", false
  }
  obj := self[index]
  str, e := obj.(string)
  if !e {
	return "", false
  }
  return str, true
}

func (self JsonArray) GetInt(index int, name string) (attr int, err bool) {
  if index < 0 || index >= len(self) {
	return 0, false
  }
  obj := self[index]
  n, e := obj.(float64)
  if !e {
	return 0, false
  }
  return int(n), true
}

func (self JsonArray) GetFloat(index int, name string) (attr float64, err bool) {
  if index < 0 || index >= len(self) {
	return 0, false
  }
  obj := self[index]
  n, e := obj.(float64)
  if !e {
	return 0, false
  }
  return n, true
}

func (self JsonArray) GetArray(index int, name string) (attr JsonArray, err bool) {
  if index < 0 || index >= len(self) {
	return nil, false
  }
  obj := self[index]
  a, e := obj.([]interface{})
  if !e {
	return nil, false
  }
  return JsonArray(a), true
}

// ------------------------------------------------------------------------------
// Operational Transformation

func IsInsertMutation( obj interface{} ) bool {
  switch t := obj.(type) {
	case map[string]interface{}:
	  if _, ok := obj.(map[string]interface{})["$object"]; ok {
		return false
	  }
	  if _, ok := obj.(map[string]interface{})["$array"]; ok {
		return false
	  }
	  if _, ok := obj.(map[string]interface{})["$text"]; ok {
		return false
	  }
	  if _, ok := obj.(map[string]interface{})["$skip"]; ok {
		return false
	  }
	  if _, ok := obj.(map[string]interface{})["$delete"]; ok {
		return false
	  }
	  if _, ok := obj.(map[string]interface{})["$lift"]; ok {
		return false
	  }
	  if _, ok := obj.(map[string]interface{})["$squeeze"]; ok {
		return false
	  }	
	  return true
	case []interface{}, float64, bool, string, int, int64:
	  return true
  }
  return false
}

func IsObjectMutation( obj interface{} ) bool {
  switch t := obj.(type) {
	case map[string]interface{}:
	  o, err := obj.(map[string]interface{})["$object"]
	  if !err {
		return false
	  }
	  b, err := o.(bool)
	  if !err || !b {
		return false
	  }
	  return true
  }
  return false
}

func isMutationArray( obj interface{}, kind string ) bool {
  switch t := obj.(type) {
	case map[string]interface{}:
	  o, err := obj.(map[string]interface{})[kind]
	  if !err {
		return false
	  }
	  _, err = o.([]interface{})
	  if !err {
		return false
	  }
	  return true
  }
  return false
}

func IsArrayMutation( obj interface{} ) bool {
  return isMutationArray(obj, "$array")
}

func IsTextMutation( obj interface{} ) bool {
  return isMutationArray(obj, "$text")
}

func isSkipOrDeleteMutation( obj interface{}, kind string ) bool {
  switch t := obj.(type) {
	case map[string]interface{}:
	  o, err := obj.(map[string]interface{})[kind]
	  if !err {
		return false
	  }
	  n, err := o.(float64)
	  if !err || n < 0 || math.Ceil(n) != n {
		return false
	  }
	  return true
  }
  return false
}

func IsSkipMutation( obj interface{} ) bool {
  return isSkipOrDeleteMutation( obj, "$skip" )
}

func IsDeleteMutation( obj interface{} ) bool {
  return isSkipOrDeleteMutation( obj, "$delete" )
}

func isSqueezeOrLiftMutation( obj interface{}, kind string ) bool {
  switch t := obj.(type) {
	case map[string]interface{}:
	  o, err := obj.(map[string]interface{})[kind]
	  if !err {
		return false
	  }
	  _, err = o.(string)
	  if !err {
		return false
	  }
	  return true
  }
  return false
}

func IsSqueezeMutation( obj interface{} ) bool {
  return isSqueezeOrLiftMutation( obj, "$squeeze" )
}

func IsLiftMutation( obj interface{} ) bool {
  return isSqueezeOrLiftMutation( obj, "$lift" )
}

func IsDocumentMutation( obj interface{} ) bool {  
  o, ok := obj.(map[string]interface{})
  if !ok {
	return false
  }
  _, ok1 := o["_rev"]
  _, ok2 := o["_data"]
  _, ok3 := o["_meta"]
  if !ok1 || (!ok2 && !ok3) {
	return false
  }
  return true
}

//----------------------------------------------------------------
// Mutation

type Mutation interface{}

type LiftMutation map[string]interface{}

func (self LiftMutation) Id() string {
  return self["$lift"].(string)
}

func (self LiftMutation) Mutation() interface{} {
  return self["$mutation"]
}

func (self LiftMutation) HasMutation() bool {
  _, err := self["$mutation"]
  return err
}

func toLiftMutation( m interface{} ) LiftMutation {
  return LiftMutation( m.(map[string]interface{}) )
}

type SqueezeMutation map[string]interface{}

func (self SqueezeMutation) Id() string {
  return self["$squeeze"].(string)
}

func toSqueezeMutation( m interface{} ) SqueezeMutation {
  return SqueezeMutation( m.(map[string]interface{}) )
}

type SkipMutation map[string]interface{}

func (self SkipMutation) Count() int {
  return int(self["$skip"].(float64))
}

func (self SkipMutation) SetCount(count int) {
  self["$skip"] = float64(count)
}

func toSkipMutation( m interface{} ) SkipMutation {
  return SkipMutation( m.(map[string]interface{}) )
}

func NewSkipMutation(count int) SkipMutation {
  s := SkipMutation( make(map[string]interface{}) )
  s.SetCount(count)
  return s
}

type DeleteMutation map[string]interface{}

func toDeleteMutation( m interface{} ) DeleteMutation {
  return DeleteMutation( m.(map[string]interface{}) )
}

func (self DeleteMutation) Count() int {
  return int(self["$delete"].(float64))
}

func (self DeleteMutation) SetCount(count int) {
  self["$delete"] = float64(count)
}

func NewDeleteMutation(count int) DeleteMutation {
  s := DeleteMutation( make(map[string]interface{}) )
  s.SetCount(count)
  return s
}

type ObjectMutation map[string]interface{}

func (self ObjectMutation) RemoveAttribute(attr string) {
  self[attr] = nil, false
}

func toObjectMutation( m interface{} ) ObjectMutation {
  return ObjectMutation( m.(map[string]interface{}) )
}

type TextMutation map[string]interface{}

func (self TextMutation) Array() []interface{} {
  return self["$text"].([]interface{})
}

func (self TextMutation) SetArray(arr []interface{} ) {
  self["$text"] = arr
}

func toTextMutation( m interface{} ) TextMutation {
  return TextMutation( m.(map[string]interface{}) )
}

type ArrayMutation map[string]interface{}

func (self ArrayMutation) Array() []interface{} {
  return self["$array"].([]interface{})
}

func (self ArrayMutation) SetArray(arr []interface{} ) {
  self["$array"] = arr
}

func toArrayMutation( m interface{} ) ArrayMutation {
  return ArrayMutation( m.(map[string]interface{}) )
}

type InsertMutation Mutation

//----------------------------------------------------------------
// Transformer

type Transformer struct {
  sLifts map[string]LiftMutation
  cLifts map[string]LiftMutation
  sLiftCounterpart map[string]Mutation
  cLiftCounterpart map[string]Mutation
}

func (self Transformer) Transform( s, c ObjectMutation ) os.Error {
  self.transform_pass0_object( s, c )
  self.transform_pass1_object( s, c )
  return nil
}

func (self Transformer) transform_pass0_object( sobj, cobj ObjectMutation ) {  
  for name, sval := range sobj {
	// Do not handle the $object attribute
	if name == "$object" {
	  continue;
	}

	cval, err := cobj[name]
	if !err {
	  // There is nothing to transform for this attribute
	  continue
	}
	
	if IsInsertMutation(sval) || IsInsertMutation(cval) {
	  continue
	} else if ( IsObjectMutation(sval) && IsObjectMutation(cval) ) {
	  self.transform_pass0_object( toObjectMutation( sval ), toObjectMutation( cval ) )
	} else if ( IsArrayMutation(sval) && IsArrayMutation(cval) ) {
	  self.transform_pass0_array( toArrayMutation( sval ), toArrayMutation( cval ) )
	} else if ( IsTextMutation(sval) && IsTextMutation(cval) ) {
	  continue
	} else {
	  panic("The two mutations of the object are not compatible")
	}
  }
}

func (self Transformer) transform_pass0_array( sobj, cobj ArrayMutation ) {
  var (
    sindex int = 0
    cindex int = 0
    sinside int = 0
    cinside int = 0
	sarr []interface{} = sobj["$array"].([]interface{})
	carr []interface{} = cobj["$array"].([]interface{})
  )
  
  for sindex < len(sarr) || cindex < len(carr) {
	var smut, cmut interface{}
	if sindex < len(sarr) {
	  smut = sarr[sindex]
	}
	if cindex < len(carr) {
	  cmut = carr[cindex]
	}

	// Skip all server inserts
	for IsInsertMutation(smut) || IsSqueezeMutation(smut) {        
	  sindex++
	  if sindex < len(sarr) {
		smut = sarr[sindex]
	  } else {
		smut = nil
	  }
	}

	// Skip all client inserts
	for IsInsertMutation(cmut) || IsSqueezeMutation(cmut) {        
	  cindex++
	  if cindex < len(carr) {
		cmut = sarr[cindex]
	  } else {
		cmut = nil
	  }
	}

	// End of mutation reached?
	if sindex == len(sarr) || cindex == len(carr) {
	  break
	}
	
	if ( IsLiftMutation(smut) ) {
	  id := toLiftMutation(smut).Id()
	  self.sLifts[id] = toLiftMutation(smut);
	  self.cLiftCounterpart[id] = cmut;
	}
	if IsLiftMutation(cmut) {
	  id := toLiftMutation(cmut).Id()
	  self.cLifts[id] = toLiftMutation(cmut);
	  self.sLiftCounterpart[id] = smut;
	}
	
	if IsLiftMutation(smut) && IsLiftMutation(cmut) && toLiftMutation(smut).HasMutation() && toLiftMutation(cmut).HasMutation() {
	  self.transform_pass0_lift( toLiftMutation(smut).Mutation(), toLiftMutation(cmut).Mutation() )
	} else if IsLiftMutation(smut) && toLiftMutation(smut).HasMutation() && (IsArrayMutation(cmut) || IsTextMutation(cmut) || IsObjectMutation(cmut) ) {
	  self.transform_pass0_lift( toLiftMutation(smut).Mutation(), cmut )
	} else if IsLiftMutation(cmut) && toLiftMutation(cmut).HasMutation() && (IsArrayMutation(smut) || IsTextMutation(smut) || IsObjectMutation(smut) ) {
	  self.transform_pass0_lift( smut, toLiftMutation(cmut).Mutation() )
	}

	if ( IsDeleteMutation(smut) || IsSkipMutation(smut) ) && ( IsDeleteMutation(cmut) || IsSkipMutation(cmut)) {
	  var sdel, cdel int
	  if IsDeleteMutation(smut) {
		sdel = toDeleteMutation(smut).Count()
	  } else {
		sdel = toSkipMutation(smut).Count()
	  }
	  if IsDeleteMutation(cmut) {
		cdel = toDeleteMutation(cmut).Count()
	  } else {
		cdel = toSkipMutation(cmut).Count()
	  }
	  del := IntMin(sdel - sinside, cdel - cinside);
	  sinside += del;
	  cinside += del;
	  if sdel == del {
		sinside = 0;
		sindex++;
	  }
	  if cdel == del {
		cinside = 0;
		cindex++;
	  }
	} else if IsSkipMutation(smut) { // ... and mutation at the client
	  // Q_ASSERT(cmut.isArrayMutation() || cmut.isObjectMutation() || cmut.isRichTextMutation() || cmut.isTextMutation() || cmut.isLiftMutation() );
	  cindex++;
	  sinside++;
	  if toSkipMutation(smut).Count() == sinside {
		sinside = 0;
		sindex++;
	  }
	} else if IsSkipMutation(cmut) { // ... and mutation at the srver        
	  // Q_ASSERT(smut.isArrayMutation() || smut.isObjectMutation() || smut.isRichTextMutation() || smut.isTextMutation() || smut.isLiftMutation() );
	  sindex++;
	  cinside++;
	  if toSkipMutation(cmut).Count() == cinside {
		cinside = 0;
		cindex++;
	  }
	} else if IsDeleteMutation(smut) { // ... and mutation at the client
	  // Q_ASSERT(cmut.isArrayMutation() || cmut.isObjectMutation() || cmut.isRichTextMutation() || cmut.isTextMutation() || cmut.isLiftMutation() );
	  cindex++;
	  sinside++;
	  if toDeleteMutation(smut).Count() == sinside {            
		sinside = 0;
		sindex++;
	  }
	} else if IsDeleteMutation(cmut) { // ... and mutation at the server
	  // Q_ASSERT(smut.isArrayMutation() || smut.isObjectMutation() || smut.isRichTextMutation() || smut.isTextMutation() || smut.isLiftMutation() );
	  sindex++;
	  cinside++;
	  if toDeleteMutation(cmut).Count() == cinside {
		cinside = 0;
		cindex++;
	  }
	} else if IsLiftMutation(smut) {
	  // Q_ASSERT(cmut.isArrayMutation() || cmut.isObjectMutation() || cmut.isRichTextMutation() || cmut.isTextMutation() || cmut.isLiftMutation() );
	  sindex++;
	  cindex++;
	} else if IsLiftMutation(cmut) {
	  // Q_ASSERT(smut.isArrayMutation() || smut.isObjectMutation() || smut.isRichTextMutation() || smut.isTextMutation() || smut.isLiftMutation() );
	  sindex++;
	  cindex++;
	} else if IsArrayMutation(smut) && IsArrayMutation(cmut) {
	  self.transform_pass0_array( toArrayMutation(smut), toArrayMutation(cmut) )
	  cindex++;
	  sindex++;
	} else if IsObjectMutation(smut) && IsObjectMutation(cmut) {
	  self.transform_pass0_object( toObjectMutation(smut), toObjectMutation(cmut) )
	  cindex++;
	  sindex++;
	} else if IsTextMutation(smut) && IsTextMutation(cmut) {
	  cindex++;
	  sindex++;
	} else {
	  panic("The two mutations in the array do not match")
	}
  }  
}

func (self Transformer) transform_pass0_lift( s, c Mutation ) {
  if IsObjectMutation(s) && IsObjectMutation(c) {
	self.transform_pass0_object(toObjectMutation(s), toObjectMutation(c))
	self.transform_pass1_object(toObjectMutation(s), toObjectMutation(c))
  } else if IsArrayMutation(s) && IsArrayMutation(c) {
	self.transform_pass0_array(toArrayMutation(s), toArrayMutation(c))
	self.transform_pass1_array(toArrayMutation(s), toArrayMutation(c))
  } else if IsTextMutation(s) && IsTextMutation(c) {
	self.transform_pass1_text(toTextMutation(s), toTextMutation(c) )
  } else {
	panic("The two mutations are either incompatible or they are not allowed inside a lift")
  }
}

func (self Transformer) transform_pass1_object( sobj, cobj ObjectMutation ) {
  for name, smut := range sobj {
	
	if name == "$object" {
	  continue;
	}
	cmut, err := cobj[name]
	if !err {
	  continue
	}
	
	if IsInsertMutation(smut) {
	  cobj.RemoveAttribute(name);
	} else if IsObjectMutation(smut) {
	  if IsObjectMutation(cmut) {
		self.transform_pass1_object(toObjectMutation(smut), toObjectMutation(cmut) )
	  } else if IsInsertMutation(cmut) {
		sobj.RemoveAttribute(name)
	  } else {
		panic("The two mutations are not compatible and/or not allowed inside an object mutation. This should have been detected in pass 0")
	  }
	} else if IsArrayMutation(smut) {
	  if IsArrayMutation(cmut) {
		self.transform_pass1_array(toArrayMutation(smut), toArrayMutation(cmut) )
	  } else if IsInsertMutation(cmut) {
		sobj.RemoveAttribute(name)
	  } else {
		panic("The two mutations are not compatible and/or not allowed inside an object mutation. This should have been detected in pass 0")
	  }
	} else if IsTextMutation(smut) {
	  if IsTextMutation(cmut) {
		self.transform_pass1_text(toTextMutation(smut), toTextMutation(cmut) )
	  } else if IsInsertMutation(cmut) {
		sobj.RemoveAttribute(name);
	  } else {
		panic("The two mutations are not compatible and/or not allowed inside an object mutation. This should have been detected in pass 0")
	  }
	} else {
	  panic("This mutation is not allowed inside an object mutation. This should have been detected in pass 0")
	}
  }
}

func (self Transformer) transform_pass1_array( sobj, cobj ArrayMutation ) {
  var (
	sarr vec.Vector = sobj.Array()
	carr vec.Vector = cobj.Array()
    sindex int = 0
    cindex int = 0
    sinside int = 0
    cinside int = 0
  )
  
  // Loop until end of one mutation is reached
  for sindex < len(sarr) || cindex < len(carr) {
	var smut, cmut interface{}
	if sindex < len(sarr) {
	  smut = sarr[sindex]
	}
	if cindex < len(carr) {
	  cmut = carr[cindex]
	}

	// Server insert/squeeze go first
	for IsInsertMutation(smut) || IsSqueezeMutation(smut) {
	  if cinside > 0 {
		self.split(carr, cindex, cinside )
		cinside = 0
		cindex++
		cmut = carr[cindex]
	  }

	  if IsInsertMutation(smut) {
		// TODO: check the Insert mutation individually for correctness
		carr.Insert(cindex, NewSkipMutation(1))
		cindex++
		sindex++
	  } else if IsSqueezeMutation(smut) {
		sSqueeze := toSqueezeMutation(smut)
		// Which operation does the client side have for this object?
        c, hasCounterpart := self.cLiftCounterpart[sSqueeze.Id()]
        if !hasCounterpart || IsSkipMutation(c) {
		  // Server lift remains, client skips it at the new position
		  carr.Insert(cindex, NewSkipMutation(1))
		  cindex++
		  sindex++
		} else if IsDeleteMutation(c) {
		  // Client deletes the object at its new position
		  carr.Insert(cindex, NewDeleteMutation(1))
		  cindex++
		  // Server removes squeeze because it is already deleted by the client
		  sarr.Delete(sindex)
		} else if IsLiftMutation(c) {
		  // sLift := self.sLifts[sSqueeze.Id()]
		  cLift := toLiftMutation(c)
		  if cLift.HasMutation() {
			// The client does not lift it and therefore it mutates it at the new position
			carr.Insert(cindex, cLift.Mutation())
			cindex++
			// Server keeps its squeeze at this position
			sindex++
		  } else {
			// Client skips the squeezed object and does not lift it
			carr.Insert(cindex, NewSkipMutation(1))
			cindex++
			// Server keeps its squeeze at this position
			sindex++
		  }
		} else {
		  // The client somehow mutates the object. It must do so at its new position
		  carr.Insert(cindex, c)
		  cindex++
		  sindex++
		}
	  }
            
	  if sindex < len(sarr) {
		smut = sarr[sindex];
	  } else {
		smut = nil
	  }
	}
	
	// Client insert/squeeze go next
	for IsInsertMutation(cmut) || IsSqueezeMutation(cmut) {
	  if sinside > 0 {
		self.split(sarr, sindex, sinside )
		sinside = 0
		sindex++
		smut = sarr[sindex]
	  }

	  if IsInsertMutation(cmut) {
		// TODO: check the Insert mutation individually for correctness
		cindex++
		sarr.Insert(sindex, NewSkipMutation(1));
		sindex++
	  } else {
		cSqueeze := toSqueezeMutation(cmut)
		// Which operation does the server side have for this object?
		s, hasCounterpart := self.sLiftCounterpart[cSqueeze.Id()]
		if !hasCounterpart || IsSkipMutation(s) {
		  sarr.Insert(sindex, NewSkipMutation(1))
		  sindex++
		  cindex++
		} else if IsDeleteMutation(s) {
		  // Server deletes the object at its the new position
		  sarr.Insert(sindex, NewDeleteMutation(1))
		  sindex++
		  // Client removes squeeze because it is already deleted by the client
		  carr.Delete(cindex)
		} else if IsLiftMutation(s) {
		  // The server must lift the object here instead
		  sarr.Insert(sindex, s)
		  sindex++
		  // The server lifted this object as well -> the client cannot lift it ->
		  // then the client cannot squeeze it in here
		  carr.Delete(cindex)
		} else {
		  // The server mutates the object at its new position
		  sarr.Insert(sindex, s)
		  sindex++
		  // The client squeezes the object in here
		  cindex++
		}
	  }
	  
	  if cindex < len(carr) {
		cmut = carr[sindex];
	  } else {
		cmut = nil
	  }
	}

	if sindex == len(sarr) && cindex == len(carr) {
	  break
	}	
	if sindex == len(sarr) || cindex == len(carr) {
	  panic("The mutations do not have equal length")
	}

	//
	// Lift, Skip, Delete, mutations
	//

	if IsLiftMutation(smut) {
	  // Both are lifting the same object?
	  if IsLiftMutation(cmut) {
		// The client removes its lift. If it has a mutation then it will be moved to the corresponding squeeze
		carr.Delete(cindex)
		// The server removes its lift. It will be placed where the client moved it.
		sarr.Delete(sindex)
	  } else if IsDeleteMutation(cmut) {
		// The server does not lift the object because it is deleted
		sarr.Delete(sindex)
		// The client removes its delete. The delete is put where the server squeezed it.
		cindex, cinside = self.shorten(carr, cindex, cinside, 1)		
	  } else if IsSkipMutation(cmut) {
		// The client does not skip this element here. It is skipped where it is squeezed in
		cindex, cinside = self.shorten(carr, cindex, cinside, 1)
		// Server remains with its lift
		sindex++
	  } else {
		// Q_ASSERT(cmut.isArrayMutation() || cmut.isObjectMutation() || cmut.isTextMutation() || cmut.isRichTextMutation() );
		// The client removes its mutation here. It is shifted to the new position where the object is squeezed in.
		carr.Delete(cindex);
		// Server remains with its lift
		sindex++
	  }
	} else if IsLiftMutation(cmut) {
	  if IsDeleteMutation(smut) {
		// The client does not lift the object because it is deleted
		carr.Delete(cindex)
		// The server removes its delete. The delete is put where the client squeezed it.
		sindex, sinside = self.shorten(sarr, sindex, sinside, 1)
	  } else if IsSkipMutation(smut) {
		// The server does not skip this element here. It is skipped where it is squeezed in
		sindex, sinside = self.shorten(sarr, sindex, sinside, 1)
		// Client remains with its lift
		cindex++;
	  } else {
		// Q_ASSERT(smut.isArrayMutation() || smut.isObjectMutation() || smut.isTextMutation() || smut.isRichTextMutation() );
		// The server removes its mutation here. It is shifted to the new position where the object is squeezed in.
		sarr.Delete(sindex)
		// Client remains with its lift
		cindex++
	  }
	} else if IsDeleteMutation(smut) && IsDeleteMutation(cmut) {
	  sdel := toDeleteMutation(smut).Count()
	  cdel := toDeleteMutation(cmut).Count()
	  del := IntMin(sdel - sinside, cdel - cinside)
	  sindex, sinside = self.shorten(sarr, sindex, sinside, del)
	  cindex, cinside = self.shorten(carr, cindex, cinside, del)
	} else if IsSkipMutation(smut) && IsSkipMutation(cmut) {
	  sskip := toSkipMutation(smut).Count()
	  cskip := toSkipMutation(cmut).Count()
	  skip := IntMin(sskip - sinside, cskip - cinside)
	  sinside += skip;
	  cinside += skip;
	  if sinside == sskip {
		sinside = 0
		sindex++
	  }
	  if cinside == cskip {
		cinside = 0;
		cindex++
	  }
	} else if IsDeleteMutation(smut) && IsSkipMutation(cmut) {
	  sdel := toDeleteMutation(smut).Count()
	  cskip := toSkipMutation(cmut).Count()
	  count := IntMin(sdel - sinside, cskip - cinside)
	  sinside += count
	  if sinside == sdel {
		sinside = 0
		sindex++
	  }
	  cindex, cinside = self.shorten(carr, cindex, cinside, count)
	} else if IsSkipMutation(smut) && IsDeleteMutation(cmut) {
	  sskip := toSkipMutation(smut).Count()
	  cdel := toDeleteMutation(cmut).Count()
	  count := IntMin(sskip - sinside, cdel - cinside)
	  sindex, sinside = self.shorten(sarr, sindex, sinside, count)
	  cinside += count
	  if cinside == cdel {
		cinside = 0;
		cindex++
	  }
	} else if IsSkipMutation(smut) { // ... and mutation at the client
	  cindex++
	  sinside++
	  if toSkipMutation(smut).Count() == sinside {
		sinside = 0
		sindex++
	  }
	} else if IsSkipMutation(cmut) { // ... and mutation at the srver
	  sindex++
	  cinside++
	  if toSkipMutation(cmut).Count() == cinside {
		cinside = 0
		cindex++
	  }
	} else if IsDeleteMutation(smut) { // ... and mutation at the client
	  carr.Delete(cindex)
	  sinside++
	  if toDeleteMutation(smut).Count() == sinside {
		sinside = 0
		sindex++
	  }
	} else if IsDeleteMutation(cmut) { // ... and mutation at the server        
	  sarr.Delete(sindex)
	  cinside++
	  if toDeleteMutation(cmut).Count() == cinside {
		cinside = 0
		cindex++
	  }
	} else {
	  if IsObjectMutation(smut) && IsObjectMutation(cmut) {
		self.transform_pass1_object(toObjectMutation(smut), toObjectMutation(cmut) )
	  } else if IsArrayMutation(smut) && IsArrayMutation(cmut) {
		self.transform_pass1_array(toArrayMutation(smut), toArrayMutation(cmut) )
	  } else if IsTextMutation(smut) && IsTextMutation(cmut) {
		self.transform_pass1_text(toTextMutation(smut), toTextMutation(cmut) )
	  } else {
		  panic("The mutations are not compatible or not allowed inside an array mutation")
	  }
	  sindex++
	  cindex++
	}
  }

  sobj.SetArray(sarr)
  cobj.SetArray(carr)
}

func (self Transformer) transform_pass1_text( sobj, cobj TextMutation ) {
  var (
	sarr vec.Vector = sobj.Array()
	carr vec.Vector = cobj.Array()
    sindex int = 0
    cindex int = 0
    sinside int = 0
    cinside int = 0
  )
  
  // Loop until end of one mutation is reached
  for sindex < len(sarr) || cindex < len(carr) {
	var smut, cmut interface{}
	if sindex < len(sarr) {
	  smut = sarr[sindex]
	}
	if cindex < len(carr) {
	  cmut = carr[cindex]
	}
  
	// Server insertions go first
	if IsInsertMutation(smut) {
	  // In the middle of a client skip/delete? Split it
	  if cinside > 0 {
		self.split(carr, cindex, cinside )
		cinside = 0
		cindex++
		cmut = carr[cindex]
	  }
	  str, err := smut.(string)
	  if !err {
		panic("Only strings allowed inside a text mutation")
	  }
	  sindex++;
	  if len(str) > 0 {
		carr.Insert(cindex, NewSkipMutation(len(str)))
		cindex++
	  }
	  continue;
	}
	// Client insertions go next
	if IsInsertMutation(cmut)  {
	  // In the middle of a server skip/delete? Split it
	  if sinside > 0 {
		self.split(sarr, sindex, sinside )
		sinside = 0
		sindex++
		smut = sarr[sindex]
	  }

	  str, err := cmut.(string)
	  if !err {
		panic("Only strings allowed inside a text mutation")
	  }
	  cindex++
	  if len(str) > 0 {
		sarr.Insert(sindex, NewSkipMutation(len(str)))
		sindex++
	  }
	  continue;
	}

	if sindex == len(sarr) && cindex == len(carr) {
	  break
	}	
	if sindex == len(sarr) || cindex == len(carr) {
	  panic("The mutations do not have equal length")
	}

	if IsDeleteMutation(smut) && IsDeleteMutation(cmut) {
	  sdel := toDeleteMutation(smut).Count()
	  cdel := toDeleteMutation(cmut).Count()
	  del := IntMin(sdel - sinside, cdel - cinside)
	  sindex, sinside = self.shorten(sarr, sindex, sinside, del)
	  cindex, cinside = self.shorten(carr, cindex, cinside, del)
	} else if IsSkipMutation(smut) && IsSkipMutation(cmut) {
	  sskip := toSkipMutation(smut).Count()
	  cskip := toSkipMutation(cmut).Count()
	  skip := IntMin(sskip - sinside, cskip - cinside)
	  sinside += skip;
	  cinside += skip;
	  if sinside == sskip {
		sinside = 0
		sindex++
	  }
	  if cinside == cskip {
		cinside = 0;
		cindex++
	  }
	} else if IsDeleteMutation(smut) && IsSkipMutation(cmut) {
	  sdel := toDeleteMutation(smut).Count()
	  cskip := toSkipMutation(cmut).Count()
	  count := IntMin(sdel - sinside, cskip - cinside)
	  sinside += count
	  if sinside == sdel {
		sinside = 0
		sindex++
	  }
	  cindex, cinside = self.shorten(carr, cindex, cinside, count)
	} else if IsSkipMutation(smut) && IsDeleteMutation(cmut) {
	  sskip := toSkipMutation(smut).Count()
	  cdel := toDeleteMutation(cmut).Count()
	  count := IntMin(sskip - sinside, cdel - cinside)
	  sindex, sinside = self.shorten(sarr, sindex, sinside, count)
	  cinside += count
	  if cinside == cdel {
		cinside = 0;
		cindex++
	  }
	} else {
	  panic("Mutation not allowed in a text mutation")
	}
  }

  sobj.SetArray(sarr)
  cobj.SetArray(carr)
}

func (self Transformer) split( arr vec.Vector, index, inside int ) {
  mut := arr[index]
  if IsDeleteMutation(mut) {
	arr.Insert(index+1, NewDeleteMutation( toDeleteMutation(mut).Count() - inside))
	toDeleteMutation(mut).SetCount(inside);
  } else if IsSkipMutation(mut) {
	arr.Insert(index+1, NewSkipMutation( toSkipMutation(mut).Count() - inside))
	toSkipMutation(mut).SetCount(inside);
  } else {
	panic("Unsupported mutation for split()")
  }
}

func (self Transformer) shorten( arr vec.Vector, index, inside, count int ) (rindex int, rinside int) {
  mut := arr[index]
  if IsDeleteMutation(mut) {
	del := toDeleteMutation(mut)
	del.SetCount( del.Count() - count )
	if inside == del.Count() {
	  inside = 0;
	  if del.Count() == 0 {
		arr.Delete(index)
	  } else {
		index++;
	  }
	}
  } else if IsSkipMutation(mut) {
	skip := toSkipMutation(mut)
	skip.SetCount( skip.Count() - 1 );
	if inside == skip.Count() {
	  inside = 0;
	  if skip.Count() == 0 {
		arr.Delete(index)
	  } else {
		index++
	  }
	}  
  } else {
	panic("Unsupported mutation for shorten()")
  }

  return index, inside
}

/*
func main() {

  text1 := "{\"$object\":true, \"abc\":\"xyz\", \"num\": 123, \"foo\":{\"$object\":true, \"a\":1,\"b\":3}, \"arr\":{\"$array\":[1,2,3]}}"
  dest1 := make(map[string]interface{})
  err := json.Unmarshal([]byte(text1), &dest1)
  if ( err != nil ) {
	log.Exitf("Cannot parse: %v", err);
  }

  text2 := "{\"$object\":true, \"torben\":\"weis\", \"num\": 321, \"foo\":{\"$object\":true, \"a\":6,\"c\":33}, \"arr\":{\"$array\":[4,5,6]}}"
  dest2 := make(map[string]interface{})
  err = json.Unmarshal([]byte(text2), &dest2)
  if ( err != nil ) {
	log.Exitf("Cannot parse: %v", err);
  }

  var t Transformer
  t.Transform( toObjectMutation(dest1), toObjectMutation(dest2) )
  
  enc1, _ := json.Marshal(dest1);
  fmt.Printf("t1=%v\n", string(enc1))
  enc2, _ := json.Marshal(dest2);
  fmt.Printf("t2=%v\n", string(enc2))
}
*/