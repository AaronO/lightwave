#ifndef FCGIPROTOCOLDRIVER_H
#define FCGIPROTOCOLDRIVER_H

#include <QObject>
#include <QTcpSocket>
#include <vector>
#include <map>

class QTcpSocket;

namespace FCGI
{
    class FCGIRequest;
    class FCGIServer;

    /**
      * Represents one TCP connection between web server and wave server.
      * A FCGIProtocol object can multiplex requests. For each request it
      * creates a FCGIRequest object.
      */
    class FCGIProtocol : public QObject
    {
        Q_OBJECT
    public:
        FCGIProtocol(QTcpSocket* socket, FCGIServer* parent);
        ~FCGIProtocol();

    protected:
        friend class FCGIRequest;
        /**
          * @return false on failure.
          *
          * Invoked from FCGIRequest. Sends data back to the web server
          */
        bool write( const char* buf, qint64 count );
        void terminateRequest(quint16 id);

    private slots:
        /**
          * Connected to the socket.
          */
        void stop();
        /**
          * Connected to the socket.
          */
        void stopOnError(QAbstractSocket::SocketError);
        /**
          * Connected to the socket.
          */
        void readBytes();

    private:
        /**
         * Don't copy me
         */
        FCGIProtocol(const FCGIProtocol&);
        /**
         * Don't copy me
         */
        FCGIProtocol& operator= (const FCGIProtocol &);

    private:
        /**
          * This function processes data sent by the web server.
          */
        bool processInput(const char* buf, size_t count);
        bool process_begin_request(quint16 id, const quint8* buf, quint16 len);
        bool process_abort_request(quint16 id, const quint8* buf, quint16 len);
        bool process_params(quint16 id, const quint8* buf, quint16 len);
        bool process_stdin(quint16 id, const quint8* buf, quint16 len);
        bool process_unknown(quint8 type);

        typedef std::map<quint16,FCGIRequest*> ReqMap;

        ReqMap m_reqMap;
        std::vector<quint8> m_inputBuffer;
        QTcpSocket* m_socket;

        static qint64 s_id;
    };
}

#endif // FCGIPROTOCOLDRIVER_H
