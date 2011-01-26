/*
 * Author: Kai Chang, Torben Weis
 * 
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 *
 */

if ( !window.LW ) {
  LW = { };
}

LW.Tensor = {
    // Points to an instance of LW.Doc
    currentDoc : null,
    participantsViewController : null
};

// Called when a new column is shown.
// This includes the case that the user clicked on the inbox and a new document is shown
LW.Tensor.createColumnContent_ = function(id, list, isNewWave) {
    var comments = null;
    // Clicked on a conversation in the inbox? -> Install event handlers
    if ( id.indexOf('!') == -1 ) {
        // This is now the current document
        LW.Tensor.currentDoc = LW.Inbox.getOrCreateDoc(id);
        // When the title changes update the first comment because it contains the title
        LW.Tensor.currentDoc.content._data._cb_title = function(doc, obj, key, mutations, event) {
            if ( LW.Tensor.currentDoc.content == doc ) {
                if ( list.children.length > 0 && list.firstChild.firstChild.className != "editor" ) {
                    LW.Tensor.commentModifiedCallback_(LW.Tensor.currentDoc.content._data.comments, 0);
                }
            }
        };
        // Render Participants
        if ( !LW.Tensor.participantsViewController ) {
            var objfactory = function() {
                var dom = document.createElement("li");
                dom.className = "user";
                return new LW.Controller.ObjectController(dom, LW.Tensor.renderParticipant, null );
            };
            var arrfactory = function() {
                return new LW.Controller.ListController( document.getElementById("meta-bar"), null, objfactory);
            };
            LW.Tensor.participantsViewController = new LW.Controller.AttributeController(arrfactory);
        }
        LW.Tensor.participantsViewController.bind(LW.Tensor.currentDoc, LW.Tensor.currentDoc.content._meta, "participants");
        // Update the list of contacts that could be added to this wave
        LW.Tensor.updateAddParticipants();
        // When a new list of top-level comments is inserted -> register event handlers
        LW.Tensor.currentDoc.content._data._cb_comments = function(doc, obj, key, mutations, event) {
            if ( LW.Tensor.currentDoc.content == doc ) {
                if ( event == LW.JsonOT.AttributeInserted ) {
                    list.objectid = LW.Tensor.currentDoc.url + "!" + doc._data.comments._id;
                    doc._data.comments._cb_inserted = LW.Tensor.commentInsertedCallback_;
                }
            }
        };
        comments = LW.Tensor.currentDoc.content._data.comments;
    } else {
        comments = LW.Inbox.getElementById(id).comments;
    }
    // Clear the HTML for this columns
    list.innerHTML = "";
    // Is the document already loaded? -> Show it
    if ( comments ) {
        // Generate the HTML for the column
        list.objectid = LW.Tensor.currentDoc.url + "!" + comments._id;
        for( var i = 0; i < comments.length; i++ ) {
            LW.Tensor.commentModifiedCallback_(comments, i)
        }
        comments._cb_inserted = LW.Tensor.commentInsertedCallback_;
    }
    // Create the HTML for the edit box that allows users to type new comments
    var box = document.createElement("div");
    var html = '<div class="editor">';
    html += '<div class="title"><span class="label">Title:</span> <input type="text" class="titleInput"></input></div>';
    html += '<div><textarea></textarea></div>';
    html += '<div><span><a href="#">Submit</a> <a href="#">Cancel</a></span></div>';
    html += '</div>';
    html += '<div class="info"><a href="#">Click here to write a message</a></div>';
    box.innerHTML = html;
    box.className = "infoBox";
    // Event handlers for the edit box.
    box.lastChild.firstChild.onclick = function() {
        box.className = "inputBox";
        if ( LW.Tensor.currentDoc.content._data.comments && LW.Tensor.currentDoc.content._data.comments.length == 0 ) {
            box.firstChild.firstChild.lastChild.focus();
        } else {
            var textarea = box.firstChild.children[1].firstChild;
            textarea.focus();
        }
        return false;
    };
    // Cancel clicked
    box.firstChild.lastChild.firstChild.lastChild.onclick = function() {
        box.className = "infoBox";
        return false;
    };
    // Submit clicked
    box.firstChild.lastChild.firstChild.firstChild.onclick = function() {
        if ( LW.Tensor.currentDoc.url == id && LW.Tensor.currentDoc.content._data.comments && LW.Tensor.currentDoc.content._data.comments.length == 0 ) {
            var title = box.firstChild.firstChild.lastChild;
            LW.Model.setTitle(LW.Tensor.currentDoc, title.value);
            title.value = "";
        }
        var textarea = box.firstChild.children[1].firstChild;
        LW.Model.createNewComment(LW.Tensor.currentDoc, list.objectid, textarea.value);
        textarea.value = "";
        textarea.focus();
        return false;
    };
    list.appendChild(box);
    // If the user just created a new document then show the input box
    if ( isNewWave ) {
        box.className = "inputBox";
        box.firstChild.firstChild.lastChild.focus();
    }
};

