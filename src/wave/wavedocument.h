#ifndef WAVEDOCUMENT_H
#define WAVEDOCUMENT_H

#include <QObject>
#include <QString>
#include <QRegExp>
#include <QList>
#include "json/jsonobject.h"
#include "ot/documentmutation.h"
#include "fcgi/fcgirequest.h"

class DocumentMutation;

class WaveDocument : public QObject
{
public:
    WaveDocument(const QString& docId);

    bool processMutation(FCGI::FCGIRequest* req, DocumentMutation docop);

    QString docId() const { return m_docId; }
    QString revision() const { return m_rev; }
    int revisionNumber() const { return m_revNumber; }
    JSONObject jsonObject() const { return m_json; }

    DocumentMutation mutation(int revision) { return m_mutations.at(revision); }
    QList<DocumentMutation> getMutations( const QString& sinceRevision );

protected:
    bool setContent(FCGI::FCGIRequest* req, JSONObject obj);
    virtual void update();

private:
    QString m_docId;
    QString m_rev;
    int m_revNumber;
    JSONObject m_json;

    QList<DocumentMutation> m_mutations;

    static QRegExp* s_revRegExp;
};

#endif // WAVEDOCUMENT_H
