#ifndef GETOPTS_H
#define GETOPTS_H

#include <QString>
#include <QStringList>
#include <QTextStream>
#include <QSet>

class QOption
{
public:
    QOption();
    QOption(int kind, const QString& longCode) : m_kind(kind), m_occurrence(false), m_longCode(longCode), m_needsValue(false) { }
    QOption(const QOption& other);
    ~QOption() { }

    bool isNull() const { return m_kind == -1; }

    void setLongCode(const QString& codeName) {
        Q_ASSERT(!codeName.startsWith("--"));
        Q_ASSERT(!codeName.isEmpty());
        m_longCode.reserve(codeName.length() + 2);
        m_longCode.append("--").append(codeName);
        Q_ASSERT(m_longCode.length() > 2);
    }
    void setShortCode(const QChar& code) {
        Q_ASSERT(code != '-');
        m_shortCode.reserve(2);
        m_shortCode.append("-").append(code);
        Q_ASSERT(m_shortCode.length() == 2);
    }
    void setDescription(const QString & description) { m_desc = description; }
    void setNeedsValue(bool needsValue, const QString& defaultValue, const QString& valueDescription);

    int kind() const { return m_kind; }
    QString longCode() const { return m_longCode; }
    QString shortCode() const { return m_shortCode; }
    QString description() const { return m_desc; }
    bool needsValue() const { return m_needsValue; }
    QString valueDescription() const { return m_valueDescription; }
    QString defaultValue() const { return m_defaultValue; }

    bool operator==(const QOption& other) const { return other.kind() == kind(); }
    bool operator!=(const QOption& other) const { return other.kind() != kind(); }
    QOption& operator=(const QOption& other );

    uint getHash() const { return m_kind + 1; }

    void setOccurrence( const QString& value = QString::null ) { m_occurrence = true; m_value = value; }
    bool occurrence() const { return m_occurrence; }
    QString value() const { return m_value; }

private:
    int m_kind;
    bool m_occurrence;
    QString m_longCode;
    QString m_shortCode;
    QString m_desc;
    bool m_needsValue;
    QString m_value;
    QString m_valueDescription;
    QString m_defaultValue;
};

uint qHash(const QOption& option);

class QOptions
{
public:
    QOptions(int argc, const char ** argv);
    QOptions(const QStringList& arguments) : m_maxValues(0), m_minValues(0), m_arguments(arguments) { }
    ~QOptions() { }

    void setProgramName(const QString& name) { m_programName = name; }
    QString programName() const { return m_programName; }

    void setMaxNonOptionalValues(const int& maxValues) { m_maxValues = maxValues; }
    int maxNonOptionalValuesArgs() const { return m_maxValues; }
    void setMinNonOptionalValues(const int& minValues) { m_minValues = minValues; }
    int minNonOptionalValuesArgs() const { return m_minValues; }

    const QStringList& nonOptionValues() const { return m_nonOptionValues; }
    QOption option(int kind) const { return m_options[kind]; }
    QOption option(const QString& longCode) const;
    QOption option(const QChar& shortCode) const;

    QOptions& operator+=(const QOption& option) { m_options[ option.kind() ] = option; return *this; }
    void addOption(const QOption& option) { m_options[ option.kind() ] = option; }

    void printHelp(QTextStream& output);

    bool parse();
    void printError(QTextStream& output);

private:
    QString m_programName;
    int m_maxValues;
    int m_minValues;
    QStringList m_arguments;
    QStringList m_nonOptionValues;
    QHash<int,QOption> m_options;
    QString m_error;
};

#endif // GETOPTS_H