LW.Tensor.deselectAll_ = function(list) {
    if ( LW.Tensor.participantsViewController ) {
        LW.Tensor.participantsViewController.unbind();
    }
    LW.Tensor.currentDoc = null;
    var content = $('#content');
    content.animate({left : '+=' + content.position().left}, 240);
    var i = 1;
    while( true ) {
        var list = $('#list-' + i.toString());
        if ( list.length == 0 ) {
            break;
        }
        list.removeClass('grey');
        list.children('.selected').removeClass('selected');
        if ( i > 1 ) {
            list.fadeOut(0);
        }
        i++;
    }
};

LW.Tensor.deselect_ = function(nextlist, nextnextlist, list, selected) {
    // Deselecting a document in the inbox -> do not show participants any more
    if ( list[0].id == "list-1" && LW.Tensor.participantsViewController ) {
        LW.Tensor.participantsViewController.unbind();
    }
    nextnextlist.fadeOut();
    list.removeClass('grey');
    selected.removeClass('selected');
    nextlist.fadeOut(240, function() {
        nextlist.removeClass('grey');
        nextlist.children('.selected').removeClass('selected');
    });
};

LW.Tensor.select_ = function(nextlist, list, selected, item) {
  list.addClass('grey');
  selected.removeClass('selected'); 
  item.addClass('selected');
  $('#container').scrollTop(0);
    LW.Tensor.createColumnContent_(item.context.id, nextlist.toArray()[0], false);
  nextlist.fadeIn(240);
};

LW.Tensor.reload_ = function(nextlist, nextnextlist, selected, item) {
  nextnextlist.fadeOut(240);
  selected.removeClass('selected'); 
  item.addClass('selected');
  nextlist.fadeOut(240, function() {
    nextlist.children('.selected').removeClass('selected');
    nextlist.removeClass('grey');
      LW.Tensor.createColumnContent_(item.context.id, nextlist.toArray()[0], false);
  } );
  $('html').scrollTop(0);
  nextlist.fadeIn();
};

LW.Tensor.shiftleft_ = function(col, nextlist) {
  nextcol  = nextlist.parent();
  colwidth = col.width();
  $('#leftcol').attr('id', '');
  $('#midcol').attr('id', 'leftcol');
  $('#rightcol').attr('id', 'midcol');
  nextcol.attr('id', 'rightcol');
  $('#content').animate({left : '-=' + colwidth}, 340);
};

LW.Tensor.shiftright_ = function(col, prevlist) {
  prevcol  = prevlist.parent();
  colwidth = col.width();
  $('#rightcol').attr('id', '');
  $('#midcol').attr('id', 'rightcol');
  $('#leftcol').attr('id', 'midcol');
  prevcol.attr('id', 'leftcol');
  $('#content').animate({left : '+=' + colwidth}, 340);
};

