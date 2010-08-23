#ifndef SESSION_H
#define SESSION_H

#include <QObject>
#include <QString>
#include <QSet>
#include <QHash>
#include "fcgi/fcgirequest.h"
#include "waveid.h"
#include "wavecontainer.h"

class WaveDocument;
class SessionContainer;

class Session : public WaveContainer
{
public:
    Session(SessionContainer* parent, const QString& sessionId);

    JSONObject get(FCGI::FCGIRequest* req, const QString& docKind);

    bool isRemote() const { return false; }
    bool buildsDigest() const { return false; }

    void notify( const QHash<QString,int>& revisions );

    void viewChanged(const QString& viewId, int revisionNumber);

    WaveContainer* createWaveContainer(const QString& name);

protected:
    virtual void onDocumentUpdate(WaveDocument* wdoc);

private:
    WaveDocument* doc() { return document("_default"); }
    JSONObject sendEvents(FCGI::FCGIRequest* req);
    JSONObject sendDeltas(FCGI::FCGIRequest* req);

    void update();
    bool openWave(const WaveId& waveId, const QString waveName);
    void closeWave(const WaveId& waveId);
    void annotateWaveError( const QString& waveName, const QString& error );
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

    QHash<QString,QString> m_annotations;
    bool m_blockUpdate;

    FCGI::FCGIRequest* m_eventListener;
    FCGI::FCGIRequest* m_deltaListener;
};

#endif // SESSION_H
