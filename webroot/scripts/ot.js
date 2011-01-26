if ( !window.LW ) {
  LW = { };
}

if ( !LW.JsonOT ) {
  LW.JsonOT = { };
}

// Flags for mutation application
LW.JsonOT.CreateIDs = 1;

LW.JsonOT.uniqueIdCounter_ = 1;
LW.JsonOT.uniqueId_ = function() {
  return LW.JsonOT.uniqueIdCounter_++;
};

// Flags for mutation events
LW.JsonOT.AttributeModified = 1;
LW.JsonOT.AttributeInserted = 10;
LW.JsonOT.AttributeDeleted = 11;
LW.JsonOT.ObjectModified = 2;
LW.JsonOT.ArrayModified = 4;
LW.JsonOT.ArrayElementModified = 5;
LW.JsonOT.ArrayElementDeleted = 6;
LW.JsonOT.ArrayElementInserted = 7;
LW.JsonOT.ArrayElementLifted = 8;
LW.JsonOT.ArrayElementSqueezed = 9;

LW.JsonOT.applyDocMutation = function( doc, docmutation, flags ) {
    if ( docmutation._data ) {
        var inserted = doc._data == null;
        var callback = function(val) {
            if ( doc._cb_data ) { doc._data = val; doc._cb_data(doc, doc, "_data", docmutation._data, LW.JsonOT.AttributeInserted); }
            if ( doc._cb ) { doc._data = val; doc._cb(doc, doc, "_data", docmutation._data, LW.JsonOT.AttributeInserted); }
        }
        doc._data = LW.JsonOT.applyMutation_(doc, doc._data, docmutation._data, flags, callback );
        // Event
        if ( doc._data_cb && !inserted ) {
            doc._data_cb(doc, doc, "_data", docmutation._data, LW.JsonOT.AttributeModified );
        }
        if ( doc._cb && !inserted ) {
            doc._cb(doc, doc, "_data", docmutation._data, LW.JsonOT.AttributeModified );
        }
    }
    if ( docmutation._meta ) {
        var inserted = doc._meta == null;
        var callback = function(val) {
            if ( doc._cb_meta ) { doc._meta = val; doc._cb_meta(doc, doc, "_meta", docmutation._meta, LW.JsonOT.AttributeInserted); }
            if ( doc._cb ) { doc._meta = val; doc._cb(doc, doc, "_meta", docmutation._meta, LW.JsonOT.AttributeInserted); }
        }
        doc._meta = LW.JsonOT.applyMutation_(doc, doc._meta, docmutation._meta, flags, callback );
        // Event
        if ( doc._meta_cb && !inserted ) {
            doc._meta_cb(doc, doc, "_meta", docmutation._meta, LW.JsonOT.AttributeModified );
            if ( doc._meta_cb && !inserted ) {
                doc._meta_cb(doc, doc, "_meta", docmutation._meta, LW.JsonOT.AttributeModified );
            }
            if ( doc._cb && !inserted ) {
                doc._cb(doc, doc, "_meta", docmutation._meta, LW.JsonOT.AttributeModified );
            }
        }
    }
    // Event
    if ( doc._cb ) {
        doc._cb(doc, doc, null, docmutation, LW.JsonOT.ObjectModified );
    }
};

LW.JsonOT.applyMutation_ = function( doc, val, mutation, flags, insertCallback ) {
    if ( mutation["$object"] == true ) {
        if ( !val ) {
            val = { };
            if ( insertCallback ) { insertCallback(val); }
        }
        return LW.JsonOT.applyObjMutation_( doc, val, mutation, flags )
    } else if ( mutation["$array"] ) {
        if ( !val ) {
            val = [ ];
            if ( insertCallback ) { insertCallback(val); }
        }
        return LW.JsonOT.applyArrayMutation_( doc, val, mutation, flags )
    } else if ( mutation["$text"] ) {
        if ( !val ) {
            val = "";
            if ( insertCallback ) { insertCallback(val); }
        }
        return LW.JsonOT.applyTextMutation_( doc, val, mutation, flags )
    } else if ( mutation["$rtf"] ) {
        if ( !val ) {
            val = new LW.Richtext();
            if ( insertCallback ) { insertCallback(val); }
        }
        return LW.JsonOT.applyRichtextMutation_( doc, val, mutation, flags )
    } else {
        return LW.JsonOT.applyInsertMutation_( doc, mutation, flags, insertCallback )
    }
};

