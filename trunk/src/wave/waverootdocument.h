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

    bool addDocument(FCGI::FCGIRequest* req, WaveDocument* wdoc);

protected:
    virtual void update();

private:
    WaveContainer* m_container;
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

#endif // WAVEROOTDOCUMENT_H
