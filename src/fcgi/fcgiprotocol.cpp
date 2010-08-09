#include "fcgiprotocol.h"
#include "fcgirequest.h"
#include "fcgi.h"
#include "fcgiserver.h"
#include <QByteArray>

using namespace FCGI;

FCGIProtocol::FCGIProtocol(QTcpSocket* socket, FCGIServer* parent)
        : QObject(parent), m_socket( socket )
{
    connect( m_socket, SIGNAL(disconnected()), SLOT(stop()));
    connect( m_socket, SIGNAL(error(QAbstractSocket::SocketError)), SLOT(stopOnError(QAbstractSocket::SocketError)));
    connect( m_socket, SIGNAL(readyRead()), SLOT(readBytes()));
}

FCGIProtocol::~FCGIProtocol()
{
    for(ReqMap::iterator i = m_reqMap.begin(); i != m_reqMap.end(); ++i)
    {
        i->second->onAbort();
        delete i->second;
    }

    delete m_socket;
}

bool FCGIProtocol::process_unknown(quint8 type)
{
    qDebug("Unknown");
    UnknownTypeMsg msg(type);
    return write((const char*)&msg, sizeof(UnknownTypeMsg));
}

bool FCGIProtocol::process_begin_request(quint16 id, quint8 const * buf, quint16 len)
{
    Q_ASSERT( len >= sizeof(BeginRequest) );

    // Check whether we have an open request with that id already and
    // if, throw an exception.

    if (m_reqMap.find(id) != m_reqMap.end())
    {
        qDebug("FCGIProtocol received duplicate BEGIN_REQUEST id %u.", id);
        return false;
    }

    // Create a new request instance and store it away. The user may
    // get it after we've read all parameters.

    const BeginRequest* br = reinterpret_cast<const BeginRequest*>(buf);
    m_reqMap[id] = new FCGIRequest(this, id,
                                   FCGIRequest::Role((br->roleB1 << 8) + br->roleB0),
                                   (br->flags & FLAG_KEEP_CONN) == 1);
    return true;
}

bool FCGIProtocol::process_abort_request(quint16 id, quint8 const *, quint16)
{
    // Find request instance for this id. Ignore message if non
    // exists, set ignore flag otherwise.

    ReqMap::iterator req = m_reqMap.find(id);
    if (req == m_reqMap.end())
    {
        qDebug("FCGIProtocol received ABORT_REQUEST for non-existing id %i", id );
        return false;
    }

    req->second->onAbort();
    return true;
}

bool FCGIProtocol::process_params(quint16 id, quint8 const * buf, quint16 len)
{
    // Find request instance for this id. Ignore message if non
    // exists.

    ReqMap::iterator req = m_reqMap.find(id);
    if (req == m_reqMap.end())
    {
        qDebug("FCGIProtocol received PARAMS for non-existing id %i", id);
        return false;
    }

    // Is this the last message to come? Then queue the request for
    // the user.

    if (len == 0)
        return true;

    // Process message.

    const quint8* bufend = buf + len;
    quint32 name_len;
    quint32 data_len;
    while(buf != bufend)
    {
        // TODO: This looks strange!!!
        if (*buf >> 7 == 0)
            name_len = *(buf++);
        else
        {
            name_len = ((buf[0] & 0x7F) << 24) + (buf[1] << 16) + (buf[2] << 8) + buf[3];
            buf += 4;
        }
        if (*buf >> 7 == 0)
            data_len = *(buf++);
        else
        {
            data_len = ((buf[0] & 0x7F) << 24) + (buf[1] << 16) + (buf[2] << 8) + buf[3];
            buf += 4;
        }
        Q_ASSERT(buf + name_len + data_len <= bufend);
        std::string name(reinterpret_cast<char const *>(buf), name_len);
        buf += name_len;
        std::string data(reinterpret_cast<char const *>(buf), data_len);
        buf += data_len;
        //#ifdef DEBUG_FASTCGI
        //    std::cerr << "request #" << id << ": FCGIProtocol received PARAM '" << name << "' = '" << data << "'"
        //              << std::endl;
        //#endif
        req->second->appendParam(name, data);
    }

    return true;
}

