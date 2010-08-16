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
    : m_hostUriRegExp("_host"), m_remoteUriRegExp("_remote")
{
    m_rootContainer = new RootContainer();
    // Create a container for local waves
    HostContainer* h = new HostContainer(m_rootContainer, Settings::settings()->domain(), true);
    h->makePersistent();
    // Create a container for sessions
    m_sessionContainer = new SessionContainer(m_rootContainer, "session");
    m_sessionContainer->makePersistent();
    qDebug("===================================");
}

WaveProvider* WaveProvider::self()
{
    if ( !s_self )
        s_self = new WaveProvider();
    return s_self;
}

Session* WaveProvider::session(const QString& sessionId) const
{
    return static_cast<Session*>(m_sessionContainer->childContainer(sessionId));
}

WaveContainer* WaveProvider::container(const WaveId& waveId) const
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
        // Perhaps the request will be answered later because it may involve talking to a hosting server
        if ( !response.isNull())
            req->replyJson(response.toJSON());
    }
    else
        req->errorReply("Error: URI syntax error");
}
