#ifndef ABSTRACTMUTATION_H
#define ABSTRACTMUTATION_H

#include "json/jsonabstractobject.h"

class ObjectMutation;
class ArrayMutation;
class TextMutation;
class SkipMutation;
class DeleteMutation;
class LiftMutation;
class SqueezeMutation;
class InsertMutation;
class RichTextMutation;

class AbstractMutation : public JSONAbstractObject
{
public:
    AbstractMutation();
    AbstractMutation(const AbstractMutation& mutation);
    AbstractMutation(const JSONAbstractObject& mutation);

    bool isObjectMutation() const;
    bool isArrayMutation() const;
    bool isTextMutation() const;
    bool isSkipMutation() const;
    bool isDeleteMutation() const;
    bool isLiftMutation() const;
    bool isSqueezeMutation() const;
    bool isInsertMutation() const;
    bool isRichTextMutation() const;

    ObjectMutation toObjectMutation() const;
    ArrayMutation toArrayMutation() const;
    TextMutation toTextMutation() const;
    SkipMutation toSkipMutation() const;
    DeleteMutation toDeleteMutation() const;
    LiftMutation toLiftMutation() const;
    SqueezeMutation toSqueezeMutation() const;
    InsertMutation toInsertMutation() const;
    RichTextMutation toRichTextMutation() const;
};

#endif // ABSTRACTMUTATION_H
