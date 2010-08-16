#include "waveid.h"
#include "utils/settings.h"
#include <QRegExp>

QRegExp* WaveId::s_waveUriRegExp = 0;

WaveId::WaveId()
{
    _init();
}

WaveId::WaveId(const QString& id)
{
    _init();
    if ( !s_waveUriRegExp->exactMatch(id) )
        return;
    m_host = s_waveUriRegExp->cap(1);
    m_pathItems = s_waveUriRegExp->cap(2).split('/', QString::SkipEmptyParts);
    m_docId = s_waveUriRegExp->cap(3);
    if ( !m_docId.isEmpty() )
        m_docId = m_docId.mid(1);
}

WaveId::WaveId(const WaveId& id)
    : m_host(id.m_host), m_pathItems(id.m_pathItems), m_docId( id.m_docId )
{
    _init();
}

WaveId::WaveId(const QString& host, const QStringList& pathItems, const QString& documentId)
    : m_host(host), m_pathItems(pathItems), m_docId(documentId)
{
    _init();
}

void WaveId::_init()
{
    if ( !s_waveUriRegExp )
        s_waveUriRegExp = new QRegExp("([+A-Za-z0-9.-]+)(/[+/A-Za-z0-9.-]+)(/_[+A-Za-z0-9.-_/]+)?");
}

WaveId& WaveId::operator=(const WaveId& id)
{
    m_host = id.m_host;
    m_pathItems = id.m_pathItems;
    m_docId = id.m_docId;
    return *this;
}

void WaveId::normalize()
{
    if ( m_host == "local")
        m_host = Settings::settings()->domain();
    if ( m_docId.isEmpty() )
        m_docId = "_default";
}

bool WaveId::isRemote() const
{
    return !isNull() && (m_host != "local" && m_host != Settings::settings()->domain());
}

QString WaveId::toString() const
{
    QString result = m_host;
    foreach( QString s, m_pathItems )
    {
        result += "/" + s;
    }
    if ( !m_docId.isEmpty())
        result += "/" + m_docId;
    return result;
}
