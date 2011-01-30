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
        doc._data = LW.JsonOT.applyMutation_(doc, doc._data, docmutation._data, flags );
        // Event
        if ( doc._data_cb ) {
            doc._data_cb(doc, doc, "_data", docmutation._data, LW.JsonOT.AttributeModified );
        }
        if ( doc._cb ) {
            doc._cb(doc, doc, "_data", docmutation._data, LW.JsonOT.AttributeModified );
        }
    }
    if ( docmutation._meta ) {
        doc._meta = LW.JsonOT.applyMutation_(doc, doc._meta, docmutation._meta, flags );
        // Event
        if ( doc._meta_cb ) {
            doc._meta_cb(doc, doc, "_meta", docmutation._meta, LW.JsonOT.AttributeModified );
        }
        if ( doc._cb ) {
            doc._cb(doc, doc, "_meta", docmutation._meta, LW.JsonOT.AttributeModified );
        }
    }
    // Event
    if ( doc._cb ) {
        doc._cb(doc, doc, null, docmutation, LW.JsonOT.ObjectModified );
    }
};

LW.JsonOT.applyMutation_ = function( doc, val, mutation, flags ) {
    if ( mutation["$object"] == true ) {
        if ( !val ) {
            val = { };
        }
        return LW.JsonOT.applyObjMutation_( doc, val, mutation, flags )
    } else if ( mutation["$array"] ) {
        if ( !val ) {
            val = [ ];
        }
        return LW.JsonOT.applyArrayMutation_( doc, val, mutation, flags )
    } else if ( mutation["$text"] ) {
        if ( !val ) {
            val = "";
        }
        return LW.JsonOT.applyTextMutation_( doc, val, mutation, flags )
    } else if ( mutation["$rtf"] ) {
        if ( !val ) {
            val = new LW.Richtext();
        }
        return LW.JsonOT.applyRichtextMutation_( doc, val, mutation, flags )
    } else {
        return LW.JsonOT.applyInsertMutation_( doc, mutation, flags )
    }
};

