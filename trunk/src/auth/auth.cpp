#include "auth.h"
#include "json/jsonscanner.h"
#include "json/jsonobject.h"
#include <QCryptographicHash>
#include <QRegExp>
#include <QDateTime>

Authenticator* Authenticator::s_self = 0;

Authenticator::Authenticator()
{
    s_self = this;
}

bool Authenticator::authenticateByCookie(FCGI::FCGIRequest* req)
{
    QRegExp regexp("AuthSession=([A-Za-z0-9]+)");
    if ( !regexp.exactMatch(req->cookie()))
        return true;
    if ( !m_cookies.contains(regexp.cap(1)))
    {
        JSONObject result(true);
        result.setAttribute("ok", false);
        result.setAttribute("error", "Cookie is invalid");
        req->replyJson(result.toJSON());
        return false;
    }
    req->setAuthUser( m_cookies[regexp.cap(1)].user );
    return true;
}

QString Authenticator::authenticate(const QByteArray& data)
{
    JSONScanner scanner(data.constData(), data.count());
    bool ok = false;
    JSONObject doc = scanner.scan(&ok);
    if ( !ok )
        return QString::null;
    return authenticate(doc.attributeString("user"), doc.attributeString("passwd"));
}

QString Authenticator::authenticate(const QString& user, const QString& passwd)
{
    // TODO: Check the passwd and user

    QCryptographicHash hash(QCryptographicHash::Md5);
    hash.addData(user.toUtf8().append("/").append(passwd.toUtf8()));
    QString cookie = QString(hash.result().toHex());

    // Remember the cookie
    Cookie c;
    c.user = user;
    c.time = QDateTime::currentDateTime().toTime_t();
    m_cookies[cookie] = c;

    return cookie;
}

Authenticator* Authenticator::self()
{
    if ( s_self )
        return s_self;
    return new Authenticator();
}
