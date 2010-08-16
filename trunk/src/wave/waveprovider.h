#ifndef WAVEPROVIDER_H
#define WAVEPROVIDER_H

#include <QObject>
#include <QString>
#include <QHash>
#include <QRegExp>
#include "fcgi/fcgirequest.h"
#include "waveid.h"

class Session;
class RootContainer;
class WaveContainer;
class SessionContainer;

class WaveProvider : public QObject
{
public:
    WaveProvider();

    static WaveProvider* self();

    void put(FCGI::FCGIRequest* req);
    void get(FCGI::FCGIRequest* req);

    Session* session(const QString& sessionId) const;
    WaveContainer* container(const WaveId& waveId) const;

private:
    QHash<QString, Session*> m_sessions;

    QRegExp m_hostUriRegExp;
    QRegExp m_remoteUriRegExp;

    RootContainer* m_rootContainer;
    SessionContainer* m_sessionContainer;

    static WaveProvider* s_self;
};

#endif // WAVEPROVIDER_H
