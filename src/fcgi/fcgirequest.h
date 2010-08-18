#ifndef FCGIREQUEST_H
#define FCGIREQUEST_H

#include <QtGlobal>
#include <QHash>
#include <QString>
#include <string>

namespace webclient
{
    class Response;
}

namespace FCGI
{
    class FCGIProtocol;

    class FCGIRequest
    {
    public:
        enum Role
        {
            RESPONDER  = 1,
            AUTHORIZER = 2,
            FILTER     = 3
        };

        enum ProtocolStatus
        {
            REQUEST_COMPLETE = 0,
            CANT_MPX_CONN    = 1,
            OVERLOADED       = 2,
            UNKNOWN_ROLE     = 3
        };

        ~FCGIRequest();

        void replyHtml(const QByteArray& arr);
        void replyJson(const QString& data);
        void replyJson(const QByteArray& arr);
        void errorReply(const QString& str);
        void replyNothing();

        QString requestUri() const;
        QString requestMethod() const;

        QString cookie() const { return m_params["HTTP_COOKIE"]; }
        void setAuthUser(const QString& user) { m_authUser = user; }
        QString authUser() const { return m_authUser; }
        bool isAuthenticated() const { return !m_authUser.isEmpty(); }

        std::string m_stdinStream;

    protected:
        friend class FCGIProtocol;

        FCGIRequest(FCGIProtocol* driver, quint16 id, Role role, bool keepConnection);

        enum OStreamType
        {
            STDOUT,
            STDERR
        };

        bool write(const std::string& buf, OStreamType stream = STDOUT);
        bool write(const char* buf, size_t count, OStreamType stream = STDOUT);
        void endRequest(quint32 appStatus, ProtocolStatus protStatus);

        /**
          * Invoked if the web server aborts the request. Clean up all allocated resources here.
          */
        void onAbort();
        void appendStdin( const char* data, size_t len );
        void appendParam( const std::string& name, const std::string& value );
        void process();

    private:

        quint16 const m_id;
        Role const m_role;
        bool const m_keepConnection;
        QHash<QString,QString> m_params;
        bool m_stdinEOF;
        bool m_aborted;
        FCGIProtocol* m_driver;       
        QString m_authUser;
    };
}

#endif // FCGIREQUEST_H
