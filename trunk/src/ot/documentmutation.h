#ifndef DOCUMENTMUTATION_H
#define DOCUMENTMUTATION_H

#include "objectmutation.h"
#include "liftmutation.h"
#include "squeezemutation.h"

#include <QHash>
#include <QString>

class ObjectMutation;
class JSONObject;

class DocumentMutation
{
public:
    DocumentMutation();
    DocumentMutation(const DocumentMutation& mutation);
    DocumentMutation(const AbstractMutation& mutation);

    AbstractMutation mutation() const { return m_mutation; }
    QString documentId() const;
    void setDocumentId( const QString& docId );
    int revisionNumber() const;
    QString revision() const;
    void setRevision( const QString& rev );
    QString author() const;
    void setAuthor( const QString& author );

    /**
      * If the mutation could not be applied, then the JSON object
      * is not modified. The function checks the correctness of
      * the mutation before it applies it.
      */
    JSONObject apply(JSONObject obj, bool* ok);

private:
    class Data
    {
    public:
        Data(bool* ok) { this->ok = ok; if ( ok ) *ok = true; }

        void setError() { if ( ok ) *ok = false; }
        inline bool hasError() const { return !*ok; }

        bool* ok;        
        QHash<QString,JSONAbstractObject> lifted;
        QHash<QString,SqueezeMutation> squeezes;
        QHash<QString,LiftMutation> lifts;
    };

    void check(JSONAbstractObject dest, AbstractMutation mutation, Data* data);
    void check(JSONAbstractObject dest, ObjectMutation mutation, Data* data);
    void check(JSONAbstractObject dest, ArrayMutation mutation, Data* data);
    void check(JSONAbstractObject dest, TextMutation mutation, Data* data);
    void check(JSONAbstractObject dest, RichTextMutation mutation, Data* data);
    void check(JSONAbstractObject dest, LiftMutation mutation, Data* data);
    void check(InsertMutation mutation, Data* data, bool ignore_underscore = false);
    void check(SqueezeMutation mutation, Data* data);

    JSONAbstractObject apply(JSONAbstractObject dest, AbstractMutation mutation, Data* data);
    JSONAbstractObject apply(JSONAbstractObject dest, ObjectMutation mutation, Data* data);
    JSONAbstractObject apply(JSONAbstractObject dest, ArrayMutation mutation, Data* data);
    JSONAbstractObject apply(JSONAbstractObject dest, TextMutation mutation, Data* data);
    JSONAbstractObject apply(JSONAbstractObject dest, RichTextMutation mutation, Data* data);
//    JSONAbstractObject apply(InsertMutation mutation, Data* data);

    AbstractMutation m_mutation;
};

#endif // DOCUMENTMUTATION_H