// Called when a user clicks on an entry in the inbox or a comment.
// "this" is the DIV element the user clicked on.
LW.Tensor.onDivClick_ = function() {
  // Do nothing while the element is being edited
  if ( $(this).find(".editBox").length > 0 ) {
    return;
  }
  
  list     = $(this).parent();
  col      = list.parent();
  selected = list.children('.selected');

  numlist  = parseInt(list.attr('id').substring(5));
  
  nextnum  =  numlist + 1;
  nextlist = $('#list-' + nextnum);
  nextnextnum =  numlist + 2;
  nextnextlist = $('#list-' + nextnextnum);
    
  if ( numlist > 1 ) {
    prevnum  = numlist - 1;
    prevlist = $('#list-' + prevnum);
    if ( numlist > 2 ) {
      prevprevnum  = numlist - 2;
      prevprevlist = $('#list-' + prevprevnum);
    }
  }

  if ( col.attr('id') == 'rightcol' ) {
    LW.Tensor.shiftleft_(col, nextlist);
    LW.Tensor.select_(nextlist, list, selected, $(this));
  }
  else if ( col.attr('id') == 'midcol' ) {
    if ( $(this).hasClass('selected') ) {
      LW.Tensor.deselect_(nextlist, nextnextlist, list, selected);
      if ( numlist > 2) {
        LW.Tensor.shiftright_(col, prevprevlist);
      }
    } else if ( list.children().hasClass('selected') ) {
      LW.Tensor.reload_(nextlist, nextnextlist, selected, $(this));
    } else {
      LW.Tensor.select_(nextlist, list, selected, $(this));
    }
  }
  else if ( col.attr('id') == 'leftcol' ) {
    if ( $(this).hasClass('selected') ) {
      LW.Tensor.deselect_(nextlist, nextnextlist, list, selected);
      if ( numlist > 1) {
        LW.Tensor.shiftright_(col, prevlist);
      }
    } else if ( list.children().hasClass('selected') ) {
      LW.Tensor.reload_(nextlist, nextnextlist, selected, $(this));
      if ( numlist > 1) {
        LW.Tensor.shiftright_(col, prevlist);
      }
    } else {
      LW.Tensor.select_(nextlist, list, selected, $(this));
    }
  }
  
  if ( numlist == 1 && $(this).hasClass('selected') ) {
    LW.Session.open(this.id, true, false);
  }
  return false;
};

// Install the event handlers for a comment that has been inserted in an array at the specified index.
LW.Tensor.commentInsertedCallback_ = function(doc, arr, index, mut, event) {
  arr[index]._cb = function(d, obj, key, mut, event) {
    if ( event == LW.JsonOT.ObjectModified ) {
      LW.Tensor.commentModifiedCallback_(arr, index);
    } else if ( event == LW.JsonOT.AttributeInserted && key == "comments" ) {
      obj.comments._cb_inserted = LW.Tensor.commentInsertedCallback_;
    }
  }
};

LW.Tensor.fillCommentDiv_ = function(comments, index, div, contentOnly) {
  if ( div._block ) {
    return;
  }
  var comment = comments[index];
  var newreplies = comment.comments.length;
  // TODO: newreplies
  var title = "";
  if ( LW.Tensor.currentDoc.content._data.comments[0] == comment ) {
    title = LW.Tensor.currentDoc.content._data.title + " ";
  }  
  var html1 = '<span class="text">' + esc(comment._meta.author) + ": " + title + '</span> <span class="updates">' + (newreplies > 0 ? ('(' + newreplies.toString() + ')') : "") + '</span> <span class="date"> ' + comment._meta.date + ' </span>';
  var html2 = esc(comment.content);
  if ( contentOnly ) {
    div.children[0].innerHTML = html1;
    div.children[1].innerHTML = html2;
    return;
  }
  var html3 = '<div class="tools">';
  html3 += '<span class="view"><a href="#">Edit</a></span><span class="reply"><a href="#">Reply</a></span><span class="history"><a href="#">History</a></span><span class="edit"><a href="#">Delete</a></span>';
  html3 += '</div>';
  div.innerHTML = '<h3>' + html1 + '</h3><h4>' + html2 + '</h4>' + html3;
  div.id = LW.Tensor.currentDoc.url + "!" + comment._id;
  div.onclick = LW.Tensor.onDivClick_;
  var d = $(div);
  // Edit clicked
  d.find(".view").get(0).firstChild.onclick = function(e) {
    if ( !e ) e = window.event;
    e.cancelBubble = true;
    if ( e.stopPropagation ) e.stopPropagation();
    var html = '<div class="editBox">';
    if ( LW.Tensor.currentDoc.content._data.comments[0] == comment ) {
      html += '<div class="title"><span class="label">Title:</span> <input type="text" class="titleInput"></input></div>';
    }
    html += '<div><textarea class="textInput"></textarea></div>';
    html += '<div><span class="submit"><a href="#">Submit</a> <a href="#" class="cancel">Cancel</a></span></div>';
    html += '</div>';
    div.innerHTML = html;
    if ( LW.Tensor.currentDoc.content._data.comments[0] == comment ) {
      d.find(".titleInput").get(0).value = LW.Tensor.currentDoc.content._data.title;
    }
    d.find(".textInput").get(0).value = comment.content;
    // Submit clicked
    d.find(".submit").get(0).onclick = function(e) {
      if ( !e ) e = window.event;
      e.cancelBubble = true;
      if ( e.stopPropagation ) e.stopPropagation();
      div._block = true;
      var text = d.find(".textInput").get(0).value;
        LW.Model.changeComment(LW.Tensor.currentDoc, comments, index, text);
      if ( LW.Tensor.currentDoc.content._data.comments[0] == comment ) {
        var title = d.find(".titleInput").get(0).value;
          LW.Model.setTitle(LW.Tensor.currentDoc, title);
      }
      delete div._block;
      LW.Tensor.fillCommentDiv_(comments, index, div, false);
      return false;
    };
    // Cancel clicked
    d.find(".cancel").get(0).onclick = function(e) {
      if ( !e ) e = window.event;
      e.cancelBubble = true;
      if ( e.stopPropagation ) e.stopPropagation();
      LW.Tensor.fillCommentDiv_(comments, index, div, false);
      return false;
    };      
    return false;
  };
};

