#ifndef SETTINGS_H
#define SETTINGS_H

#include <QSettings>

class Settings : public QSettings
{
public:
    Settings(const QString& filename, QObject* parent = 0);

    /**
      * @returns the settings object created first, or creates a new default settings object.
      */
    static Settings* settings();

    void setLogFile( const QString& logfile );
    QString logFile() const;

    void setDomain( const QString& domain );
    /**
      * Something like mycompany.com, i.e. the JID of your users is user@mycompany.com.
      */
    QString domain() const;

    bool federationEnabled() const;
    void setFederationEabled( bool enabled );

    QString certificateFile() const;
    void setCertificateFile( const QString& file );

    QString privateKeyFile() const;
    void setPrivateKeyFile( const QString& file );

    bool fcgiEnabled() const;
    void setFcgiEnabled( bool enabled );

    void setFcgiPort( int port );
    /**
      * Usually "9871"
      */
    int fcgiPort() const;

private:
    static Settings* s_settings;
};

#endif // SETTINGS_H
