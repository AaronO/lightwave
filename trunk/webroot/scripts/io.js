LW = {
};

LW.Rpc = {
  // TODO
  user : "weis"
};

LW.Rpc.post = function(url, jsonData, callback, errCallback) {
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
	xmlHttp.open('POST', url, true);
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
	xmlHttp.send(jsonData);
	return true;
  }
  return false;
};

LW.Rpc.get = function(url, callback, errCallback) {
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
	xmlHttp.open('GET', url, true);
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
	xmlHttp.send();
	return true;
  }
  return false;
};

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
	LW.Session.open("/localhost/foo")
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
  var json = JSON.stringify(reply);
  // TODO ...
  LW.Session.sessionPoll_();
};

LW.Session.sessionPollErrCallback_ = function() {
  alert("Offline", "You seem to be offline");
};
	
LW.Session.open = function(prefix, recursive, mimeTypes, schemas) {
  if ( !mimeTypes ) {
	mimeTypes = [];
  }
  if ( !schemas ) {
	schemas = [];
  }
  var json = { "_rev":LW.Session.version, "_data":{ "$object":true, "filters":{"$object":true} } };
  json._data.filters[prefix] = {"recursive":recursive, "mimeTypes":mimeTypes, "schemas":schemas};
  console.log("-> " + JSON.stringify(json))
  LW.Rpc.post("/client/_session/" + LW.Rpc.user + "/" + LW.Session.id, JSON.stringify(json), LW.Session.openCallback_, LW.Session.openErrCallback_);
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
	
LW.Session.openErrCallback_ = function() {
  alert("Offline", "You seem to be offline");
};