LW.JsonOT.applyObjMutation_ = function( doc, obj, mutation, flags ) {
    // Iterate over all attributes in the mutation
    for ( key in mutation ) {
        var m = mutation[key];
        // Skip OT instructions, e.g. $object
        if ( key[0] == '$' ) {
            continue;
        }
        // Set IDs
        if ( flags & LW.JsonOT.CreateIDs == LW.JsonOT.CreateIDs ) {
            if ( !obj._id ) {
                obj._id = LW.JsonOT.uniqueId_();
            }
            obj["_rev"] = doc._rev;
        }
        var event = obj["_cb_" + key];
        // Attribute is being deleted?
        if ( m === null ) {
            delete obj[key];      
            // Event
            if ( event ) {
                event(doc, obj, key, m, LW.JsonOT.AttributeDeleted);
            }
            if ( obj._cb ) {
                obj._cb(doc, obj, key, m, LW.JsonOT.AttributeDeleted);
            }
        } else {
            var v = obj[key];
            obj[key] = LW.JsonOT.applyMutation_(doc, obj[key], m, flags);
            if ( typeof(obj[key]) == "object" ) {
                obj[key]._parent = obj;
            }
            // Event
            if ( event ) {
                event(doc, obj, key, m, (v == null ? LW.JsonOT.AttributeInserted : LW.JsonOT.AttributeModified) );
            }
            if ( obj._cb ) {
                obj._cb(doc, obj, key, m, (v == null ? LW.JsonOT.AttributeInserted : LW.JsonOT.AttributeModified) );
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
            var val = LW.JsonOT.applyInsertMutation_(doc, mut, flags);
            arr.splice(index, 0, val);
            if ( typeof(val) == "object" ) {
                val._parent = arr;
            }
            // Event
            if ( arr._cb_inserted ) {
                arr._cb_inserted( doc, arr, index, mut, LW.JsonOT.ArrayElementInserted );
            }
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
            index += mut["$skip"];
        } else {
            // Must be an insert mutation, e.g. a string
            txt = txt.substr(0, index) + mut + txt.substring(index, txt.length);
            index += mut.length;
        }
    }
    return txt;
};

LW.JsonOT.applyInsertMutation_ = function( doc, mutation, flags ) {
    if ( Array.isArray(mutation) ) {
        return LW.JsonOT.applyInsertArrayMutation_( doc, mutation, flags );
    } else if ( mutation != null && typeof(mutation) == "object" && mutation["$rtf"] ) {
        return LW.JsonOT.applyInsertRichtextMutation_( doc, mutation, flags );
    } else if ( mutation != null && typeof(mutation) == "object" ) {
        return LW.JsonOT.applyInsertObjectMutation_( doc, mutation, flags );
    } else {
        return mutation;
    }
};

LW.JsonOT.applyInsertRichtextMutation_ = function( doc, mutation, flags ) {
    var r = new LW.Richtext();
    if (flags & LW.JsonOT.CreateIDs == LW.JsonOT.CreateIDs) {
        r._id = LW.JsonOT.uniqueId_();    
        r._rev = doc._rev;
    }
    LW.JsonOT.applyRichtextMutation_(doc, r, mutation, flags);
    return r;
};

LW.JsonOT.applyInsertObjectMutation_ = function( doc, mutation, flags ) {
    var m = {};
    if (flags & LW.JsonOT.CreateIDs == LW.JsonOT.CreateIDs) {
        m._id = LW.JsonOT.uniqueId_();    
        m._rev = doc._rev;
    }
    for ( var key in mutation ) {
        m[key] = LW.JsonOT.applyInsertMutation_(doc, mutation[key], flags);
        if ( typeof(m[key]) == "object" ) {
            m[key]._parent = m;
        }
    }
    return m;
};

LW.JsonOT.applyInsertArrayMutation_ = function( doc, mutation, flags ) {
    var a = [];
    if (flags & LW.JsonOT.CreateIDs == LW.JsonOT.CreateIDs) {
        a._id = LW.JsonOT.uniqueId_();    
        a._rev = doc._rev;
    }
    for ( var i = 0; i < mutation.length; i++ ) {
        a[i] = LW.JsonOT.applyInsertMutation_(doc, mutation[i], flags);
        if ( typeof(a[i]) == "object" ) {
            a[i]._parent = a;
        }
    }
    return a;
};

Array.isArray = Array.isArray || function(o) { return Object.prototype.toString.call(o) === '[object Array]'; };

LW.JsonOT.applyRichtextMutation_ = function( doc, richtext, mutation, flags ) {
    richtext.startUpdate();
    var pos = {paragIndex:-1, charCount:0}
    for ( var i = 0; i < mutation["text"].length; ++i ) {
        mut = mutation["text"][i];
        if ( mut["$skip"] != null ) {
            var count = mut["$skip"]
            while( count > 0 ) {
                if ( pos.paragIndex >= richtext.paragraphs.length ) throw "FUCK";
                if ( pos.paragIndex == -1 || richtext.paragraphs[pos.paragIndex].text.length == pos.charCount ) {
                    pos.paragIndex++;
                    pos.charCount = 0;
                    count--;
                } else {
                    var min = Math.min(richtext.paragraphs[pos.paragIndex].text.length - pos.charCount, count);
                    pos.charCount += min;
                    count -= min;
                }
            }
        } else if ( mut["$delete"] != null ) {
            var count = mut["$delete"]
            var pos1 = {paragIndex:pos.paragIndex, charCount:pos.charCount};
            while( count > 0 ) {
                if ( pos1.paragIndex >= richtext.paragraphs.length ) throw "FUCK";

                if ( pos1.paragIndex == -1 || richtext.paragraphs[pos1.paragIndex].text.length == pos1.charCount ) {
                    pos1.paragIndex++;
                    pos1.charCount = 0;
                    count--;
                } else {
                    var min = Math.min(richtext.paragraphs[pos1.paragIndex].text.length - pos1.charCount, count);
                    pos1.charCount += min;
                    count -= min;
                }
            }
            richtext.deleteRange(pos, pos1);
        } else if ( typeof(mut) == "string" ) {
            if ( pos.paragIndex >= richtext.paragraphs.length || pos.paragIndex < 0 ) throw "fuck";
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
            if ( pos.paragIndex >= richtext.paragraphs.length || pos.paragIndex < 0 ) throw "fuck";
            richtext.insertObject(pos, mut);
            pos.charCount += 1;
        } else {
            throw "Fuck";
        }
    }

    richtext.finishUpdate();
    return richtext;
};
