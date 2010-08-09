#include "settings.h"
#include <QVariant>

Settings* Settings::s_settings = 0;

Settings::Settings(const QString& filename, QObject* parent)
        : QSettings(filename, QSettings::NativeFormat, parent)
{
    s_settings = this;
}

Settings* Settings::settings()
{
    if ( !s_settings )
        s_settings = new Settings( "./waveserver.conf" );
    return s_settings;
}

void Settings::setLogFile( const QString& logfile )
{
    setValue( "logFile", QVariant( logfile ) );
}

QString Settings::logFile() const
{
    return value("logFile", QVariant("commit.log") ).toString();
}

void Settings::setDomain( const QString& domain )
{
    setValue( "domain", QVariant( domain ) );
}

QString Settings::domain() const
{
    return value("domain", QVariant("localhost") ).toString();
}

bool Settings::federationEnabled() const
{
    return value("federationEnabled", QVariant(false) ).toBool();
}

void Settings::setFederationEabled( bool enabled )
{
    setValue( "federationEnabled", QVariant( enabled ) );
}

bool Settings::fcgiEnabled() const
{
    return value("fcgiEnabled", QVariant(true) ).toBool();
}

void Settings::setFcgiEnabled( bool enabled )
{
    setValue( "fcgiEnabled", QVariant( enabled ) );
}

void Settings::setFcgiPort( int port )
{
    setValue( "fcgiPort", QVariant( port ) );
}

int Settings::fcgiPort() const
{
    return value("fcgiPort", QVariant((int)9871) ).toInt();
}

QString Settings::certificateFile() const
{
    return value("certificateFile", QVariant("waveserver.crt") ).toString();
}

void Settings::setCertificateFile( const QString& file )
{
    setValue( "certificateFile", QVariant( file ) );
}

QString Settings::privateKeyFile() const
{
    return value("privateKeyFile", QVariant("waveserver.key") ).toString();
}

void Settings::setPrivateKeyFile( const QString& file )
{
    setValue( "privateKeyFile", QVariant( file ) );
}

