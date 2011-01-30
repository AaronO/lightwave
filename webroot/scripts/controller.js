if ( !window.LW ) {
    LW = { };
}
LW.Controller = {
};

LW.Controller.getDomElement_ = function(doc, objectid) {
    var domid = doc.url + "!" + objectid;
    return document.getElementById(domid);
};

/**
 * Binds another controller to the attribute of a javascript object.
 * If the attribute is replaced or deleted the other controller is bound of unbound accordingly.
 */
LW.Controller.AttributeController = function(dom, jsParentKey, onBindFunc, onUnbindFunc) {
    this.jsParentKey = jsParentKey;
    this.onBindFunc = onBindFunc;
    this.onUnbindFunc = onUnbindFunc;
    this.state = { dom: dom };
};

LW.Controller.AttributeController.prototype.bind = function(jsDoc, jsParentObj) {
    this.unbind();
    this.jsDoc = jsDoc;
    this.jsParentObj = jsParentObj;

    var self = this;
    var func = function(doc, obj, key, mutation, event) {
        if ( event == LW.JsonOT.AttributeInserted ) {
            if ( self.onUnbindFunc ) {
                self.onUnbindFunc(self.jsDoc, self.jsParentObj[self.jsParentKey], self.state);
            }
            if ( self.onBindFunc ) {
                self.onBindFunc(self.jsDoc, obj[key], self.state);
            }
        }
        else if ( event == LW.JsonOT.AttributeDeleted ) {
            if ( self.onUnbindFunc ) {
                self.onUnbindFunc(self.jsDoc, self.jsParentObj[self.jsParentKey], self.state);
            }
        }
    }

    jsParentObj["_cb_" + this.jsParentKey] = func;
    if ( jsParentObj[this.jsParentKey] != null ) {
        func(jsDoc.content, jsParentObj, this.jsParentKey, null, LW.JsonOT.AttributeInserted );
    }
};

LW.Controller.AttributeController.prototype.unbind = function() {
    if ( this.jsParentObj ) {
        delete this.jsParentObj["_cb_" + this.jsParentKey];
        if ( this.onUnbindFunc ) {
            this.onUnbindFunc(this.jsDoc, this.jsParentObj[this.jsParentKey], this.state);
        }
    }
    delete this.jsDoc;
    delete this.jsParentObj;
};

/**
 * Binds another controller to the attribute of a javascript object.
 * If the attribute is replaced or deleted the other controller is bound of unbound accordingly.
 */
LW.Controller.ConditionController = function(dom, condition, onBindFunc, onUnbindFunc) {
    this.condition = condition;
    this.onBindFunc = onBindFunc;
    this.onUnbindFunc = onUnbindFunc;
    this.isbound_ = false;
    this.state = { dom: dom };
};

LW.Controller.ConditionController.prototype.bind = function(jsDoc, jsObject) {
    this.unbind();
    this.jsDoc = jsDoc;
    this.jsObject = jsObject;

    var self = this;
    var func = function(doc, obj, key, mutation, event) {
        if ( event == LW.JsonOT.ObjectModified ) {
            var bound = self.condition(self.jsDoc, self.jsObject, self.state);
            if ( bound == self.isbound_ ) {
                return;
            }
            self.isbound_ = bound;
            if ( !bound ) {
                if ( self.onUnbindFunc ) {
                    self.onUnbindFunc(self.jsDoc, self.jsObject, self.state);
                }
                return;
            }
            if ( self.onBindFunc ) {
                self.onBindFunc(self.jsDoc, self.jsObject, self.state);
            }
        }
    }

    jsObject._cb = func;
    func(jsDoc.content, jsObject, null, null, LW.JsonOT.ObjectModified );
};

LW.Controller.ConditionController.prototype.unbind = function() {
    if ( this.jsObject ) {
        delete this.jsObject._cb;
        if ( this.onUnbindFunc ) {
            this.onUnbindFunc(this.jsDoc, this.jsObject, this.state);
        }
    }
    delete this.jsDoc;
    delete this.jsObject;
};

/**
 * ObjectController
 */

/**
 * @param is a list of keys. If an attribute with a name in keys of the bound object changes, then the updateFunc is called.
 *        If the keys are null, any change to the bound object triggers the updateFunc.
 */
