#ifndef WAVEROOTDOCUMENT_H
#define WAVEROOTDOCUMENT_H

#include "wavedocument.h"
#include <QString>
#include <QObject>
#include <QByteArray>
#include <QNetworkReply>

class WaveContainer;

class WaveRootDocument : public WaveDocument
{
public:
    WaveRootDocument(WaveContainer* container, const QString& docId);

    bool addDocument(FCGI::FCGIRequest* req, WaveDocument* wdoc, bool suppressReply = false);

protected:
    virtual void update();

private:
    WaveContainer* m_container;
};

#endif // WAVEROOTDOCUMENT_H
