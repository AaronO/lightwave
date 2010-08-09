#ifndef WAVEPROVIDER_H
#define WAVEPROVIDER_H

#include <QObject>
#include <QString>
#include <QHash>
#include <QRegExp>
#include "fcgi/fcgirequest.h"

class WaveContainer;
class Session;

class WaveProvider : public QObject
{
public:
    WaveProvider();

    static WaveProvider* self();

    void put(FCGI::FCGIRequest* req);
    void get(FCGI::FCGIRequest* req);

    WaveContainer* waveContainer(const QString& host, const QString& waveId);
    WaveContainer* createWaveContainer(const QString& host, const QString& waveId);
    Session* createSession(FCGI::FCGIRequest* req, const QString& sessionId);
    Session* session(const QString& sessionId);

private:
    QHash<QString, WaveContainer*> m_container;
    QHash<QString, Session*> m_sessions;

    QRegExp m_waveUriRegExp;
    QRegExp m_docUriRegExp;
    QRegExp m_sessionUriRegExp;
    QRegExp m_sessionEventsUriRegExp;
    QRegExp m_sessionDeltasUriRegExp;
    QRegExp m_hostWaveUriRegExp;
    QRegExp m_hostDocUriRegExp;
    QRegExp m_remoteWaveUriRegExp;

    static WaveProvider* s_self;
};

#endif // WAVEPROVIDER_H
