#include "jid.h"
#include "settings.h"

JID::JID()
{
}

JID::JID( const JID& jid )
        : m_name(jid.m_name), m_domain(jid.m_domain)
{
}

JID::JID( const QString& jid )
{
    int index = jid.indexOf('@');
    if ( index == -1 )
        return;
    // TODO: Check that only valid characters are used
    m_name = jid.left( index );
    m_domain = jid.mid( index + 1 );
}

JID::JID( const QString& name, const QString& domain )
        : m_name( name ), m_domain( domain )
{
}

QString JID::toString() const
{
    if ( isNull() )
        return QString::null;
    return m_name + "@" + m_domain;
}

bool JID::isLocal() const
{
    return ( m_domain == Settings::settings()->domain() );
}
