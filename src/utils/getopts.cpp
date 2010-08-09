#include "getopts.h"

QOption::QOption()
    : m_kind(-1), m_occurrence(0), m_needsValue(false)
{
}

QOption::QOption(const QOption& other)
    : m_kind(other.m_kind),
      m_occurrence(other.m_occurrence),
      m_longCode(other.m_longCode),
      m_shortCode(other.m_shortCode),
      m_desc(other.m_desc),
      m_needsValue(other.m_needsValue),
      m_value(other.m_value),
      m_valueDescription(other.m_valueDescription),
      m_defaultValue(other.m_defaultValue)
{
}

QOption& QOption::operator=(const QOption& other )
{
    m_kind = other.m_kind;
    m_occurrence = other.m_occurrence;
    m_longCode = other.m_longCode;
    m_shortCode = other.m_shortCode;
    m_desc = other.m_desc;
    m_needsValue = other.m_needsValue;
    m_value = other.m_value;
    m_valueDescription = other.m_valueDescription;
    m_defaultValue = other.m_defaultValue;
    return *this;
}

void QOption::setNeedsValue(bool needsValue, const QString& defaultValue, const QString& valueDescription)
{
    m_needsValue = needsValue;
    m_valueDescription = valueDescription;
    m_defaultValue = defaultValue;
    m_value = defaultValue;
}

uint qHash(const QOption& option)
{
    return option.getHash();
}

QOptions::QOptions(int argc, const char ** argv)
        : m_maxValues(0)
{
    for (int i = 0; i < argc; ++i)
    {
        m_arguments += QString(argv[i]);
    }
}

void QOptions::printHelp(QTextStream& out)
{
    out << "Usage: " << programName();
    if ( m_options.count() > 0 )
        out << " [options]";
    if ( m_maxValues == 1 )
        out << " [file]";
    else if ( m_maxValues > 0 )
        out << " [files]";
    out << endl << endl;
    foreach( QOption o, m_options )
    {
        if ( !o.shortCode().isEmpty() )
            out << "-" << o.shortCode() << ", --" << o.longCode();
        else
            out << "    --" << o.longCode();
        if ( o.needsValue() )
            out << " " << o.valueDescription();
        out << "\t" << o.description() << endl;
    }
    out << endl << endl;
}

bool QOptions::parse()
{
    bool incomplete = false;
    int incompleteKind;
    QString incompleteName;

    int i = 1;
    while( i < m_arguments.length() )
    {
        QString str = m_arguments[i];
        if ( incomplete )
        {
            m_options[ incompleteKind ].setOccurrence(str);
            i++;
            incomplete = false;
            continue;
        }
        if ( str.startsWith("--") )
        {
            str = str.mid(2);
            QOption o = option(str);
            if ( o.isNull() )
            {
                m_error = "No such argument: --" + str;
                return false;
            }
            if ( o.needsValue() )
            {
                incomplete = true;
                incompleteKind = o.kind();
                incompleteName = "--" + str;
            }
            else
                m_options[ o.kind() ].setOccurrence();
        }
        else if ( str.startsWith("-") && str.length() == 2 )
        {
            QOption o = option(str[1]);
            if ( o.isNull() )
            {
                m_error = "No such argument: " + str;
                return false;
            }
            if ( o.needsValue() )
            {
                incomplete = true;
                incompleteKind = o.kind();
                incompleteName = str;
            }
            else
                m_options[ o.kind() ].setOccurrence();
        }
        else
            m_nonOptionValues.append( str );
        i++;
    }

    if ( incomplete )
    {
        m_error = "Missing value for option " + incompleteName;
        return false;
    }

    if ( m_nonOptionValues.length() > maxNonOptionalValuesArgs() )
    {
        m_error = "Too many files";
        return false;
    }
    if ( m_nonOptionValues.length() < minNonOptionalValuesArgs() )
    {
        m_error = "Too few files";
        return false;
    }

    return true;
}

void QOptions::printError(QTextStream& out)
{
    out << m_error << endl;
}

QOption QOptions::option(const QString& longCode) const
{
    foreach( QOption o, m_options.values() )
    {
        if ( o.longCode() == longCode )
            return o;
    }
    return QOption();
}

QOption QOptions::option(const QChar& shortCode) const
{
    foreach( QOption o, m_options.values() )
    {        
        if ( o.shortCode()[1] == shortCode )
            return o;
    }
    return QOption();
}
