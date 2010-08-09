#ifndef JID_H
#define JID_H

#include <QString>

/**
  * Class for parsing a Jabber identifier, i,e, a string of the form "user@domain".
  */
class JID
{
public:
    JID();
    JID( const JID& jid );
    JID( const QString& jid );
    JID( const QString& name, const QString& domain );

    QString toString() const;
    QString name() const { return m_name; }
    QString domain() const { return m_domain; }
    bool isValid() const { return !m_name.isEmpty() && !m_domain.isEmpty(); }
    bool isNull() const { return m_name.isEmpty() && m_domain.isEmpty(); }

    bool isLocal() const;

    bool operator==( const JID& jid ) const { return m_name == jid.m_name && m_domain == jid.m_domain; }
    bool operator!=( const JID& jid ) const { return m_name != jid.m_name || m_domain != jid.m_domain; }

private:
    QString m_name;
    QString m_domain;
};

#endif // JID_H
