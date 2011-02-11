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

LW.Model.createDocument = function(text) {
    var url = "/" + LW.Rpc.domain + "/" + LW.Inbox.uniqueId();
    var doc = LW.Inbox.getOrCreateDoc(url);
    // Send delta to the server to persist the new document
    doc.submitDocMutation( {"_rev":0, "_meta":{"$object":true, "schema":"//lightwave/blips", "participants":[{userid:LW.Rpc.user + "@" + LW.Rpc.domain, displayName:LW.Rpc.displayName}]},
                            "_data":{"$object":true, "blips":[{"content":{"$rtf":true, "text":[{"_type":"parag"},text]}, "_meta":{"author":LW.Rpc.user + "@" + LW.Rpc.domain}}]}});
    return doc;
};

// Creates and sends a document mutation to insert a new blip.
//
// @param doc is a LW.Doc object
// @param objectid is the ID denoting a JSON array that contains a list of blips.
//                 It has the form "doc-uri!id"
// @param text is a string
LW.Model.createBlip = function(doc, objectid, text) {
    // Get the JSON array that stores these comments
    var blips = LW.Inbox.getElementById(objectid);
    // Create a mutation adding a new comment
    var mutation = [{"content":{"$rtf":true, "text":[{"_type":"parag"},text]}, "_meta":{"author":LW.Rpc.user + "@" + LW.Rpc.domain, "date":"Dec 4"}}];
    if ( blips.length > 0 ) {
        mutation.unshift({"$skip":blips.length});
    }
    var arrmut = {"$array":mutation};    
    var mut = doc.createMutationForId(blips._id, arrmut);
    // send the mutation to the server
    doc.submitDocMutation( mut );
};

// Creates and sends a document mutation to insert a new thread containing a new blip.
//
// @param doc is a LW.Doc object
// @param objectid is the ID denoting a JSON array that contains a list of threads.
//                 It has the form "doc-uri!id"
// @param text is a string
LW.Model.createThreadAndBlip = function(doc, objectid, text) {
    // Get the JSON array that stores these comments
    var blip = LW.Inbox.getElementById(objectid);
    // Create a mutation adding a new comment
    var mutation = {"$object":true, "threads":{"$array":[{"blips":[{"content":{"$rtf":true, "text":[{"_type":"parag"},text]}, "_meta":{"author":LW.Rpc.user + "@" + LW.Rpc.domain, "date":"Dec 4"}}]}]}};
    if ( blip.threads && blip.threads.length > 0 ) {
        mutation.threads["$array"].unshift({"$skip":blip.threads.length});
    }
    var mut = doc.createMutationForId(blip._id, mutation);
    // send the mutation to the server
    doc.submitDocMutation( mut );
};

LW.Model.addParticipant = function(doc, user) {
    var mut = {"_meta":{"$object":true, "participants":{"$array":[{"$skip":doc.content._meta.participants.length},{"userid":user.userid, "displayName":user.displayName}]}}};
    doc.submitDocMutation(mut);
};

LW.Model.hasParticipant = function(doc, userid ) {
    if ( !doc.content._meta.participants ) {
        return false;
    }
    for( var i = 0; i < doc.content._meta.participants.length; ++i ) {
        var p = doc.content._meta.participants[i];
        if ( p.userid == userid ) {
            return true;
        }
    }
    return false;
};