// Called when a comment has changed or been inserted
LW.Tensor.commentModifiedCallback_ = function(arr, index) {
  var comment = arr[index];
  var div = document.getElementById(LW.Tensor.currentDoc.url + "!" + comment._id);
  if ( !div ) {
    var list = $('.list[objectid=' + LW.Tensor.currentDoc.url + "!" + arr._id + "]").toArray()[0];
    // Do not display the comment?
    if ( !list ) {
      return;
    }
    div = document.createElement("div");
    div.className = "wave new";
    LW.Tensor.fillCommentDiv_(arr, index, div, false);
    list.insertBefore( div, list.children[index] );
  } else {
    LW.Tensor.fillCommentDiv_(arr, index, div, true);
  }
};

LW.Tensor.renderParticipant = function(dom, doc, obj) {
    dom.innerHTML = '<a href="#" class="button">' + esc(obj.displayName) + '</a>';
};

// Invoked when an item of the inbox has changed or been inserted
LW.Tensor.inboxModifiedCallback_ = function(arr, index) {
  var entry = arr[index];
    var html = '<h3><span class="text">' + entry.authors + '</span> <span class="updates">(' + esc(entry.msgcount) + ')</span> <div style="margin-left:4px; float:right"><img src="http://www.uni-due.de/favicon.ico"></div><span class="date"> ' + "uni-due.de" + ' </span></h3>';
  html += '<h4>' + esc(entry.digest) + ' <span class="author">' + "Today" + '</span></h4>';
  var div = document.getElementById(entry.uri);
  if ( !div ) {
    var list = document.getElementById("list-1");
    div = document.createElement("div")
    div.className = "wave new";
    div.id = entry.uri;
    div.onclick = LW.Tensor.onDivClick_;
    div.innerHTML = html;
    list.insertBefore(div, list.children[index]);
    // Is this the current document?
    if ( LW.Tensor.currentDoc && LW.Tensor.currentDoc.url == entry.uri ) {
      $(div).addClass("selected");
    }
  } else {
    div.innerHTML = html;
  }
};

// Install the event handlers for the Inbox
LW.Tensor.init = function() {
    // Retrieve session info from the server
    LW.Rpc.enqueueGet("/_sessioninfo", LW.Tensor.initCallback_, function() { } );
};

LW.Tensor.initCallback_ = function(reply) {
    // Parse the reply
    var info = JSON.parse(reply);
    LW.Rpc.user = info.user;
    LW.Rpc.domain = info.domain;
    LW.Rpc.displayName = info.displayName;
    // Show which user is logged in
    document.getElementById("username").innerText = LW.Rpc.user + "@" + LW.Rpc.domain;
    LW.Inbox.init();
    LW.Inbox.self.content._data._cb_docs = function(doc, obj, key, mut, event) {
        if ( event == LW.JsonOT.AttributeInserted ) {
            // Wait for documents being inserted in the "docs" object
            doc._data.docs._cb_inserted = function(doc, arr, index, mut, event) {
                // Wait for changes in the inserted docs object
                arr[index]._cb = function(doc, obj, key, mut, event) {
                    if ( event == LW.JsonOT.ObjectModified ) {
                        LW.Tensor.inboxModifiedCallback_(arr, index);
                    }
                }
            }
        }
    }    
    LW.Session.init();
};

// UGLY
function esc(str) {
  // TODO
  return str;
}

LW.Tensor.populateAddParticipants = function() {
    var dom = document.getElementById("footer-content");
    dom.innerHTML = '<div class="contactsearch"><input type="search"> <span>Search for people</div>';
    LW.Rpc.enqueueGet("/client/" + LW.Rpc.domain + "/_user/" + LW.Rpc.user + "?kind=friends", LW.Tensor.populateAddParticipantsCallback_);
};

