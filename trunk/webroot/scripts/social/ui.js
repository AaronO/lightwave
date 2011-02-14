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
    document.getElementById("username2").innerText = LW.Rpc.displayName;
    document.getElementById("userid").innerText = LW.Rpc.user + "@" + LW.Rpc.domain;
    LW.Inbox.init();
    LW.Session.init();

    var renderFriendRequest = function(jsDoc, jsObject, state) {
        var html = "";
        html += '<table style="font-size:10pt; width:100%">'
        html += '  <tbody>';
        html += '    <tr>';
        html += '      <td style="vertical-align:top; width:42px">';
        html += '        <img style="height:40px" src="images/unknown.png">';
        html += '      </td>';
        html += '      <td>';
        html += '        <div style="text-align:right; float:right; margin-top:3px"><span style="color:#aaa">12:48 pm</span></div>';
        html += '        <div style="margin-bottom:2px"><span style="color:#3b5998">' + jsObject.value.author + ' </span>';
        html += '        ' + jsObject.value.text + '<br>';
        if ( jsObject.value.request.userid == LW.Rpc.user + "@" + LW.Rpc.domain ) {
            html += '<button>Cancel</button>';
        } else if ( jsObject.value.response.userid == LW.Rpc.user + "@" + LW.Rpc.domain ) {
            html += '<button>Confirm</button> <button>Not Now</button>';
        }
        html += '        </div>';
        html += '      </td>';
        html += '    </tr>';
        html += '  </tbody>';
        html += '</table>';
        html += '<div style="margin-top:3px; margin-bottom:1px" class="blip-digest-border"></div>';
        state.dom.innerHTML = html;
    };

    var renderInboxEntry = function(jsDoc, jsObject, state) {
        if ( jsObject.value.schema == "//lightwave/friend-request" ) {
            renderFriendRequest(jsDoc, jsObject, state);
            return;
        }
        // state.dom.innerHTML = '<span class="title">' + jsObject.value + '</span>';
        var html = "";
        html += '<table style="font-size:10pt; width:100%">'
        html += '  <tbody>';
        html += '    <tr>';
        html += '      <td style="vertical-align:top; width:42px">';
        html += '        <img style="height:40px" src="images/unknown.png">';
        html += '      </td>';
        html += '      <td>';
        html += '        <div style="text-align:right; float:right; margin-top:3px"><span style="color:#aaa">12:48 pm</span><br><span style="color:#3b5998">Like</span></div>';
        html += '        <div style="color:#3b5998; margin-bottom:2px">' + jsObject.value.author + '</div>';
        html += '        <div>' + jsObject.value.text + '</div>';
//        html += '        <div style="margin-top:3px"><span style="color:#aaa">12:48 pm</span> <span style="color:#3b5998">Like</span> <span style="color:#3b5998">Comment</span></div>';
        if ( jsObject.value.comments.length > 0 ) {
            html += '<div class="thread-digest">';
            if ( jsObject.value.blipCount > 1 + jsObject.value.comments.length ) {
                html += '<div class="blip-digest">';
                html += '  <div class="blip-digest-border"></div>';
                html += '  <span style="color:#3b5998">View all ' + jsObject.value.blipCount.toString() + ' comments</span>';
                html += '</div>';
            }
            if ( jsObject.value.likes ) {
                html += '<div class="blip-digest">';
                html += '  <div class="blip-digest-border"></div>';
                html += '  <span style="color:#3b5998">Torben</span> and 2 people like this';
                html += '</div>';
            }
            for ( var i = 0; i < jsObject.value.comments.length; ++i ) {
                var c = jsObject.value.comments[i];
                html += '<div class="blip-digest">';
                html += '  <div class="blip-digest-border"></div>';
                html += '  <table style="font-size:10pt; margin:0px">';
                html += '    <tbody>';
                html += '      <tr>';
                html += '        <td style="vertical-align:top">';
                html += '          <img style="height:27px" src="../images/unknown.png">';
                html += '        </td>';
                html += '        <td>';
                html += '          <div><span style="color:#3b5998; margin-bottom:2px">' + c.author + ' </span>' + c.text + '</div>';
                html += '          <div><span style="color:#aaa">12:48 pm</span> <span style="color:#3b5998">Like</span></div>';
                html += '        </td>';
                html += '      </tr>';
                html += '    </tbody>';
                html += '  </table>';
                html += '</div>';
            }
        }
        html += '      </td>';
        html += '    </tr>';
        html += '  </tbody>';
        html += '</table>';
        html += '<div style="margin-top:3px; margin-bottom:1px" class="blip-digest-border"></div>';
        state.dom.innerHTML = html;
        state.dom.onclick = function() {
            $("#friends-panel").hide();
            $("#info-panel").hide();
            $("#home-panel").hide();
            $("#show-home")[0].className = "navi-link";
            $("#show-info")[0].className = "navi-link";
            $("#show-friends")[0].className = "navi-link";
            var newdoc = LW.Inbox.getOrCreateDoc(jsObject.key);
            LW.Social.documentController.bind(newdoc, newdoc.content._data);
            LW.Social.participantsController.bind(newdoc, newdoc.content._meta);
            LW.Session.open(jsObject.key, true);
            $("#document-panel").show();
        };
    };
    // Controller for the inbox
    var createEntry = function(jsDoc, jsObject, states, index) {
        var state = states[index];
        state.dom = document.createElement("div");
        // state.dom.className = "inbox-entry";
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
    LW.Social.inboxController = new LW.Controller.AttributeController(document.getElementById("inbox"), "items", docsBindFunc);
    var inbox = LW.Inbox.getOrCreateDoc("/_view/inbox")
    LW.Social.inboxController.bind(inbox, inbox.content._data);

    var friends = LW.Session.openView("friends", 'map=digest&with=friend:' + LW.Rpc.user  + "@" + LW.Rpc.domain);
    LW.Social.friendsController = LW.Social.createFriendsController(document.getElementById("friends"));
    LW.Social.friendsController.bind(friends, friends.content._data);

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
    $("#show-share-status").click( function() {
        $("#share-status").show();
        $("#share-link").hide();
    });
    $("#show-share-link").click( function() {
        $("#share-link").show();
        $("#share-status").hide();
    });
    $("#share-status-edit").focus( function() {
        if ( this.innerText == "What's on your mind?" ) {
            this.innerHTML = "";
            this.style.color = "black";
            $("#share-status-button").show();
        }
    });
    $("#share-status-edit").blur( function() {
        if ( !this.innerText || trim(this.innerText) == "" ) {
            this.innerHTML = "What's on your mind?"
            this.style.color = "#999";
            $("#share-status-button").hide();
        }
    });
    $("#share-status-button").click( function() {
        var edit = $("#share-status-edit")[0]
        if ( edit.innerText && trim(edit.innerText) != "" ) {
            var newdoc = LW.Model.createDocument(trim(edit.innerText))
            LW.Session.open(newdoc.url, false);
            edit.innerHTML = "What's on your mind?"
            edit.style.color = "#999";
            $("#share-status-button").hide();
        }
    });
    $("#share-link-edit").focus( function() {
        if ( this.innerText == "http://" ) {
            this.innerHTML = "";
            this.style.color = "black";
            $("#share-link-button").show();
        }
    });
    $("#share-link-edit").blur( function() {
        if ( !this.innerText || trim(this.innerText) == "" ) {
            this.innerHTML = "http://"
            this.style.color = "#999";
            $("#share-link-button").hide();
        }
    });
    $("#show-home").click( function() {
        $("#document-panel").hide();
        $("#friends-panel").hide();
        $("#info-panel").hide();
        $("#home-panel").show();
        $("#show-home")[0].className = "selected-navi-link";
        $("#show-info")[0].className = "navi-link";
        $("#show-friends")[0].className = "navi-link";
    });
    $("#show-info").click( function() {
        $("#document-panel").hide();
        $("#friends-panel").hide();
        $("#info-panel").show();
        $("#home-panel").hide();
        $("#show-home")[0].className = "navi-link";
        $("#show-info")[0].className = "selected-navi-link";
        $("#show-friends")[0].className = "navi-link";
    });
    $("#show-friends").click( LW.Social.showFriends );
};

