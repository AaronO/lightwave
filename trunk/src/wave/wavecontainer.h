#ifndef WAVECONTAINER_H
#define WAVECONTAINER_H

#include <QObject>
#include <QHash>
#include <QString>
#include <QSet>
#include <QList>
#include <QNetworkReply>
#include <QByteArray>

#include "wavedocument.h"
#include "ot/documentmutation.h"
#include "json/jsonobject.h"
#include "fcgi/fcgirequest.h"
#include "utils/jid.h"

class WaveRootDocument;
class Session;
class QNetworkAccessManager;
class QRegExp;

class WaveContainer : public QObject
{
public:
    WaveContainer(const QString& host, const QString& waveId);
    ~WaveContainer();

    /**
      * Invoked on behalf of a local client.
      */
    bool putDocument( FCGI::FCGIRequest* req, const QString& docId );
    bool putDocumentFromHost( FCGI::FCGIRequest* req, const QString& docId, JSONObject doc );
    bool putRootDocument( FCGI::FCGIRequest* req );
    bool putRootDocumentFromHost( FCGI::FCGIRequest* req, JSONObject doc );
    /**
      * Invoked on behalf of a remote server.
      *
      * If the docId is empty, then the target of the put is the root document.
      */
    bool putDocumentFromRemote( FCGI::FCGIRequest* req, const QString& docId = QString::null );
    void getDocument( FCGI::FCGIRequest* req, const QString& docId );    
    void getRootDocument( FCGI::FCGIRequest* req );

    WaveDocument* document(const QString& docId) { return m_documents[docId]; }

    void registerSession( const QString& sessionId );
    void deregisterSession( const QString& sessionId );

    QList<DocumentMutation> getMutations( const QString& docId, const QString& sinceRevision = QString::null );

    bool isRemote() const;
    QString host() const { return m_host; }
    QString waveId() const { return m_waveId; }

    static QNetworkAccessManager* networkManager();

private:
    friend class WaveRootDocument;

    void notifySessions(WaveDocument* doc, bool sendRootDoc);

    QString m_host;
    QString m_waveId;
    QHash<QString,WaveDocument*> m_documents;
    WaveRootDocument* m_rootDoc;
    QSet<QString> m_sessions;
    QSet<QString> m_authors;
    QSet<QString> m_remoteHosts;

    static QNetworkAccessManager* s_networkManager;
};

class SubmitToHostJob : public QObject
{
    Q_OBJECT
public:
    /**
      * If docId is empty, then the submit targets the wave root document.
      */
    SubmitToHostJob(WaveContainer* parent, FCGI::FCGIRequest* req, const QString& docId, const QByteArray& data);
    ~SubmitToHostJob();

private slots:
    void onError (QNetworkReply::NetworkError code);
    void onFinished();
    void onSslErrors( const QList<QSslError> & errors );

private:
    FCGI::FCGIRequest* m_clientRequest;
    QString m_docId;
    QByteArray m_data;
    QNetworkReply* m_serverReply;
};

class SubmitToRemoteJob : public QObject
{
    Q_OBJECT
public:
    SubmitToRemoteJob(WaveContainer* parent, const QString& host, const QByteArray& data );
    ~SubmitToRemoteJob();

private slots:
    void onError (QNetworkReply::NetworkError code);
    void onFinished();
    void onSslErrors( const QList<QSslError> & errors );

private:
    QByteArray m_data;
    QNetworkReply* m_serverReply;
};

#endif // WAVECONTAINER_H
