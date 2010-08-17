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
#include "waverootdocument.h"
#include "waveid.h"
#include "ot/documentmutation.h"
#include "json/jsonobject.h"
#include "fcgi/fcgirequest.h"
#include "utils/jid.h"

class Session;
class HostContainer;
class WaveMetaDocument;
class QNetworkAccessManager;
class QRegExp;

class WaveContainer : public QObject
{
public:
    WaveContainer(WaveContainer* parent, const QString& name);
    ~WaveContainer();

    QString name() const { return m_name; }

    WaveContainer* childContainer( const QString& name ) { return m_children.value(name); }
    WaveContainer* parentContainer() const { return static_cast<WaveContainer*>(parent()); }
    HostContainer* hostContainer() const;

    bool isTemporary() const { return m_isTemp; }
    void makePersistent();

    virtual JSONObject get(FCGI::FCGIRequest* req, const QString& docKind);
    JSONObject put( JSONObject doc, const QString& docKind, FCGI::FCGIRequest* req = 0 );
    JSONObject putFromHost( JSONObject doc, const QString& docKind );
    JSONObject putSnapshotFromHost( JSONObject data, const QString& docKind );
    JSONObject putFromRemote( JSONObject data, const QString& docKind );

    WaveDocument* document(const QString& docKind) const { return m_documents.value(docKind); }
    WaveMetaDocument* metaDocument() const  { return static_cast<WaveMetaDocument*>(m_documents.value("_meta")); }

    void registerSession( const QString& sessionId );
    void deregisterSession( const QString& sessionId );

    QList<DocumentMutation> getMutations( const QString& docKind, int sinceRevision = 0 );

    virtual bool isRemote() const;

    virtual WaveContainer* createWaveContainer(const QString& name);

    static QNetworkAccessManager* networkManager();

protected:
    void addContainer( WaveContainer* child );
    virtual void onDocumentUpdate(WaveDocument* wdoc);

private:
    void updateFromMetaDocument();
    void notifySessions(WaveDocument* doc, bool sendMetaDoc);

    JSONObject snapshot();
    void snapshot(JSONObject obj);

    QString m_name;
    bool m_isTemp;
    QHash<QString,WaveContainer*> m_children;
    QHash<QString,WaveDocument*> m_documents;

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
    SubmitToHostJob(WaveContainer* parent, FCGI::FCGIRequest* req, const WaveId& waveId, JSONObject data);
    ~SubmitToHostJob();

private slots:
    void onError (QNetworkReply::NetworkError code);
    void onFinished();
    void onSslErrors( const QList<QSslError> & errors );

private:
    void sendErrorToClient();

    FCGI::FCGIRequest* m_clientRequest;
    WaveId m_waveId;
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
