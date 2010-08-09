#ifndef SESSION_H
#define SESSION_H

#include <QObject>
#include <QString>
#include <QSet>
#include <QHash>
#include "fcgi/fcgirequest.h"

class WaveDocument;
class QRegExp;

class Session : public QObject
{
public:
    Session(const QString& sessionId);

    bool get(FCGI::FCGIRequest* req);
    bool put(FCGI::FCGIRequest* req);

    QString sessionId() const { return m_sessionId; }

    void notify( const QHash<QString,QString>& revisions );
    void sendEvents(FCGI::FCGIRequest* req);
    void sendDeltas(FCGI::FCGIRequest* req);

private:
    void update();
    bool openWave(const QString& host, const QString& waveId);
    void closeWave(const QString& host, const QString& waveId);
    void annotateWaveError( const QString& id, const QString& error );

    /**
      * A set of all currently opened waves.
      * To allow for garbage collection of inactive waves, we store only the string and not
      * a direct pointer to the wave itself.
      */
    QSet<QString> m_waves;
    QString m_sessionId;
    WaveDocument* m_doc;
    QHash<QString,QString> m_revisionsForEventListener;
    QHash<QString,QString> m_revisionsForDeltaListener;

    FCGI::FCGIRequest* m_eventListener;
    FCGI::FCGIRequest* m_deltaListener;

    static QRegExp* s_waveUriRegExp;
};

#endif // SESSION_H
