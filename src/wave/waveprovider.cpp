#include "waveprovider.h"
#include "rootcontainer.h"
#include "hostcontainer.h"
#include "sessioncontainer.h"
#include "utils/settings.h"
#include "session.h"
#include "json/jsonobject.h"
#include "json/jsonscanner.h"

WaveProvider* WaveProvider::s_self = 0;

WaveProvider::WaveProvider()
    : m_sessionUriRegExp("_session/([+A-Za-z0-9.-]+)"),
      m_sessionEventsUriRegExp("(_session/([+A-Za-z0-9.-]+)/_events"), m_sessionDeltasUriRegExp("_session/([+A-Za-z0-9.-]+)/_deltas"),
      m_hostUriRegExp("_host"), m_remoteUriRegExp("_remote")
{
    m_rootContainer = new RootContainer();
    // Create a container for local waves
    HostContainer* h = new HostContainer(m_rootContainer, Settings::settings()->domain(), true);
    h->makePersistent();
    m_sessionContainer = new SessionContainer(m_rootContainer, "_settings");
    m_sessionContainer->makePersistent();
}

WaveProvider* WaveProvider::self()
{
    if ( !s_self )
        s_self = new WaveProvider();
    return s_self;
}

Session* WaveProvider::createSession(FCGI::FCGIRequest* req, const QString& sessionId)
{
    if ( m_sessions.contains(sessionId))
    {
        req->errorReply("Session with this ID already exists");
        return 0;
    }

    Session* session = new Session(m_sessionContainer, sessionId);
    session->makePersistent();
    m_sessions[sessionId] = session;
    return session;
}

Session* WaveProvider::session(const QString& sessionId)
{
    return static_cast<Session*>(m_sessionContainer->childContainer(sessionId));
}

WaveContainer* WaveProvider::container(const WaveId& waveId)
{
    return m_rootContainer->container(waveId);
}

void WaveProvider::put(FCGI::FCGIRequest* req)
{
    WaveId waveId(req->requestUri());
    if ( !waveId.isNull())
    {
        waveId.normalize();
        qDebug("waveId = %s", qPrintable(waveId.toString()));
        JSONObject response = m_rootContainer->put(req, waveId);
        // Perhaps the request will be answered later because it may involve talking to a hosting server
        if ( !response.isNull())
            req->replyJson(response.toJSON());
    }
    // Federation Host Port: http://host/wave/_host
    else if ( m_hostUriRegExp.exactMatch(req->requestUri()) )
    {
        // Parse the JSON document
        JSONObject doc(true);
        JSONScanner scanner(req->m_stdinStream.data(), req->m_stdinStream.size());
        bool ok = false;
        doc = scanner.scan(&ok);
        if ( !ok )
        {
            req->errorReply("JSON parsing error");
            return;
        }

        JSONObject result(true);
        qDebug("_host: %s", qPrintable(doc.toJSON()));

        // Iterate over the deltas for all the documents
        foreach( QString id, doc.attributeNames() )
        {
            WaveId wid(id);
            if ( wid.isNull() )
            {
                qDebug("Error: Malformed wave ID: %s", qPrintable(id));
                JSONObject e(true);
                e.setAttribute("ok", true);
                e.setAttribute("error", "Malformed wave ID");
                result.setAttribute(id, e);
                continue;
            }

            JSONObject response = m_rootContainer->putFromRemoteServer(doc.attributeObject(id), wid);
            result.setAttribute(id, response);
        }

        req->replyJson(result.toJSON());
    }
    // Session: http://host/wave/_session/xxxxx
    else if ( m_sessionUriRegExp.exactMatch(req->requestUri()) )
    {
        // Parse the JSON document
        JSONScanner scanner(req->m_stdinStream.data(), req->m_stdinStream.size());
        bool ok = false;
        JSONObject doc = scanner.scan(&ok);
        if ( !ok )
        {
            req->errorReply("JSON parsing error");
            return;
        }

        Session* s = m_sessions.value(m_sessionUriRegExp.cap(1));
        if ( !s )
        {
            s = createSession(req, m_sessionUriRegExp.cap(1));
            if ( !s )
                return;
        }
        // TODO: return a JSON object here
        s->put(doc, "_default");
    }
    // http://host/wave/_remote
    else if ( m_remoteUriRegExp.exactMatch(req->requestUri()) )
    {
        // Parse the JSON document
        JSONScanner scanner(req->m_stdinStream.data(), req->m_stdinStream.size());
        bool ok = false;
        JSONObject doc = scanner.scan(&ok);
        if ( !ok )
        {
            req->errorReply("JSON parsing error");
            return;
        }

        JSONObject result(true);
        qDebug("_remote: %s", qPrintable(doc.toJSON()));

        // Iterate over the deltas for all the documents
        foreach( QString id, doc.attributeNames() )
        {
            WaveId wid(id);
            if ( wid.isNull() )
            {
                JSONObject e(true);
                e.setAttribute("ok", true);
                e.setAttribute("error", "Malformed wave ID");
                result.setAttribute(id, e);
                qDebug("Error: Malformed wave ID: %s", qPrintable(id));
                continue;
            }

            JSONObject response = m_rootContainer->putFromHostingServer( doc.attributeObject(id), wid );
            result.setAttribute(id, response);
        }

        req->replyJson(result.toJSON());
    }
    else
    {
        JSONObject result(true);
        result.setAttribute("ok", false);
        result.setAttribute("error", "URI syntax error");
        req->replyJson(result.toJSON());
    }
}

void WaveProvider::get(FCGI::FCGIRequest* req)
{
    WaveId waveId(req->requestUri());
    if ( !waveId.isNull())
    {
        waveId.normalize();
        JSONObject response = m_rootContainer->get(req, waveId);
        req->replyJson(response.toJSON());
    }
    else if ( m_sessionUriRegExp.exactMatch(req->requestUri()) )
    {        
        Session* s = m_sessions.value(m_sessionUriRegExp.cap(1));
        if ( !s )
        {
            req->errorReply("Error: Session does not exist");
            return;
        }
        s->get(req, "_default");
    }
    else if ( m_sessionEventsUriRegExp.exactMatch(req->requestUri()) )
    {
        Session* s = m_sessions.value(m_sessionEventsUriRegExp.cap(1));
        if ( !s )
        {
            req->errorReply("Error: Session does not exist");
            return;
        }
        s->sendEvents(req);
    }
    else if ( m_sessionDeltasUriRegExp.exactMatch(req->requestUri()) )
    {
        Session* s = m_sessions.value(m_sessionDeltasUriRegExp.cap(1));
        if ( !s )
        {
            req->errorReply("Error: Session does not exist");
            return;
        }
        s->sendDeltas(req);
    }
    else
        req->errorReply("Error: URI syntax error");
}
