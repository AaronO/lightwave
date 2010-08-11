#include "waverootdocument.h"
#include "json/jsonarray.h"
#include "ot/objectmutation.h"
#include "ot/insertmutation.h"
#include "ot/documentmutation.h"
#include "wavecontainer.h"
#include <QUrl>
#include <QNetworkRequest>

WaveRootDocument::WaveRootDocument(WaveContainer* container, const QString& docId)
    : WaveDocument(container, docId)
{
}

bool WaveRootDocument::addDocument(FCGI::FCGIRequest* req, WaveDocument* wdoc)
{
    ObjectMutation rootop(true);
    if ( !jsonObject().hasAttribute("documents"))
    {
        JSONObject documentsOp(true);
        documentsOp.setAttribute(wdoc->docId().mid( docId().length() + 1), JSONObject(true));
        rootop.setMutation("documents", documentsOp);
    }
    else
    {
        ObjectMutation documentsOp(true);
        documentsOp.setMutation(wdoc->docId().mid( docId().length() + 1), InsertMutation(JSONObject(true)));
        rootop.setMutation("documents", documentsOp);
    }
    DocumentMutation docop(rootop);
    docop.setDocumentId(docId());
    docop.setRevision(revision());

    qDebug("OP=%s", qPrintable(docop.mutation().toJSON()));

    return processMutation(req, docop);
}
