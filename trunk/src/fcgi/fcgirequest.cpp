#include "fcgirequest.h"
#include "fcgiprotocol.h"
#include "fcgi.h"
#include "wave/waveprovider.h"

#include <QByteArray>

FCGI::FCGIRequest::FCGIRequest(FCGIProtocol* driver, quint16 id, Role role, bool kc)
  : m_id(id), m_role(role), m_keepConnection(kc), m_stdinEOF(false), m_aborted(false), m_driver(driver)
{
}

FCGI::FCGIRequest::~FCGIRequest()
{
}

bool FCGI::FCGIRequest::write(const std::string& buf, OStreamType stream)
{    
    return write(buf.data(), buf.size(), stream);
}

bool FCGI::FCGIRequest::write(const char* buf, size_t count, OStreamType stream)
{
    // Split large messages in 64k blocks
    if (count > 0xffff)
    {
        size_t done = 0;
        while( done < count )
        {
            size_t x = 0xffff;
            if ( done + x > count )
                x = count - done;
            if ( !write( buf + done, x, stream ) )
                return false;
            done += x;
        }
        return true;
    }
    if (count == 0)
        return true;

    // Construct message.
    Header h(stream == STDOUT ? TYPE_STDOUT : TYPE_STDERR, m_id, count);
    bool ok = m_driver->write((const char*)&h, sizeof(Header));
    if ( !ok ) return false;
    ok = m_driver->write(buf, count);
    return ok;
}

void FCGI::FCGIRequest::endRequest(quint32 appStatus, FCGI::FCGIRequest::ProtocolStatus protStatus)
{
    // Terminate the stdout and stderr stream, and send the
    // end-request message.

    quint8 buf[64];
    quint8* p = buf;

    new(p) Header(TYPE_STDOUT, m_id, 0);
    p += sizeof(Header);
    new(p) Header(TYPE_STDERR, m_id, 0);
    p += sizeof(Header);
    new(p) EndRequestMsg(m_id, appStatus, protStatus);
    p += sizeof(EndRequestMsg);
    m_driver->write((const char*)buf, p - buf);
    m_driver->terminateRequest(m_id);
}

void FCGI::FCGIRequest::onAbort()
{
    qDebug("FCGIRequest aborted");
    m_aborted = true;
    // TODO
}

void FCGI::FCGIRequest::appendStdin( const char* data, size_t len )
{
    if ( len == 0 )
        m_stdinEOF = true;
    else
        m_stdinStream.append( data, len );
}

void FCGI::FCGIRequest::appendParam( const std::string& name, const std::string& value )
{
    m_params[QString::fromStdString(name)] = QString::fromStdString(value);
}

#include <iostream>
#include <sstream>

void FCGI::FCGIRequest::process()
{
    // Try to parse the data sent as a JSON protobuf
    QByteArray ba( QByteArray::fromRawData( m_stdinStream.data(), m_stdinStream.length() ) );

//    qDebug("Request %s", ba.constData());
//    foreach( QString key, m_params.keys() )
//        qDebug("%s=%s", qPrintable(key), qPrintable(m_params[key]));
//
    if ( requestMethod() == "PUT")
    {
        qDebug("Putting");
        WaveProvider::self()->put(this);
    }
    else if ( requestMethod() == "GET")
    {
        qDebug("Getting");
        WaveProvider::self()->get(this);
    }
    else
        errorReply("Http method not supported");
}

void FCGI::FCGIRequest::errorReply(const QString& str)
{
    std::ostringstream os;
    os << "Content-type: text/html\r\n"
            << "\r\n"
            << str.toUtf8().constData();
    write(os.str().data(), os.str().size());
    endRequest(0, FCGIRequest::REQUEST_COMPLETE);
}

void FCGI::FCGIRequest::replyJson(const QString& data)
{
    QByteArray ba;
    ba.append( "Content-type: application/json\r\n\r\n" );
    ba.append(data.toUtf8());
    write(ba.constData(), ba.length());
    endRequest(0, FCGIRequest::REQUEST_COMPLETE);
}

void FCGI::FCGIRequest::replyJson(const QByteArray& data)
{
    QByteArray ba;
    ba.append( "Content-type: application/json\r\n\r\n" );
    ba.append(data);
    write(ba.constData(), ba.length());
    endRequest(0, FCGIRequest::REQUEST_COMPLETE);
}

void FCGI::FCGIRequest::replyHtml(const QByteArray& data)
{
    QByteArray ba;
    ba.append( "Content-type: text/html\r\n\r\n" );
    ba.append(data);
    write(ba.constData(), ba.length());
    endRequest(0, FCGIRequest::REQUEST_COMPLETE);
}

void FCGI::FCGIRequest::replyNothing()
{
    QByteArray ba;
    ba.append( "Content-type: text/html\r\n\r\n" );
    write(ba.constData(), ba.length());
    endRequest(0, FCGIRequest::REQUEST_COMPLETE);
}

QString FCGI::FCGIRequest::requestUri() const
{
    QString uri = m_params["REQUEST_URI"];
    if ( uri.left(6) != "/wave/")
        return QString::null;
    return uri.mid(6);
}

QString FCGI::FCGIRequest::requestMethod() const
{
    return m_params["REQUEST_METHOD"];
}

