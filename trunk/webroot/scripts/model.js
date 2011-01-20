/**
* This file contains method for creating document mutations for
* the comments document schema.
*
* This is separated from LW.Doc since LW.Doc is independent of a 
* specific document schema.
*
* In MVC terms these functions are a wrapper for the model.
*/

if ( !window.LW ) {
  LW = { };
}

LW.Model = {
};

// Initialized an empty document to contain the local participant as a user and an empty list of comments
// together with an empty title.
//
// @param doc is a LW.Doc object
LW.Model.initCommentsDoc = function(doc) {
    // Send delta to the server to persist the new document
    doc.submitDocMutation( {"_rev":0, "_meta":{"$object":true, "participants":[{userid:LW.Rpc.user + "@" + LW.Rpc.domain, displayName:LW.Rpc.displayName}]},
                            "_data":{"$object":true, "title":"...", "comments":[]}});
};

LW.Model.addParticipant = function(doc, user) {
    var mut = {"_meta":{"$object":true, "participants":{"$array":[{"$skip":doc.content._meta.participants.length},{"userid":user.userid, "displayName":user.displayName}]}}};
    doc.submitDocMutation(mut);
};

// Creates a document mutation to change the title of a conversation and sends it to the server
// @param doc is a LW.Doc object
// @param title is a string
LW.Model.setTitle = function(doc, title) {
    var datamut = {"$object":true, "title":title};
    var mut = doc.createMutationForId(doc.content._data._id, datamut);
    doc.submitDocMutation( mut );
};

// Creates and sends a document mutation to insert a new comment.
//
// @param doc is a LW.Doc object
// @param objectid is the ID denoting a JSON array that contains a list of comments.
//                 It has the form "doc-uri!id"
// @param text is a string
LW.Model.createNewComment = function(doc, objectid, text) {
    if ( !text ) {
        text = "";
    }
    // Get the JSON array that stores these comments
    var comments = LW.Inbox.getElementById(objectid);
    // Create a mutation adding a new comment
    var mutation = [{"content":text, "_meta":{"author":LW.Rpc.user + "@" + LW.Rpc.domain, "date":"Dec 4"}, "comments":[]}];
    if ( comments.length > 0 ) {
        mutation.splice(0,0, {"$skip":comments.length});
    }
    var arrmut = {"$array":mutation};    
    var mut = doc.createMutationForId(comments._id, arrmut);
    // send the mutation to the server
    doc.submitDocMutation( mut );
};

// Creates a document mutation to change the contents of a comment and sends it to the server
// @param doc is a LW.Doc object
// @param comments is a JSON array that contains a list of JSON comment objects
// @param index is a position inside the comments parameter
// @param content is a string and the new content of the comment.
LW.Model.changeComment = function(doc, comments, index, content) {
    var mutation = [{"$object":true, "content":content}];
    if ( index > 0 ) {
        mutation.splice(0,0, {"$skip":index});
    }
    if ( comments.length > index + 1 ) {
        mutation.push({"$skip":comments.length - index - 1});
    }
    var arrmut = {"$array":mutation};    
    var mut = doc.createMutationForId(comments._id, arrmut);
    doc.submitDocMutation( mut );
};
