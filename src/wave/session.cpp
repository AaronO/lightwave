#include "session.h"
#include "wavedocument.h"
#include "json/jsonscanner.h"
#include "json/jsonobject.h"
#include "json/jsonarray.h"
#include "ot/documentmutation.h"
#include "waveprovider.h"
#include "wavecontainer.h"
#include <QRegExp>

QRegExp* Session::s_waveUriRegExp = 0;

Session::Session(const QString& sessionId)
    : m_sessionId(sessionId), m_eventListener(0), m_deltaListener(0)
{
    if ( !s_waveUriRegExp )
        s_waveUriRegExp = new QRegExp("([A-Za-z0-9.-]+)/(w\\+[A-Za-z0-9]+)(/d\\+[A-Za-z0-9]+)?");
    m_doc = new WaveDocument(sessionId);
}

bool Session::get(FCGI::FCGIRequest* req)
{
    req->replyJson(m_doc->jsonObject().toJSON());
    return true;
}

bool Session::put(FCGI::FCGIRequest* req)
{
    //
    // Scan the document
    //
    JSONScanner scanner(req->m_stdinStream.data(), req->m_stdinStream.size());
    bool ok = false;
    JSONObject doc = scanner.scan(&ok);
    if ( !ok )
    {
        req->errorReply("JSON parsing error");
        return false;
    }

    //
    // Store the document
    //

    // Initial submission?
    if ( doc.attributeString("_rev").isEmpty())
    {
        if ( !m_doc->revision().isNull() )
        {
            req->errorReply("Error: Session already exists");
            return false;
        }

        // Apply the initial mutation
        AbstractMutation m(doc);
        DocumentMutation docop( m );
        if ( !m_doc->processMutation(req, docop))
            return false;
    }
    // Overwrite/mutate the document?
    else
    {
        if ( doc.attributeString("_rev") != m_doc->revision() )
        {
           req->errorReply("Error: Attemp to overwrite based on outdated version");
           return false;
        }

        AbstractMutation m(doc);
        DocumentMutation docop( m );
        if ( !m_doc->processMutation(req, docop))
            return false;
    }

    //
    // Let's see what has changed
    //

    update();

    //
    // Send a reply
    //

    JSONObject obj(true);
    obj.setAttribute("ok", true);
    obj.setAttribute("id", m_doc->docId());
    obj.setAttribute("rev", m_doc->revision());
    req->replyJson(obj.toJSON());

    return true;
}

void Session::update()
{
    QList<QString> waves = m_doc->jsonObject().attributeObject("waves").attributeNames();
    foreach( QString w, waves )
    {
        if ( !m_waves.contains(w))
        {
            // Check for malfored ID
            if ( !s_waveUriRegExp->exactMatch(w) )
                continue;
            // Open the wave    
            if ( openWave(s_waveUriRegExp->cap(1), s_waveUriRegExp->cap(2)) )
                m_waves.insert(w);
        }
    }

    foreach( QString w, m_waves )
    {
        if ( !waves.contains(w))
        {
            // Check for malfored ID
            if ( !s_waveUriRegExp->exactMatch(w) )
                continue;
            // Close the wave
            m_waves.remove(w);
            closeWave(s_waveUriRegExp->cap(1), s_waveUriRegExp->cap(2));
        }
    }
}

bool Session::openWave(const QString& host, const QString& waveId)
{
    WaveContainer* c = WaveProvider::self()->waveContainer(host, waveId);
    if ( !c )
    {
        annotateWaveError(host + "/" + waveId, "Wave does not exist");
        return false;
    }
    c->registerSession(this->sessionId());
    return true;
}

void Session::closeWave(const QString& host, const QString& waveId)
{
    WaveContainer* c = WaveProvider::self()->waveContainer(host, waveId);
    if ( !c )
    {
        qDebug("Strange, an open wave could not be closed");
        return;
    }
    c->deregisterSession(this->sessionId());
}

void Session::annotateWaveError( const QString& id, const QString& error )
{
    m_doc->jsonObject().attributeObject("waves").attributeObject(id).setAttribute("error", error);
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

void Session::sendEvents(FCGI::FCGIRequest* req)
{
    // Nothing to tell currently?
    if ( m_revisionsForEventListener.isEmpty() )
    {
        if ( m_eventListener && m_eventListener != req )
            m_eventListener->replyNothing();
        m_eventListener = req;
        return;
    }

    JSONObject result(true);
    foreach( QString id, m_revisionsForEventListener.keys() )
        result.setAttribute(id, m_revisionsForEventListener[id]);
    req->replyJson(result.toJSON());
    m_revisionsForEventListener.clear();

    if ( req == m_eventListener )
        m_eventListener = 0;
}

void Session::sendDeltas(FCGI::FCGIRequest* req)
{
    // Nothing to tell currently?
    if ( m_revisionsForDeltaListener.isEmpty() )
    {
        if ( m_deltaListener && m_deltaListener != req )
            m_deltaListener->replyNothing();
        m_deltaListener = req;
        return;
    }
    JSONObject result(true);
    foreach( QString id, m_revisionsForDeltaListener.keys() )
    {
        QString host, waveId;
        if ( !s_waveUriRegExp->exactMatch(id) )
        {
            qDebug("Malformed id");
                return;
        }
        WaveContainer* c = WaveProvider::self()->waveContainer(s_waveUriRegExp->cap(1), s_waveUriRegExp->cap(2));
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
    req->replyJson(result.toJSON());
    m_revisionsForDeltaListener.clear();

    if ( req == m_deltaListener )
        m_deltaListener = 0;
}
