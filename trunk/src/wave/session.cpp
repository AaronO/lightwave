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
            if ( wid.isNull() )
            {
                annotateWaveError(w, "Malformed ID");
                continue;
            }
            wid.normalize();
            // Open the wave    
            if ( openWave(wid))
                m_waves.insert(wid.toString());
        }
    }

    foreach( QString w, m_waves )
    {
        if ( !waves.contains(w))
        {
            // Check for malfored ID
            WaveId wid(w);
            if ( wid.isNull() )
            {
                annotateWaveError(w, "Malformed ID");
                continue;
            }
            wid.normalize();
            // Close the wave
            m_waves.remove(wid.toString());
            closeWave(wid);
        }
    }

    putAnnotations();
}

bool Session::openWave(const WaveId& waveId)
{
    WaveContainer* c = WaveProvider::self()->container(waveId);
    if ( !c )
    {
        annotateWaveError(waveId.toString(), "Wave does not exist");
        return false;
    }
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

void Session::annotateWaveError( const QString& id, const QString& error )
{
//    Q_ASSERT(doc());
//    doc()->jsonObject().attributeObject("waves").attributeObject(id).setAttribute("error", error);
    m_annotations[id] = error;
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
        obj2.setMutation(w, InsertMutation(m_annotations[w]));
    }
    obj.setMutation("waves", obj2);
    obj.toObject().setAttribute("_rev", doc()->revision() );
    m_annotations.clear();

    put(obj.toObject(), "_default");
    m_blockUpdate = false;
}

void Session::notify( const QHash<QString,QString>& revisions )
{
    foreach( QString id, revisions.keys() )
        m_revisionsForEventListener[id] = revisions[id];
    foreach( QString id, revisions.keys() )
    {
        if ( !m_revisionsForDeltaListener.contains(id))
            m_revisionsForDeltaListener[id] = revisions[id];
    }

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
    if ( m_revisionsForDeltaListener.isEmpty() )
    {
        if ( m_deltaListener && m_deltaListener != req )
            m_deltaListener->replyNothing();
        m_deltaListener = req;
        return JSONObject();
    }
    JSONObject result(true);
    foreach( QString id, m_revisionsForDeltaListener.keys() )
    {
        WaveId wid(id);
        if ( wid.isNull())
        {
            qDebug("Malformed id");
            result.setAttribute("ok", false);
            result.setAttribute("error", "Malformed ID");
            return result;
        }
        WaveContainer* c = WaveProvider::self()->container(wid);
        if ( !c )
        {
            qDebug("Strange, wave is open but could not be found %s", qPrintable(id));
            continue;
        }
        QString rev = m_revisionsForEventListener[id];
        qDebug("Get mutations for %s since %s", qPrintable(id), qPrintable(rev));
        QList<DocumentMutation> mutations = c->getMutations(id, rev);
        JSONArray arr(true);
        foreach( DocumentMutation m, mutations )
        {
            JSONObject obj = m.mutation().clone().toObject();
            obj.removeAttribute("_id");
            arr.append( m.mutation().clone() );
        }
        result.setAttribute(id, arr);
    }
    m_revisionsForDeltaListener.clear();

    if ( req == m_deltaListener )
        m_deltaListener = 0;

    return result;
}

void Session::onDocumentUpdate(WaveDocument* wdoc)
{
    if ( wdoc == doc() )
        update();
}