bool FCGIProtocol::process_stdin(quint16 id, const quint8* buf, quint16 len)
{
    // Find request instance for this id. Ignore message if non
    // exists.

    ReqMap::iterator req = m_reqMap.find(id);
    if (req == m_reqMap.end())
    {
        qDebug("FCGIProtocol received STDIN bytes %i for non-existing id %i", (int)len, id);
        return false;
    }

    if (len == 0)
    {
        req->second->process();
        return true;
    }

    // Is this the last message to come? Then set the eof flag.
    // Otherwise, add the data to the buffer in the request structure.

    req->second->appendStdin((char const *)buf, len);
    return true;
}

bool FCGIProtocol::processInput(const char* buf, size_t count)
{
    // Copy data to our own buffer.

    m_inputBuffer.insert( m_inputBuffer.end(), (const quint8*)buf, (const quint8*)buf + count );

    // If there is enough data in the input buffer to contain a
    // header, interpret it.

    while(m_inputBuffer.size() >= sizeof(Header))
    {
        const Header* hp = reinterpret_cast<const Header*>(&m_inputBuffer[0]);

        // Check whether our peer speaks the correct protocol version.

        if (hp->version != 1)
        {
            qDebug("FCGIProtocol cannot handle protocol version %u.", hp->version);
            return false;
        }

        // Check whether we have the whole message that follows the
        // headers in our buffer already. If not, we can't process it
        // yet.

        quint16 msg_len = (hp->contentLengthB1 << 8) + hp->contentLengthB0;
        quint16 msg_id  = (hp->requestIdB1 << 8) + hp->requestIdB0;

        if (m_inputBuffer.size() < sizeof(Header) + msg_len + hp->paddingLength)
            return true;

        // Process the message. In case an exception arrives here,
        // terminate the request.

        bool ok = true;

        switch (hp->type)
        {
        case TYPE_BEGIN_REQUEST:
            {
                ok = process_begin_request(msg_id, &m_inputBuffer[0]+sizeof(Header), msg_len);
            }
            break;
        case TYPE_ABORT_REQUEST:
            {
                ok = process_abort_request(msg_id, &m_inputBuffer[0]+sizeof(Header), msg_len);
            }
            break;
        case TYPE_PARAMS:
            {
                ok = process_params(msg_id, &m_inputBuffer[0]+sizeof(Header), msg_len);
            }
            break;

        case TYPE_STDIN:
            {
                ok = process_stdin(msg_id, &m_inputBuffer[0]+sizeof(Header), msg_len);
            }
            break;

        case TYPE_END_REQUEST:
        case TYPE_STDOUT:
        case TYPE_STDERR:
        case TYPE_DATA:
        case TYPE_GET_VALUES:
        case TYPE_GET_VALUES_RESULT:
        case TYPE_UNKNOWN:
        default:
            {
                ok = process_unknown(hp->type);
            }
        }

        if ( !ok )
            terminateRequest(msg_id);

        // Remove the message from our buffer and contine processing
        // if there is something left.

        m_inputBuffer.erase( m_inputBuffer.begin(), m_inputBuffer.begin() + sizeof(Header) + msg_len+hp->paddingLength );

        if ( !ok )
            return false;
    }
    return true;
}

bool FCGIProtocol::write( const char* buf, qint64 count )
{
    return m_socket->write( buf, count ) == count;
}

void FCGIProtocol::terminateRequest(quint16 id)
{
    ReqMap::iterator req;
    req = m_reqMap.find(id);
    if (req != m_reqMap.end())
    {
        delete req->second;
        m_reqMap.erase(req);
    }
}

void FCGIProtocol::stop()
{
    deleteLater();
}

void FCGIProtocol::stopOnError(QAbstractSocket::SocketError)
{
    deleteLater();
}

void FCGIProtocol::readBytes()
{
    QByteArray ba = m_socket->readAll();
    if ( ba.length() == 0 )
        return;
    processInput( ba.constData(), ba.length() );
}