LW.JsonOT.applyObjMutation_ = function( doc, obj, mutation, flags ) {
  for ( key in mutation ) {
    var m = mutation[key];
    if ( key[0] == '$' ) {
      continue;
    }
    if ( flags & LW.JsonOT.CreateIDs == LW.JsonOT.CreateIDs ) {
      if ( !obj._id ) {
        obj._id = LW.JsonOT.uniqueId_();
      }
      obj["_rev"] = doc._rev;
    }
    var event = obj["_cb_" + key];
    var v = obj[key];
    if ( m === null ) {
      // Event
      if ( event ) {
        event(doc, obj, key, m, LW.JsonOT.AttributeDeleted);
      }
      if ( obj._cb ) {
        obj._cb(doc, obj, key, m, LW.JsonOT.AttributeDeleted);
      }
      delete obj[key];      
    } else {
      var callback = function(val) {
        if ( event ) { obj[key] = val; event(doc, obj, key, m, LW.JsonOT.AttributeInserted); }
        if ( obj._cb ) { obj[key] = val; obj._cb(doc, obj, key, m, LW.JsonOT.AttributeInserted); }
      }
      obj[key] = LW.JsonOT.applyMutation_(doc, obj[key], m, flags, callback);
      // Event
      if ( event ) {
        event(doc, obj, key, m, LW.JsonOT.AttributeModified);
      }
      if ( obj._cb ) {
        obj._cb(doc, obj, key, m, LW.JsonOT.AttributeModified);
      }
    }
  }
  // Event
  if ( obj._cb ) {
    obj._cb( doc, obj, null, mutation, LW.JsonOT.ObjectModified );
  }
  return obj;
};

LW.JsonOT.applyArrayMutation_ = function( doc, arr, mutation, flags ) {
  var index = 0;

  // Find the lifts
  var lifts = {};
  for ( var i = 0; i < mutation["$array"].length; i++ ) {
    // if ( i[0] == "_" ) continue;
    // Skip event handlers
    var mut = mutation["$array"][i];
    if ( mut["$delete"] != null ) {
      index += mut["$delete"];
    } else if ( mut["$skip"] != null ) {
      index += mut["$skip"];
    } else if ( mut["$object"] == true || mut["$array"] || mut["$text"] ) {
      index++;
    } else if ( mut["$lift"] ) {
      var val = arr[index];
      if ( mut["$mutation"] ) {
        val = LW.JsonOT.applyMutation_(doc, val, mut["$mutation"], flags);
      }
      lifts[mut["$lift"]] = val;
      index++;
    } else {
      // Do nothing by intention: InsertMutation or SqueezeMutation
    }
  }

  index = 0;
  for ( var i = 0; i < mutation["$array"].length; i++ ) {
    // Skip event handlers
    //if ( i[0] == "_" ) continue;
    var mut = mutation["$array"][i];
    if ( mut["$delete"] != null ) {
      arr.splice(index, mut["$delete"]);
      // Event
      if ( arr._cb_deleted ) {
        arr._cb_deleted( doc, arr, index, mut, LW.JsonOT.ArrayElementDeleted );
      }
    } else if ( mut["$skip"] != null ) {
      index += mut["$skip"];
    } else if ( mut["$lift"] ) {
      arr.splice(index, 1);
      // Event
      if ( arr._cb_lifted ) {
        arr._cb_lifted( doc, arr, index, mut, LW.JsonOT.ArrayElementLifted );
      }
    } else if ( mut["$squeeze"] ) {
      arr.splice(index, 0, lifts[mut["$squeeze"]]);
      // Event
      if ( arr._cb_squeezed ) {
        arr._cb_squeezed( doc, arr, index, mut, LW.JsonOT.ArrayElementSqueezed );
      }      
      index++;
    } else if ( mut["$object"] == true || mut["$array"] || mut["$text"] ) {
      arr[index] = LW.JsonOT.applyMutation_(doc, arr[index], mut, flags);
      // Event
      if ( arr._cb_modified ) {
        arr._cb_modified( doc, arr, index, mut, LW.JsonOT.ArrayElementModified );
      }
      index++;
    } else {
      // Insert mutation
      var callback = function(val) {
        arr.splice(index, 0, val);
        // Event
        if ( arr._cb_inserted ) {
          arr._cb_inserted( doc, arr, index, mut, LW.JsonOT.ArrayElementInserted );
        }
      }
      LW.JsonOT.applyInsertMutation_(doc, mut, flags, callback)
      index++;
    }
  }
  
  // Event
  if ( arr._cb ) {
    arr._cb( doc, obj, null, mutation, LW.JsonOT.ArrayModified );
  }

  return arr;
};

LW.JsonOT.applyTextMutation_ = function( doc, txt, mutation, flags ) {
  var index = 0;
  for ( var i in mutation["$text"] ) {
    mut = mutation["$text"][i];
    if ( mut["$delete"] != null ) {
      txt = txt.substr(0, index) + txt.substring(index + mut["$delete"], txt.length);
    } else if ( mut["$skip"] != null ) {
      index += mut["$skip"]
    } else {
      // Must be an insert mutation, e.g. a string
      txt = txt.substr(0, index) + mut + txt.substring(index, txt.length);
      index += mut.length;
    }
  }
  return txt;
};

