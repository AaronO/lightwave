#ifndef FASTCGI_H
#define FASTCGI_H

#include <string.h>
#include <QtGlobal>

namespace FCGI
{
    enum MessageType
    {
        TYPE_BEGIN_REQUEST     =  1,
        TYPE_ABORT_REQUEST     =  2,
        TYPE_END_REQUEST       =  3,
        TYPE_PARAMS            =  4,
        TYPE_STDIN             =  5,
        TYPE_STDOUT            =  6,
        TYPE_STDERR            =  7,
        TYPE_DATA              =  8,
        TYPE_GET_VALUES        =  9,
        TYPE_GET_VALUES_RESULT = 10,
        TYPE_UNKNOWN           = 11
    };

    struct Header
    {
        quint8 version;
        quint8 type;
        quint8 requestIdB1;
        quint8 requestIdB0;
        quint8 contentLengthB1;
        quint8 contentLengthB0;
        quint8 paddingLength;
        quint8 reserved;

        Header()
        {
            memset(this, 0, sizeof(*this));
        }

        Header(MessageType t, quint16 id, quint16 len)
            : version(1), type(t)
            , requestIdB1(id >> 8)
            , requestIdB0(id & 0xff)
            , contentLengthB1(len >> 8)
            , contentLengthB0(len & 0xff)
            , paddingLength(0), reserved(0)
        {
        }
    };

    struct BeginRequest
    {
        quint8 roleB1;
        quint8 roleB0;
        quint8 flags;
        quint8 reserved[5];
    };

    static quint8 const FLAG_KEEP_CONN = 1;

    struct EndRequestMsg : public Header
    {
        quint8 appStatusB3;
        quint8 appStatusB2;
        quint8 appStatusB1;
        quint8 appStatusB0;
        quint8 protocolStatus;
        quint8 reserved[3];

        EndRequestMsg()
        {
            memset(this, 0, sizeof(*this));
        }

        EndRequestMsg(quint16 id, quint32 appStatus, FCGIRequest::ProtocolStatus protStatus)
            : Header(TYPE_END_REQUEST, id, sizeof(EndRequestMsg)-sizeof(Header))
            , appStatusB3((appStatus >> 24) & 0xff)
            , appStatusB2((appStatus >> 16) & 0xff)
            , appStatusB1((appStatus >>  8) & 0xff)
            , appStatusB0((appStatus >>  0) & 0xff)
            , protocolStatus(protStatus)
        {
            memset(this->reserved, 0, sizeof(this->reserved));
        }
    };

    struct UnknownTypeMsg : public Header
    {
        quint8 type;
        quint8 reserved[7];

        UnknownTypeMsg()
        {
            memset(this, 0, sizeof(*this));
        }

        UnknownTypeMsg(quint8 unknown_type)
            : Header(TYPE_UNKNOWN, 0, sizeof(UnknownTypeMsg) - sizeof(Header))
            , type(unknown_type)
        {
            memset(this->reserved, 0, sizeof(this->reserved));
        }
    };

}

#endif // FASTCGI_H