LW.Social.showFriends = function() {
    $("#document-panel").hide();
    $("#friends-panel").show();
    $("#info-panel").hide();
    $("#home-panel").hide();
    $("#show-home")[0].className = "navi-link";
    $("#show-info")[0].className = "navi-link";
    $("#show-friends")[0].className = "selected-navi-link";
    if ( !LW.Social.peopleView ) {
        LW.Social.peopleView = LW.Session.openView("people", 'map=digest&with=schema://lightwave/user');
        LW.Social.peopleController = LW.Social.createPeopleController(document.getElementById("people"));
        LW.Social.peopleController.bind(LW.Social.peopleView, LW.Social.peopleView.content._data);
    }
};

LW.Social.createPeopleController = function(parentdom) {
    var updateFriend = function(jsDoc, jsObject, state) {
        var html = '<td style="width:200px"><div class="friend-large"><img style="float:left" src="../images/unknown.png"><span style="color:#3b5998">' + jsObject.value.displayName + '</span></br><span style="color:#666">' + jsObject.value.userid + '</span></div></td>';
        html += '<td><button class="add-friend-button">+1 Add as Friend</button></td>'
        state.dom.innerHTML = html;
        state.dom.alt = jsObject.displayName;
        $(state.dom).find(".add-friend-button").click( function() {
            LW.Model.createFriendRequest(jsObject.value.userid);
        });
    };
    var createFriend = function(jsDoc, jsObject, states, index) {
        var state = states[index];
        state.dom = document.createElement("tr");
        state.controller = new LW.Controller.ObjectController(state.dom, updateFriend, null, null );
        state.controller.bind(jsDoc, jsObject);
    };
    var deleteFriend = function(jsDoc, jsObject, states, index) {
        var state = states[index];
        if ( state.controller ) {
            state.controller.unbind();
            delete state.controller;
        }
    };
    var bindFunc = function(jsDoc, jsObject, state) {
        if ( !state.controller ) {
            state.controller = new LW.Controller.ListController( parentdom, null, createFriend, deleteFriend);
        }
        state.controller.bind(jsDoc, jsObject);
    };
    var unbindFunc = function(jsDoc, jsObject, state) {
        if ( state.controller ) {
            state.controller.unbind();
        }
    };
    return new LW.Controller.AttributeController(parentdom, "items", bindFunc, unbindFunc);
};