LW.Controller.ObjectController = function(dom, updateFunc, clearFunc, keys) {
    this.updateFunc = updateFunc;
    this.clearFunc = clearFunc;
    this.keys = keys;
    this.state = { dom: dom };
};

LW.Controller.ObjectController.prototype.bind = function(jsDoc, jsObject) {
    this.unbind();
    this.jsDoc = jsDoc;
    this.jsObject = jsObject;
    
    var self = this;
    this.jsObject._cb = function(doc, obj, key, mutation, event) { 
        if ( self.keys ) {
            if ( event == LW.JsonOT.AttributeInserted || event == LW.JsonOT.AttributeDeleted || event == LW.JsonOT.AttributeModified ) {
                if ( self.keys.indexOf( key ) != -1 ) {
                    jsObject._dirty = true;
                }
            }
            else if ( event == LW.JsonOT.ObjectModified && jsObject._dirty ) {
                delete jsObject._dirty;
                self.updateFunc(jsDoc, jsObject, self.state);
            }       
        } else if ( event == LW.JsonOT.ObjectModified ) {
            self.updateFunc(jsDoc, jsObject, self.state);
        }
    }

    this.updateFunc(jsDoc, jsObject, this.state);
};

LW.Controller.ObjectController.prototype.unbind = function() {
    if ( !this.jsDoc ) {
        return;
    }
    if ( this.clearFunc ) {
        this.clearFunc(this.jsDoc, this.jsObject, this.state);
    }
    delete this.jsDoc;
    delete this.jsObject;
};

/**
 * Handles an array of objects and creates HTML for each object.
 * The array itself is assumed to be a field stored in some parent-object.
 * Even if the array is replaced, the controller will find out and act accordingly, i.e.
 * update the HTML.
 *
 * @param domInsertBefore may be null, then the list elements are appened to the dom element.
 * @param controllerFactory is a function that returns a new controller object.
 */
LW.Controller.ListController = function(dom, domInsertBefore, createFunc, deleteFunc) {
    this.state = { dom: dom };
    this.startChildrenIndex = 0;
    if ( domInsertBefore ) {
        for( var i = 0; i < dom.children.length; ++i ) {
            if ( dom.children[i] == domInsertBefore ) {
                break;
            }
            this.startChildrenIndex++;
        }
    } else {
        this.startChildrenIndex = dom.children.length;
    }
    this.createFunc = createFunc;
    this.deleteFunc = deleteFunc;
};

LW.Controller.ListController.prototype.bind = function(jsDoc, jsArr) {
    this.unbind();
    // An instance of LW.Doc
    this.jsDoc = jsDoc;
    this.jsArr = jsArr;
    this.states = [];

    var self = this;

    jsArr._cb_inserted = function(doc, arr, index, mutation, event) {
        self.states.splice(index, 0, { });
        self.createFunc(jsDoc, arr[index], self.states, index);
        self.state.dom.insertBefore( self.states[index].dom, self.state.dom.children[self.startChildrenIndex + index] );
    };
    jsArr._cb_deleted = function(doc, arr, index, mutation, event) {
        if ( self.deleteFunc ) {
            self.deleteFunc(jsDoc, arr[index], self.states, index);
        }
        self.states.splice(index, 1);
        self.state.dom.removeChild(self.states[index].dom);
    };
    // TODO lift and squeeze

    for( var i = 0; i < jsArr.length; ++i ) {
        this.states[i] = { };
        this.createFunc(jsDoc, jsArr[i], this.states, i);
        this.state.dom.insertBefore( this.states[i].dom, this.state.dom.children[this.startChildrenIndex + i] );
    }
};

LW.Controller.ListController.prototype.unbind = function() {
    if ( this.jsArr ) {
        delete this.jsArr._cb_inserted;
        delete this.jsArr._cb_deleted;
    }
    if ( this.states ) {
        for( var i = 0; i < this.states.length; ++i ) {
            if ( this.deleteFunc ) {
                this.deleteFunc(this.jsDoc, this.jsArr[i], this.states, i);
            }
            this.state.dom.removeChild(this.states[i].dom);
        }
        delete this.states;
    }
    delete this.jsDoc;
    delete this.jsArr;
};
