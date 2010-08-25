#include "session.h"
#include "wavedocument.h"
#include "json/jsonscanner.h"
#include "json/jsonobject.h"
#include "json/jsonarray.h"
#include "ot/documentmutation.h"
#include "waveprovider.h"
#include "wavecontainer.h"
#include "sessioncontainer.h"
#include "ot/objectmutation.h"
#include "ot/insertmutation.h"
#include "json/jsonconstant.h"
#include "view.h"
#include "viewcontainer.h"

Session::Session(SessionContainer* parent, const QString& sessionId)
    : WaveContainer(parent, sessionId), m_blockUpdate(false), m_eventListener(0), m_deltaListener(0)
{
}

void Session::update()
{
    if ( m_blockUpdate )
        return;

    Q_ASSERT(doc());
    QList<QString> waves = doc()->jsonObject().attributeObject("waves").attributeNames();
    foreach( QString w, waves )
    {
        if ( !m_waves.contains(w))
        {
            // Check for malfored ID
            WaveId wid(w);
            if ( wid.isNull() || (!wid.isView() && !wid.documentId().isEmpty() ) )
            {
                annotateWaveError(w, "Malformed ID");
                continue;
            }
            wid.normalize();
            // Open the wave    
            if ( openWave(wid, w))
            {
                m_waves.insert(w);
                annotateWave(w, "ok", JSONConstant(true));
            }
        }
    }

    foreach( QString w, m_waves )
    {
        if ( !waves.contains(w))
        {
            // Check for malfored ID
            WaveId wid(w);
            if ( wid.isNull() || (!wid.isView() && !wid.documentId().isEmpty() ) )
            {
                annotateWaveError(w, "Malformed ID");
                continue;
            }
            wid.normalize();
            // Close the wave
            m_waves.remove(w);
            closeWave(wid);
        }
    }

    putAnnotations();
}

bool Session::openWave(const WaveId& waveId, const QString waveName)
{
    WaveContainer* c = WaveProvider::self()->container(waveId);
    if ( !c )
    {
        annotateWaveError(waveName, "Wave does not exist");
        return false;
    }
    if ( waveId.isView() )
    {
        View* view = static_cast<ViewContainer*>(c)->view( waveId.documentId() );
        if ( !view )
        {
            qDebug("View does not exist");
            return false;
        }
        QString qid = view->registerSessionQuery(View::Query( name(), userJID() ));
        m_queries.insert(qid, waveName);
    }
    else
        c->registerSession(this->name());
    return true;
}

void Session::closeWave(const WaveId& waveId)
{
    WaveContainer* c = WaveProvider::self()->container(waveId);
    if ( !c )
    {
        qDebug("Strange, an open wave could not be closed");
        return;
    }
    c->deregisterSession(this->name());
}

void Session::annotateWaveError( const QString& waveName, const QString& error )
{
    ObjectMutation obj(true);
    obj.setMutation("ok", InsertMutation(JSONConstant(false)));
    obj.setMutation("error", InsertMutation(error));
    m_annotations[waveName] = obj;
}

void Session::annotateWave( const QString& waveName, const QString& key, const QString& value )
{
    ObjectMutation obj(true);
    obj.setMutation(key, InsertMutation(value));
    m_annotations[waveName] = obj;
}

void Session::annotateWave( const QString& waveName, const QString& key, const JSONAbstractObject& value )
{
    ObjectMutation obj(true);
    obj.setMutation(key, InsertMutation(value));
    m_annotations[waveName] = obj;
}

void Session::putAnnotations()
{
    if ( m_annotations.isEmpty() )
        return;

    m_blockUpdate = true;
    ObjectMutation obj(true);
    ObjectMutation obj2(true);
    foreach( QString w, m_annotations.keys())
    {
        obj2.setMutation(w, m_annotations[w]);
    }
    obj.setMutation("waves", obj2);
    obj.toObject().setAttribute("_rev", doc()->revision() );
    m_annotations.clear();

    qDebug("Put annos: %s", qPrintable(obj.toJSON()));
    put(obj.toObject(), "_default");
    m_blockUpdate = false;
}

void Session::notify( const QHash<QString,int>& revisions )
{
    foreach( QString wid, revisions.keys() )
        m_revisionsForEventListener[wid] = revisions[wid];
    foreach( QString wid, revisions.keys() )
        m_changedDocIdsForDeltaListener.insert(wid);

    // Tell the client
    if ( m_eventListener )
        sendEvents(m_eventListener);
    if ( m_deltaListener )
        sendEvents(m_deltaListener);
}

void Session::notify( const QString& viewId, const QString& queryId, const QHash<QString,View::IndexItemList>& newIndexItems )
{
    Q_UNUSED(viewId);

    ViewIndexItems& indexItems = m_indexItems[queryId];
    indexItems.unite(newIndexItems);

    // Tell the client
    if ( m_eventListener )
        sendEvents(m_eventListener);
    if ( m_deltaListener )
        sendEvents(m_deltaListener);
}