LW.Tensor.updateAddParticipants = function() {
    var contacts = $('.contacts');
    if ( contacts.length == 1 ) {
        contacts[0].parentNode.removeChild(contacts[0]);
        LW.Tensor.populateAddParticipantsCallback_();
    }
};

/**
 * @param reply is optional
 */
LW.Tensor.populateAddParticipantsCallback_ = function(reply) {
    if ( reply ) {
        // Parse the reply
        var friends = JSON.parse(reply).friends;
        // Remember the list of friends
        LW.Tensor.friends = friends;
    } else if ( !LW.Tensor.friends ) {
        return;
    }
    dom = document.getElementById("footer-content");
    var ul = document.createElement("div");
    ul.className = "contacts";
    for( var i = 0; i < LW.Tensor.friends.length; ++i ) {
        var contact = LW.Tensor.friends[i];
        // If this user is already part of the wave, then do not show it
        if ( LW.Tensor.currentDoc && LW.Model.hasParticipant( LW.Tensor.currentDoc, contact.userid ) ) {
            continue;
        }
        var li = document.createElement("div");
        li.className = "contact";
        li.innerHTML = '<div style="float:left"><img src="/images/unknown.png"></div>' + esc(contact.displayName) + '<br>' + esc(contact.userid) + '<br><a class="adduser" href="#" name="' + contact.userid + '">Add to conversation</a>';
        ul.appendChild(li);
    }
    dom.appendChild(ul);

    // Event handler for adding a contact to the list of participants
    $('.adduser').click(
        function() {
            if ( !LW.Tensor.currentDoc ) {
                return;
            }
            var userid = this.name;
            var user;
            for( var i = 0; i < LW.Tensor.friends.length; ++i ) {
                console.log(LW.Tensor.friends[i].userid + " | " + userid);
                if ( LW.Tensor.friends[i].userid == userid ) {
                    user = LW.Tensor.friends[i];
                    break;
                }
            }
            if ( !user ) {
                return;
            }
            this.parentNode.parentNode.removeChild(this.parentNode);
            LW.Model.addParticipant(LW.Tensor.currentDoc, user);
        }
    );
};

$(function() {
  // Header pulldown
  var headspace = 0;
  $('header a').toggle(
    function() {
      if (headspace != 1) {
        $(this).addClass('active');
        $('header').animate({
          top : "+=150"
        }, 500, function() {
          headspace++;
        })
      }
    }, 
    function() {
      if (headspace != 0) {
        $(this).removeClass('active');
        $('header').animate({
          top : "-=150"
        }, 500, function() {
          headspace--;
        });
      }
  });

  // Footer pullup
  var footspace = 0;
  $('footer a').toggle(
    function() {
      if (footspace != 1) {
        $(this).addClass('active');
// TODO: Only if clicked on +add button
          LW.Tensor.populateAddParticipants();
        $('footer').animate({
          bottom : "+=120"
        }, 500, function() {
          footspace++;
        })
      }
    }, 
    function() {
      if (footspace != 0) {
        $(this).removeClass('active');
        $('footer').animate({
          bottom : "-=120"
        }, 500, function() {
          footspace--;
        });
      }
  });

  // Meta pullup
  var metaspace = 0;
  $('.meta a').toggle(
    function() {
      if (metaspace != 1) {
        $(this).addClass('active');
        $('.meta').animate({
          bottom : "+=100"
        }, 500, function() {
          metaspace++;
        })
      }
    }, 
    function() {
      if (metaspace != 0) {
        $(this).removeClass('active');
        $('.meta').animate({
          bottom : "-=100"
        }, 500, function() {
          metaspace--;
        });
      }
  });

  // ------------------------------------------------
  // Added by Torben

    $('#logout').click( function() {
        LW.Rpc.logout = true;
        window.location.pathname = "/_logout";
    });

    $('.newwave').click( function() {
        if ( !LW.Session.sessionCreated ) {
            alert("No session created yet");
            return;
        }
        LW.Tensor.deselectAll_();  
        // Instantiate the document
        var url = "/" + LW.Rpc.domain + "/" + LW.Inbox.uniqueId();
        var doc = LW.Inbox.getOrCreateDoc(url);
        // Show the new document and allow the user to edit the first comment
        LW.Tensor.createColumnContent_(doc.url, document.getElementById("list-2"), true);
        $("#list-2").fadeIn();
        // Fill the document with initialization data
        LW.Model.initCommentsDoc(doc);
        // Subscribe to all changes regarding this document
        LW.Session.open(doc.url, false, false);
    } );
});