LW.Social.createFriendsController = function(parentdom) {
    var updateFriend = function(jsDoc, jsObject, state) {
        var html = '<img style="float:left" src="../images/unknown.png"><span style="color:#3b5998">' + jsObject.value.displayName + '</span></br><span style="color:#666">' + jsObject.value.userid + '</span>';
        state.dom.innerHTML = html;
        state.dom.alt = jsObject.displayName;
    };
    var createFriend = function(jsDoc, jsObject, states, index) {
        var state = states[index];
        state.dom = document.createElement("div");
        state.dom.className = "clearfix friend-large";
        state.controller = new LW.Controller.ObjectController(state.dom, updateFriend, null, null );
        state.controller.bind(jsDoc, jsObject);
    };
    var deleteFriend = function(jsDoc, jsObject, states, index) {
        var state = states[index];
        if ( state.controller ) {
            state.controller.unbind();
            delete state.controller;
        }
    };
    var bindFunc = function(jsDoc, jsObject, state) {
        if ( !state.controller ) {
            state.controller = new LW.Controller.ListController( parentdom, null, createFriend, deleteFriend);
        }
        state.controller.bind(jsDoc, jsObject);
    };
    var unbindFunc = function(jsDoc, jsObject, state) {
        if ( state.controller ) {
            state.controller.unbind();
        }
    };
    return new LW.Controller.AttributeController(parentdom, "items", bindFunc, unbindFunc);
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

/*
LW.Social.newDocument = function() {
    var newdoc = LW.Model.createDocument();
    LW.Social.documentController.bind(newdoc, newdoc.content._data);
    LW.Social.participantsController.bind(newdoc, newdoc.content._meta);
    LW.Session.open(newdoc.url, false);
};
*/

// TODO
function esc(str) {
    return str;
}

function trim(str) {
  return str.replace (/^\s+/, '').replace (/\s+$/, '');
}
