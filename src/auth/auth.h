#ifndef AUTH_H
#define AUTH_H

#include <QString>
#include <QHash>

#include "fcgi/fcgirequest.h"

class QByteArray;

class Authenticator
{
public:
    bool authenticateByCookie(FCGI::FCGIRequest* req);
    QString authenticate(const QByteArray& data);
    QString authenticate(const QString& user, const QString& passwd);

    static Authenticator* self();

private:
    Authenticator();

    struct Cookie
    {
        QString user;
        uint time;
    };

    QHash<QString,Cookie> m_cookies;

    static Authenticator* s_self;
};

#endif //  AUTH_H
