LW = {
};

LW.Rpc = {
  // TODO
  user : "weis",
  queue : [ ]
};

LW.Rpc.post = function(url, jsonData, callback, errCallback) {
  LW.Rpc.postOrGet_(url, "POST", jsonData, callback, errCallback);
}

LW.Rpc.get = function(url, callback, errCallback) {
  LW.Rpc.postOrGet_(url, "GET", null, callback, errCallback);
}

LW.Rpc.postOrGet_ = function(url, httpMethod, jsonData, callback, errCallback) {
  var xmlHttp = null;
  try {
	// Mozilla, Opera, Safari sowie Internet Explorer (ab v7)
	xmlHttp = new XMLHttpRequest();
  } 
  catch(e) {
	try {
	  // MS Internet Explorer (ab v6)
	  xmlHttp  = new ActiveXObject("Microsoft.XMLHTTP");
	} catch(e) {
	  try {
		// MS Internet Explorer (ab v5)
		xmlHttp  = new ActiveXObject("Msxml2.XMLHTTP");
	  } catch(e) {
		xmlHttp  = null;
	  }
	}
  }
  if (xmlHttp) {
	xmlHttp.open(httpMethod, url, true);
	xmlHttp.setRequestHeader("Content-type", "application/json")
	xmlHttp.onreadystatechange = function () {
	  if (xmlHttp.readyState == 4) {
		if( callback && xmlHttp.status == 200 ) {
		  callback(xmlHttp.responseText);
		} else if ( errCallback && xmlHttp.status != 200 ) {
		  errCallback();
		}
	  }
	};
	if ( jsonData ) {
	  xmlHttp.send(jsonData);
	} else {
	  xmlHttp.send();
	}
	return true;
  }
  return false;
};

LW.Rpc.enqueue = function(url, jsonData, callback, errCallback) {
  LW.Rpc.queue.push( { url:url, data:jsonData, callback:callback, errCallback:errCallback } );
  if ( LW.Rpc.queue.length == 1 ) {
	LW.Rpc.post(url, jsonData, LW.Rpc.enqueueCallback_, LW.Rpc.enqueueErrCallback_);
  }
};

LW.Rpc.enqueueCallback_ = function(reply) {
  var msg = LW.Rpc.queue[0];
  if ( msg.callback ) {
	msg.callback(reply);
  }
  LW.Rpc.queue.splice(0,1);
  // Send more?
  if ( LW.Rpc.queue.length > 0 ) {
	var msg = LW.Rpc.queue[0];
	LW.Rpc.post(msg.url, msg.data, LW.Rpc.enqueueCallback_, LW.Rpc.enqueueErrCallback_);
  }
};

LW.Rpc.enqueueErrCallback_ = function(reply) {
  var msg = LW.Rpc.queue[0];
  if ( msg.errCallback ) {
	msg.errCallback(reply);
  }
  alert("Offline", "You seem to be offline");
};

// -------------------------------------------------------------
// Session

LW.Session = {
  id : null,
  sessionCreated : false,
  version : 0
};

LW.Session.init = function() {
  if ( !LW.Session.sessionCreated ) {
	LW.Session.createSession_();
  } else {
	LW.Session.sessionPoll_();
  }
};

LW.Session.createSession_ = function() {
  LW.Session.id = "s" + Math.random().toString();
  var json = {"_rev":0, "_data":{"filters":{}}};
  LW.Rpc.post("/client/_session/" + LW.Rpc.user + "/" + LW.Session.id, JSON.stringify(json), LW.Session.createSessionCallback_, LW.Session.createSessionErrCallback_);
};

LW.Session.createSessionCallback_ = function(reply) {
  console.log("create session: " + reply)
  var json = JSON.parse(reply)
  if ( json.ok == true ) {
	LW.Session.sessionCreated = true;
	LW.Session.version = json.version;
	// Open the inbox
	// TODO
	// LW.Session.open("/localhost/foo")
	LW.Session.sessionPoll_();
  } else {
	alert("Failed to create a session");
  }
};

LW.Session.createSessionErrCallback_ = function() {
  alert("Offline", "You seem to be offline");
};
	
LW.Session.sessionPoll_ = function() {
  LW.Rpc.get("/client/_session/" + LW.Rpc.user + "/" + LW.Session.id + "/_poll", LW.Session.sessionPollCallback_, LW.Session.sessionPollErrCallback_);  
};

LW.Session.sessionPollCallback_ = function(reply) {
  console.log("session: " + reply)
  var json = JSON.parse(reply);
  for( var url in json ) {
	var doc = LW.Inbox.getOrCreateDoc(url);
	var mutations = json[url];
	for( var i in mutations ) {
	  doc.recvDocMutation(mutations[i]);
	}
  }
  LW.Session.sessionPoll_();
};

LW.Session.sessionPollErrCallback_ = function() {
  alert("Offline", "You seem to be offline");
};
	
LW.Session.open = function(prefix, snapshot, recursive, mimeTypes, schemas) {
  if ( !mimeTypes ) {
	mimeTypes = [];
  }
  if ( !schemas ) {
	schemas = [];
  }
  var json = { "_rev":LW.Session.version, "_data":{ "$object":true, "filters":{"$object":true} } };
  json._data.filters[prefix] = {"recursive":recursive, "mimeTypes":mimeTypes, "schemas":schemas, "snapshot":snapshot};
  console.log("-> " + JSON.stringify(json))
  // LW.Rpc.post("/client/_session/" + LW.Rpc.user + "/" + LW.Session.id, JSON.stringify(json), LW.Session.openCallback_, LW.Session.openErrCallback_);
  LW.Rpc.enqueue("/client/_session/" + LW.Rpc.user + "/" + LW.Session.id, JSON.stringify(json), LW.Session.openCallback_);
};

LW.Session.openCallback_ = function(reply) {
  console.log("open: " + reply)
  var json = JSON.parse(reply)
  if ( json.ok == true ) {
	LW.Session.version = json.version;
  } else {
	alert("Failed to open a document");
  }  
};
	
/*
LW.Session.openErrCallback_ = function() {
  alert("Offline", "You seem to be offline");
};
*/
