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

    void notify( const QHash<QString,QString>& revisions );
    void sendEvents(FCGI::FCGIRequest* req);
    void sendDeltas(FCGI::FCGIRequest* req);

    bool isRemote() const { return false; }

protected:
    virtual void onDocumentUpdate(WaveDocument* wdoc);

private:
    WaveDocument* doc() { return document("_default"); }

    void update();
    bool openWave(const WaveId& waveId);
    void closeWave(const WaveId& waveId);
    void annotateWaveError( const QString& id, const QString& error );

    /**
      * A set of all currently opened waves.
      * To allow for garbage collection of inactive waves, we store only the string and not
      * a direct pointer to the wave itself.
      */
    QSet<QString> m_waves;
    QHash<QString,QString> m_revisionsForEventListener;
    QHash<QString,QString> m_revisionsForDeltaListener;

    FCGI::FCGIRequest* m_eventListener;
    FCGI::FCGIRequest* m_deltaListener;
};

#endif // SESSION_H
