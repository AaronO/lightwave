if ( !window.LW ) {
    LW = { };
}

/**
var update = function(dom, doc, obj) {
   dom.innerHTML = foo(obj)l
};
var clear = function(dom, doc, obj) {
   dom.innerHTML = "";
};
var objfactory = function() {
   var dom = document.createElement("div");
   return new LW.Controller.ObjectController(dom, update, clear);
};
var arrfactory = funtion() {
  return new LW.Controller.ListController( document.getElementById("foo"), null, objfactory);
};
var attr = new LW.Controller.AttributeController(arrfactory);
attr.bind(doc, obj, "comments");

**/

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
LW.Controller.AttributeController = function(controllerFactory) {
    this.controllerFactory = controllerFactory;
};

LW.Controller.AttributeController.prototype.bind = function(jsDoc, jsParentObj, jsParentKey) {
    this.unbind();
    this.jsParentObj = jsParentObj;
    this.jsParentKey = jsParentKey;

    self = this;
    var func = function(doc, obj, key, mutation, event) {
        if ( event == LW.JsonOT.AttributeInserted ) {
            if ( this.controller ) {
                this.controller.unbind();
            }
            self.controller = self.controllerFactory();
            self.controller.bind(jsDoc, obj[key]);
        }
        else if ( event == LW.JsonOT.AttributeDeleted ) {
            if ( this.controller ) {
                this.controller.unbind();
            }
        }
    }

    jsParentObj["_cb_" + jsParentKey] = func;
    if ( jsParentObj[jsParentKey] ) {
        func(jsDoc.content, jsParentObj, jsParentKey, null, LW.JsonOT.AttributeInserted );
    }
};

LW.Controller.AttributeController.prototype.unbind = function() {
    if ( this.jsParentObj ) {
        this.jsParentObj["_cb_" + this.jsParentKey];
        if ( this.controller ) {
            this.controller.unbind();
            delete this.controller;
        }
    }
    delete this.jsParentObj;
    delete this.jsParentKey;
};

LW.Controller.AttributeController.prototype.getDom = function() {
    if ( this.controller ) {
        return this.controller.getDom();
    }
    return null;
};

/**
 * ObjectController
 */

/**
 * @param is a list of keys. If an attribute with a name in keys of the bound object changes, then the updateFunc is called.
 *        If the keys are null, any change to the bound object triggers the updateFunc.
 */
LW.Controller.ObjectController = function(dom, updateFunc, clearFunc, keys) {
    this.dom = dom;
    this.updateFunc = updateFunc;
    this.clearFunc = clearFunc;
    this.keys = keys;
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
                self.updateFunc(self.dom, jsDoc, jsObject);
            }       
        } else if ( event == LW.JsonOT.ObjectModified ) {
            self.updateFunc(self.dom, jsDoc, jsObject);
        }
    }

    this.updateFunc(this.dom, jsDoc, jsObject);
};

LW.Controller.ObjectController.prototype.unbind = function() {
    if ( !this.jsDoc ) {
        return;
    }
    if ( this.clearFunc ) {
        this.clearFunc(this.dom, this.jsDoc, this.jsObject);
    }
    delete this.jsDoc;
    delete this.jsObject;
};

LW.Controller.ObjectController.prototype.getDom = function() {
    return this.dom;
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
LW.Controller.ListController = function(dom, domInsertBefore, controllerFactory) {
    this.dom = dom;
    this.startChildrenIndex = 0;
    if ( domInsertBefore ) {
        for( var i = 0; i < dom.children.length; ++i ) {
            if ( dom.children[i] == domInsertBefore ) {
                break;
            }
            this.startChildrenIndex++;
        }
    } else {
        this.startChildrenIndex = this.dom.children.length;
    }
    this.controllerFactory = controllerFactory;
};

LW.Controller.ListController.prototype.bind = function(jsDoc, jsArr) {
    this.unbind();
    // An instance of LW.Doc
    this.jsDoc = jsDoc;
    this.jsArr = jsArr;
    this.controllers = [];

    var self = this;

    jsArr._cb_inserted = function(doc, arr, index, mutation, event) {
        var c = self.controllerFactory();
        c.bind( jsDoc, arr[index] );
        self.dom.insertBefore( c.getDom(), self.dom.children[self.startChildrenIndex + index] );
        self.controllers.splice(index, 0, c);
    };
    jsArr._cb_deleted = function(doc, arr, index, mutation, event) {
        self.dom.removeChild(self.controllers[index].getDom());
        self.controllers[index].unbind();
    };
    // TODO lift and squeeze

    for( var i = 0; i < jsArr.length; ++i ) {
        this.controllers[i] = this.controllerFactory();
        this.controllers[i].bind( jsDoc, jsArr[i] );
        this.dom.insertBefore( this.controllers[i].getDom(), this.dom.children[this.startChildrenIndex + i] );
    }
};

LW.Controller.ListController.prototype.unbind = function() {
    if ( this.jsArr ) {
        delete this.jsArr._cb_inserted;
        delete this.jsArr._cb_deleted;
        delete this.jsArr;
    }
    if ( this.controllers ) {
        for( var i = 0; i < this.controllers.length; ++i ) {
            this.dom.removeChild(this.controllers[i].getDom());
            this.controllers[i].unbind();
        }
        delete this.controllers;
    }
};

LW.Controller.ListController.prototype.getDom = function() {
    return this.dom;
};