JSONObject Session::get(FCGI::FCGIRequest* req, const QString& docKind)
{
    if ( docKind == "_events")
        return sendEvents(req);
    else if ( docKind == "_deltas")
        return sendDeltas(req);
    else
        return this->WaveContainer::get(req, docKind);
}

JSONObject Session::sendEvents(FCGI::FCGIRequest* req)
{
    // Nothing to tell currently?
    if ( m_revisionsForEventListener.isEmpty() )
    {
        if ( m_eventListener && m_eventListener != req )
            m_eventListener->replyNothing();
        m_eventListener = req;
        return JSONObject();
    }

    JSONObject result(true);
    foreach( QString id, m_revisionsForEventListener.keys() )
        result.setAttribute(id, m_revisionsForEventListener[id]);
    m_revisionsForEventListener.clear();

    if ( req == m_eventListener )
        m_eventListener = 0;

    return result;
}

JSONObject Session::sendDeltas(FCGI::FCGIRequest* req)
{
    // Nothing to tell currently?
    if ( m_changedDocIdsForDeltaListener.isEmpty() )
    {
        if ( m_deltaListener && m_deltaListener != req )
            m_deltaListener->replyNothing();
        m_deltaListener = req;
        return JSONObject();
    }

    /**
      * The returned JSON has the following for for each container w1.com/foo:
      * {
      *     "w1.com/foo": [
      *         { ... DocumentMutation ... }
      *         { ... DocumentMutation ... }
      *         ...
      *     ]
      *     "w1.com/bar": [
      *         { ... DocumentMutation ... }
      *
      *     ]
      * }
      */
    JSONObject result(true);
    foreach( QString id, m_changedDocIdsForDeltaListener )
    {
        WaveId wid(id);
        Q_ASSERT( !wid.isNull());
        WaveContainer* c = WaveProvider::self()->container(wid);
        if ( !c )
        {
            qDebug("Strange, wave is open but could not be found %s", qPrintable(id));
            continue;
        }
        int rev = m_revisionsForEventListener[id];
        qDebug("Get mutations for %s since %i", qPrintable(id), rev);
        QList<DocumentMutation> mutations = c->getMutations(wid.documentId(), rev);
        JSONArray arr(true);
        foreach( DocumentMutation m, mutations )
        {
            JSONObject obj = m.mutation().clone().toObject();
            obj.removeAttribute("_id");
            arr.append( m.mutation().clone() );
        }
        if ( !mutations.isEmpty() )
            m_revisionsForEventListener[id] = mutations.last().revisionNumber();
        result.setAttribute(id, arr);
    }
    m_changedDocIdsForDeltaListener.clear();

    /**
      * The returned JSON has the following form where <rev> is a steadily increasing revision number
      * {
      *     "_view/_v1" : [
      *     {
      *         "_object":true,
      *         "_author":"_server",
      *         "_id":"_default",
      *         "_rev":"<rev>-0",
      *         "digestdoc1" :
      *         {
      *             "rows" :
      *             [
      *                 { "key":[a,b,c], "value":"foo" }
      *                 { "key":[a,b,d], "value":"bar" }
      *             ]
      *         }
      *         "digestdoc2" :
      *         {
      *             "rows" :
      *             [
      *                 { "key":[a,b,x], "value":"foobar" }
      *             ]
      *         }
      *     } ]
      * }
      */
    // Add deltas from views
    foreach( QString qid, m_indexItems.keys() )
    {
        JSONArray arr(true);
        ObjectMutation obj(true);
        const ViewIndexItems& items = m_indexItems.value(qid);
        foreach( QString dbId, items.keys() )
        {
            JSONObject rowsObj(true);
            JSONArray rows(true);
            foreach( const View::IndexItem& item, items[dbId] )
            {
                JSONObject r(true);
                // Remove the first element from the key since it contains the user name.
                // Passing this to the client wastes bandwidth.
                JSONArray k = item.key().clone().toArray();
                k.removeAt(0);
                r.setAttribute("key", k);
                r.setAttribute("value", item.value());
                rows.append(r);
            }
            rowsObj.setAttribute("rows", rows);
            InsertMutation ins(rowsObj);
            obj.setMutation(dbId, ins);
        }
        DocumentMutation dmut(obj);
        dmut.setAuthor("_server");
        dmut.setDocumentId("_default");
        int rev = ++m_queryRevisions[qid];
        dmut.setRevision(QString("%1-0").arg(rev));
        arr.append(dmut.mutation());

        QString viewName = m_queries.value(qid);
        result.setAttribute(viewName, arr);
    }

    if ( req == m_deltaListener )
        m_deltaListener = 0;

    return result;
}

void Session::onDocumentUpdate(WaveDocument* wdoc)
{
    if ( wdoc == doc() )
        update();
}

WaveContainer* Session::createWaveContainer(const QString& name)
{
    Q_UNUSED(name);

    return 0;
}

QString Session::userJID() const
{
    return doc()->jsonObject().attributeString("user");
}
