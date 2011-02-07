if ( !window.LW ) {
  LW = { };
}

LW.Social = {
};

LW.Social.init = function() {
    // Retrieve session info from the server
    LW.Rpc.enqueueGet("/_sessioninfo", LW.Social.initCallback_, function() { } );
};

LW.Social.initCallback_ = function(reply) {
    // Parse the reply
    var info = JSON.parse(reply);
    LW.Rpc.user = info.user;
    LW.Rpc.domain = info.domain;
    LW.Rpc.displayName = info.displayName;
    // Show which user is logged in
    document.getElementById("username").innerText = LW.Rpc.displayName;
    LW.Inbox.init();
    LW.Session.init();

    var renderInboxEntry = function(jsDoc, jsObject, state) {
        state.dom.innerHTML = '<span class="title">' + jsObject.authors + '</span> <span class="digest">' + jsObject.digest + '</span>';
        state.dom.onclick = function() {
            var newdoc = LW.Inbox.getOrCreateDoc(jsObject.uri);
            LW.Social.documentController.bind(newdoc, newdoc.content._data);
            LW.Social.participantsController.bind(newdoc, newdoc.content._meta);
            LW.Session.open(jsObject.uri);
        };
    };
    // Controller for the inbox
    var createEntry = function(jsDoc, jsObject, states, index) {
        var state = states[index];
        state.dom = document.createElement("div");
        state.dom.className = "inbox-entry";
        state.controller = new LW.Controller.ObjectController(state.dom, renderInboxEntry, null );
        state.controller.bind(jsDoc, jsObject);
    };
    var deleteEntry = function(jsDoc, jsObject, states, index) {
        var state = states[index];
        if ( state.controller ) {
            state.controller.unbind();
            delete state.controller;
        }
    };
    var docsBindFunc = function(jsDoc, jsObject, state) {
        state.controller = new LW.Controller.ListController( state.dom, null, createEntry, deleteEntry);
        state.controller.bind(jsDoc, jsObject);
    };
    LW.Social.inboxController = new LW.Controller.AttributeController(document.getElementById("inbox"), "docs", docsBindFunc);
    LW.Social.inboxController.bind(LW.Inbox.self, LW.Inbox.self.content._data);

    LW.Social.documentController = LW.Social.createConversationController(document.getElementById("document"));
    LW.Social.participantsController = LW.Social.createParticipantsController(document.getElementById("document-panel"));

    $("#add-participant-button").find("button").click( function() {
        var dlg = document.getElementById("dlg-add-participants");
        if ( dlg.style.visibility != "visible" ) {
            var callback = function(reply) {
                var friends = JSON.parse(reply).friends;
                var dom = document.getElementById("newfriends");
                // Render HTML
                var html = "";
                for( var i = 0; i < friends.length; ++i ) {
                    var contact = friends[i];
                    html += '<div class="friend">';
                    html += '<img class="friend-image" src="../images/unknown.png"><div><span class="friend-name">' + esc(contact.displayName) + '</span><br><span class="friend-id">' + esc(contact.userid) + '</span></div></div>';
                }
                dom.innerHTML = html;
                // Install event handlers
                for( var i = 0; i < friends.length; ++i ) {
                    var f = function(c) {
                        $(dom.children[i]).click( function() {
                            if ( !LW.Model.hasParticipant(LW.Social.documentController.jsDoc, c.userid) ) {
                                LW.Model.addParticipant(LW.Social.documentController.jsDoc, c);
                            }
                        });
                    };
                    f(friends[i]);
                }
            };
            dlg.style.visibility = "visible";
            LW.Rpc.enqueueGet("/client/" + LW.Rpc.domain + "/_user/" + LW.Rpc.user + "?kind=friends", callback);
        } else {
            dlg.style.visibility = "hidden";
        }
    });

    $('#logout').click( function() {
        LW.Rpc.logout = true;
        window.location.pathname = "/_logout";
    });
};

LW.Social.createParticipantsController = function(parentdom) {
    var updateParticipant = function(jsDoc, jsObject, state) {
        state.dom.src = "images/unknown.png";
        state.dom.alt = jsObject.displayName;
    };
    var createParticipant = function(jsDoc, jsObject, states, index) {
        var state = states[index];
        state.dom = document.createElement("img");
        state.dom.className = "participant";
        state.controller = new LW.Controller.ObjectController(state.dom, updateParticipant, null, null );
        state.controller.bind(jsDoc, jsObject);
    };
    var deleteParticipant = function(jsDoc, jsObject, states, index) {
        var state = states[index];
        if ( state.controller ) {
            state.controller.unbind();
            delete state.controller;
        }
    };
    var bindFunc = function(jsDoc, jsObject, state) {
        if ( !state.controller ) {
            state.controller = new LW.Controller.ListController( $(parentdom).children(".participants")[0], document.getElementById("add-participant-button"), createParticipant, deleteParticipant);
        }
        state.controller.bind(jsDoc, jsObject);
    };
    var unbindFunc = function(jsDoc, jsObject, state) {
        if ( state.controller ) {
            state.controller.unbind();
        }
    };
    return new LW.Controller.AttributeController(parentdom, "participants", bindFunc, unbindFunc);
};

