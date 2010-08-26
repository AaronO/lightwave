#ifndef WAVECONTAINER_H
#define WAVECONTAINER_H

#include <QObject>
#include <QHash>
#include <QString>
#include <QSet>
#include <QList>
#include <QNetworkReply>
#include <QByteArray>
#include <QScriptValue>

#include "wavedocument.h"
#include "waverootdocument.h"
#include "waveid.h"
#include "ot/documentmutation.h"
#include "json/jsonobject.h"
#include "json/jsonarray.h"
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
    WaveId waveId() const;

    QList<WaveContainer*> childContainers() const { return m_children.values(); }
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

    QList<WaveDocument*> documents() const { return m_documents.values(); }
    WaveDocument* document(const QString& docKind) const { return m_documents.value(docKind); }
    WaveMetaDocument* metaDocument() const  { return static_cast<WaveMetaDocument*>(m_documents.value("_meta")); }

    void registerSession( const QString& sessionId );
    void deregisterSession( const QString& sessionId );

    QList<DocumentMutation> getMutations( const QString& docKind, int sinceRevision = 0 );

    virtual bool isRemote() const;
    virtual bool buildsDigest() const { return true; }

    virtual WaveContainer* createWaveContainer(const QString& name);

    void addView(const QString& viewId, int revisionNumber ) { m_views.insert(viewId, revisionNumber); }

//    QScriptValue digestMapping(const QString& viewId) const { return m_digestMap.value(viewId); }
//    QScriptValue digestReduction(const QString& viewId) const { return m_digestReduce.value(viewId); }

    static QNetworkAccessManager* networkManager();

protected:
    void addContainer( WaveContainer* child );
    /**
      * Invoked when a document changes.
      * The default implementation triggers and update of the digest and indices.
      */
    virtual void onDocumentUpdate(WaveDocument* wdoc);
    virtual WaveDocument* createDocument(const QString& docId);
    virtual void updateDigest();
    // virtual void updateDigestReduce(const QString& viewId );

private:
    /**
      * Called when the meta document has changed to sync internal data structures
      * with the new meta document content.
      */
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

    /**
      * The key is the waveid of the view and the value is the revision that
      * has been used to build digest and indices.
      */
    QHash<QString,int> m_views;
//    /**
//      * The key is the view name and the value is the result of digest mapping.
//      */
//    QHash<QString,QScriptValue> m_digestMap;
//    /**
//      * The key is the view name concatenated with "/" and the index name and the value is the result of digest reduction.
//      */
//    QHash<QString,QScriptValue> m_digestReduce;
    /**
      * The key is the view name concatenated with "/" and the index name and the value is the result of index mapping.
      */
    // QHash<QString,JSONArray> m_indices;

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
