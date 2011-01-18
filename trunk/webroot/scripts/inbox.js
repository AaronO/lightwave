if ( !window.LW ) {
    LW = { };
}

// -------------------------------------------------------
// Inbox

LW.Inbox = {
    docs_ : {},
    self : null
};

LW.Inbox.uniqueId = function() {
    var str = Math.random().toString();
    return str.substr(2, str.length - 2);
};

LW.Inbox.init = function() {
    // Create the json document that contains the inbox
    LW.Inbox.self = LW.Inbox.getOrCreateDoc("/" + LW.Rpc.domain + "/_user/" + LW.Rpc.user + "/inbox");
};

// @param url has the format "/host-name/conversation-id"
LW.Inbox.getOrCreateDoc = function(url) {
    if ( LW.Inbox.docs_[url] ) {
	return LW.Inbox.docs_[url];
    }
    
    var d = new LW.Doc(url);
    LW.Inbox.docs_[url] = d;
    return d;
};

// The id has the form "/host-name/conversation-id!object-id" where conversation-id has the form (/{identifier})*
LW.Inbox.getElementById = function(id) {
    var i = id.indexOf("!");
    var convid;
    var objectid;
    if ( i == -1 ) {
	convid = id;
    } else {
	convid = id.substr(0, i);
	objectid = id.substring(i + 1, id.length);
    }
    var d = LW.Inbox.docs_[convid];
    if ( !objectid ) {
	return d;
    }
    return d.getElementById(objectid);
};

