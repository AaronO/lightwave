#ifndef WAVEID_H
#define WAVEID_H

#include <QString>
#include <QStringList>

class WaveId
{
public:
    WaveId();
    WaveId(const QString& id);
    WaveId(const WaveId& id);
    WaveId(const QString& host, const QStringList& pathItems, const QString& documentId);

    void normalize();

    bool isNull() const { return m_host.isEmpty(); }

    WaveId& operator=(const WaveId& id);
    bool operator==(const WaveId& id) const { return m_host == id.m_host && m_pathItems == id.m_pathItems && m_docId == id.m_docId; }
    bool operator!=(const WaveId& id) const { return m_host != id.m_host || m_pathItems != id.m_pathItems || m_docId != id.m_docId; }

    bool isLocal() const { return !isNull() && !isRemote(); }
    bool isRemote() const;
    QString pathItem(int i) const { return m_pathItems.at(i); }
    QStringList pathItems() const { return m_pathItems; }
    int pathItemCount() const { return m_pathItems.count(); }
    QString documentId() const { return m_docId; }
    void clearDocumentId() { m_docId = QString::null; }
    void setDocumentId(const QString& docId) { m_docId = docId; }
    QString host() const { return m_host; }

    QString toString() const;

private:
    void _init();

    QString m_host;
    QStringList m_pathItems;
    QString m_docId;

    static QRegExp* s_waveUriRegExp;
    static QRegExp* s_sessionUriRegExp;
};

#endif // WAVEID_H