LW.Social.createConversationController = function(parentdom) {
    var condition = function(jsDoc, jsObject, controller) {
        return jsObject.blips && jsObject.blips.length > 0;
    };
    var trueFunc = function(jsDoc, jsObject, state) {
        state.dom.style.display = "block";
        state.dom.className = "thread";
        state.dom.innerHTML = "";
        // state.dom.innerHTML = '<div class="thread-border"><div class="inline-reply-container"><img class="reply-image" src="../images/followup.png"><div class="inline-reply"></div></div></div>';

        if ( !state.controller ) {
            state.controller = LW.Social.createConversationController(state.dom);
        }
        state.controller.bind(jsDoc, jsObject);
    };
    var falseFunc = function(jsDoc, jsObject, state) {
        if ( state.controller ) {
            state.controller.unbind();
        }
        state.dom.style.display = "none";
    };
    var createThread = function(jsDoc, jsObject, states, index) {
        var state = states[index];
        state.dom = document.createElement("div");
        state.controller = new LW.Controller.ConditionController( state.dom, condition, trueFunc, falseFunc);
        state.controller.bind(jsDoc, jsObject);
    };
    var deleteThread = function(jsDoc, jsObject, states, index) {
        var state = states[index];
        if ( state.controller ) {
            state.controller.unbind();
            delete state.controller;
        }
    };
    var bindThreadsFunc = function(jsDoc, jsObject, state) {
        if ( !state.controller ) {
            state.controller = new LW.Controller.ListController( state.dom, null, createThread, deleteThread);
        }
        state.controller.bind(jsDoc, jsObject);
    };
    var unbindThreadsFunc = function(jsDoc, jsObject, state) {
        if ( state.controller ) {
            state.controller.unbind();
        }
    };    
    var updateBlip = function(jsDoc, jsObject, state) {
        var html = '<div class="border"></div>';
        html += '<div><img class="author" src="images/unknown.png"><div class="date">4:38 pm</div><div class="authorname">' + jsObject._meta.author + ': </div>'
        html += '<div class="editor" contentEditable="true"></div><div class="editor-buttons"><button><b>Done</b> <span style="color:#666">[Shift+Enter]</span></button></div></div>';
        html += '<div class="blip-border" style="position:relative; height:0px; clear:both;"><div class="inline-reply-container"><div class="inline-reply-image"> </div><div class="inline-reply"></div></div></div>';
        // html += '<div class="threads"></div>';
        state.dom.innerHTML = html;
        if ( !state.controller ) {
            $(state.dom).find(".inline-reply-container").click( function() {
                if ( state.dom.nextSibling ) {
                    LW.Model.createThreadAndBlip(jsDoc, jsDoc.url + "!" + jsObject._id, "Hallo Thread");
                } else {
                    LW.Model.createBlip(jsDoc, jsDoc.url + "!" + jsObject._parent._id, "Hallo Follow up");
                }
            });
            state.controller = new LW.Controller.AttributeController(state.dom, "threads", bindThreadsFunc, unbindThreadsFunc);
            state.controller.bind(jsDoc, jsObject);
            state.editor = new LW.Editor(jsDoc, jsObject.content, $(state.dom).find(".editor")[0]);
        }
    };
    var clearBlip = function(jsDoc, jsObject, state) {
        if ( state.controller ) {
            state.controller.unbind();
        }
    };
    var createBlip = function(jsDoc, jsObject, states, index) {
        var state = states[index]
        state.dom = document.createElement("div");
        state.dom.className = "blip clearfix";
        state.controller = new LW.Controller.ObjectController(state.dom, updateBlip, clearBlip, ["_meta"] );
        state.controller.bind(jsDoc, jsObject);
    };
    var deleteBlip = function(jsDoc, jsObject, states, index) {
        var state = states[index];
        if ( state.controller ) {
            state.controller.unbind();
        }
    };
    var bindBlipsFunc = function(jsDoc, jsObject, state) {
        // Wire events
        var reply = $(parentdom).children(".reply");
        if ( reply.length > 0 ) {
            reply.show();
            reply.bind( 'click', function() {
                LW.Model.createBlip(jsDoc, jsDoc.url + "!" + jsObject._id, "Hallo Welt");
            } );
        }
        // Create controller
        var dom = $(parentdom).children(".mainthread");
        if ( dom.length == 0 ) {
            dom = $(parentdom); // .children(".thread");
        }
        if ( !state.controller ) {
            state.controller = new LW.Controller.ListController( dom[0], null, createBlip, deleteBlip);
        }
        state.controller.bind(jsDoc, jsObject);
    };
    var unbindBlipsFunc = function(jsDoc, jsObject, state) {
        if ( state.controller ) {
            state.controller.unbind();
        }
        // Unwire events
        var reply = $(parentdom).children(".reply");
        if ( reply.length > 0 ) {
            reply.hide();
            reply.unbind('click');
        }
    };
    return new LW.Controller.AttributeController(parentdom, "blips", bindBlipsFunc, unbindBlipsFunc);
};

LW.Social.newDocument = function() {
    var newdoc = LW.Model.createDocument();
    LW.Social.documentController.bind(newdoc, newdoc.content._data);
    LW.Social.participantsController.bind(newdoc, newdoc.content._meta);
    LW.Session.open(newdoc.url);
};

// TODO
function esc(str) {
    return str;
}
