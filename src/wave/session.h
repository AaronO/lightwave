#ifndef SESSION_H
#define SESSION_H

#include <QObject>
#include <QString>
#include <QSet>
#include <QHash>
#include "fcgi/fcgirequest.h"
#include "waveid.h"
#include "wavecontainer.h"
#include "ot/abstractmutation.h"
#include "view.h"

class WaveDocument;
class SessionContainer;
class JSONAbstractObject;

class Session : public WaveContainer
{
public:
    Session(SessionContainer* parent, const QString& sessionId);

    JSONObject get(FCGI::FCGIRequest* req, const QString& docKind);

    bool isRemote() const { return false; }
    bool buildsDigest() const { return false; }

    void notify( const QHash<QString,int>& revisions );
    void notify( const QString& viewId, const QString& queryId, const QHash<QString,View::IndexItemList>& newIndexItems );

    QString userJID() const;

protected:
    WaveContainer* createWaveContainer(const QString& name);
    virtual void onDocumentUpdate(WaveDocument* wdoc);

private:
    WaveDocument* doc() const { return document("_default"); }
    JSONObject sendEvents(FCGI::FCGIRequest* req);
    JSONObject sendDeltas(FCGI::FCGIRequest* req);

    void update();
    bool openWave(const WaveId& waveId, const QString waveName);
    void closeWave(const WaveId& waveId);
    void annotateWaveError( const QString& waveName, const QString& error );
    void annotateWave( const QString& waveName, const QString& key, const QString& value );
    void annotateWave( const QString& waveName, const QString& key, const JSONAbstractObject& value );
    void putAnnotations();

    /**
      * A set of all currently opened waves.
      * To allow for garbage collection of inactive waves, we store only the waveId string and not
      * a direct pointer to the wave itself.
      */
    QSet<QString> m_waves;

    QHash<QString,int> m_revisionsForEventListener;
    QHash<QString,int> m_revisionsForDeltaListener;
    QSet<QString> m_changedDocIdsForDeltaListener;

    QHash<QString,AbstractMutation> m_annotations;
    bool m_blockUpdate;

    /**
      * The key is a query ID.
      * Value is a wave name as used in the jsonObject(). The name denotes a view.
      */
    QHash<QString,QString> m_queries;
    /**
      * The key is a query ID.
      */
    QHash<QString,int> m_queryRevisions;

    typedef QHash<QString,View::IndexItemList> ViewIndexItems;
    /**
      * The key is the query ID
      */
     QHash<QString,ViewIndexItems> m_indexItems;

    FCGI::FCGIRequest* m_eventListener;
    FCGI::FCGIRequest* m_deltaListener;
};

#endif // SESSION_H