LW.JsonOT.applyInsertMutation_ = function( doc, mutation, flags, insertCallback ) {
  if ( Array.isArray(mutation) ) {
    return LW.JsonOT.applyInsertArrayMutation_( doc, mutation, flags, insertCallback );
  } else if ( mutation != null && typeof(mutation) == "object" ) {
    return LW.JsonOT.applyInsertObjectMutation_( doc, mutation, flags, insertCallback );
  } else {
    if ( insertCallback ) insertCallback(mutation);
    return mutation;
  }
};

LW.JsonOT.applyInsertObjectMutation_ = function( doc, mutation, flags, insertCallback ) {
  var m = {};
  if (flags & LW.JsonOT.CreateIDs == LW.JsonOT.CreateIDs) {
    m._id = LW.JsonOT.uniqueId_();    
    m._rev = doc._rev;
  }
  if ( insertCallback ) insertCallback(m);
  for ( var key in mutation ) {
    var callback = function(val) {
      m[key] = val;
      if ( m["_cb_" + key] ) { m["_cb_" + key](doc, m, key, mutation[key], LW.JsonOT.AttributeInserted); }
      if ( m._cb ) { m._cb(doc, m, key, mutation[key], LW.JsonOT.AttributeInserted); }    
    };
    LW.JsonOT.applyInsertMutation_(doc, mutation[key], flags, callback);
  }
  if ( m._cb ) {
    m._cb(doc, m, null, mutation, LW.JsonOT.ObjectModified);
  }
  return m;
};

LW.JsonOT.applyInsertArrayMutation_ = function( doc, mutation, flags, insertCallback ) {
  var a = [];
  if (flags & LW.JsonOT.CreateIDs == LW.JsonOT.CreateIDs) {
    a._id = LW.JsonOT.uniqueId_();    
    a._rev = doc._rev;
  }
  if ( insertCallback ) insertCallback(a);
  for ( var i = 0; i < mutation.length; i++ ) {
    var callback = function(val) {
      a[i] = val;
      if ( a._cb_inserted ) { a._cb_inserted(doc, a, i, mutation[i], LW.JsonOT.ArrayElementInserted); }
    };
    LW.JsonOT.applyInsertMutation_(doc, mutation[i], flags, callback);
  }
  return a;
};

Array.isArray = Array.isArray || function(o) { return Object.prototype.toString.call(o) === '[object Array]'; };

LW.JsonOT.applyRichtextMutation_ = function( doc, richtext, mutation, flags ) {
    richtext.startUpdate();

    var pos = {paragIndex:0, charCount:0}
    for ( var i = 0; i < mutation["text"].length; ++i ) {
        mut = mutation["text"][i];
        if ( mut["$skip"] != null ) {
            var count = mut["$skip"]
            while( count > 0 ) {
                if ( pos.paragIndex >= richtext.paragraphs.length ) throw "FUCK";
                var min = Math.min(richtext.paragraphs[pos.paragIndex].text.length - pos.charCount, count);
                pos.charCount += min;
                count -= min;
                if ( count > 0 ) {
                    pos.paragIndex++;
                    pos.charCount = 0;
                    count--;
                }
            }
        } else if ( mut["$delete"] != null ) {
            var count = mut["$delete"]
            var pos1 = {paragIndex:pos.paragIndex, charCount:pos.charCount};
            while( count > 0 ) {
                if ( pos1.paragIndex >= richtext.paragraphs.length ) throw "FUCK";
                var min = Math.min(richtext.paragraphs[pos1.paragIndex].text.length - pos1.charCount, count);
                pos1.charCount += min;
                count -= min;
                if ( count > 0 ) {
                    pos1.paragIndex++;
                    pos1.charCount = 0;
                    count--;
                }
            }
            richtext.deleteRange(pos, pos1);
        } else if ( typeof(mut) == "string" ) {
            if ( pos.paragIndex >= richtext.paragraphs.length ) throw "fuck";
            richtext.insertText(pos, mut);
            pos.charCount += mut.length;
        } else if ( mut._type == "parag" ) {
            richtext.insertReturn(pos, mut);
            pos.charCount = 0;
            pos.paragIndex++;
        } else if ( mut["$object"] ) {
            // Transform a paragraph?
            if ( pos.charCount == richtext.paragraphs[paragIndex].text.length ) {
                LW.JsonOT.applyObjectMutation_(doc, richtext.paragraphs[pos.paragIndex].paragFormat, mut, flags);
            } else {
                LW.JsonOT.applyObjectMutation_(doc, richtext.paragraphs[paragIndex].paragFormat, mut, flags);
            }                
        } else if ( typeof("mut") == "object" ) {
            richtext.insertObject(pos, mut);
            pos.charCount += 1;
        } else {
            throw "Fuck";
        }
    }

    richtext.finishUpdate();
    return richtext;
};
