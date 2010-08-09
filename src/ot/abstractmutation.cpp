#include "abstractmutation.h"
#include "objectmutation.h"
#include "arraymutation.h"
#include "textmutation.h"
#include "skipmutation.h"
#include "deletemutation.h"
#include "liftmutation.h"
#include "squeezemutation.h"
#include "insertmutation.h"
#include "richtextmutation.h"
#include "json/jsonobject.h"

AbstractMutation::AbstractMutation()
{
}

AbstractMutation::AbstractMutation(const AbstractMutation& mutation)
    : JSONAbstractObject(mutation)
{
}

AbstractMutation::AbstractMutation(const JSONAbstractObject& mutation)
    : JSONAbstractObject(mutation)
{
}

bool AbstractMutation::isObjectMutation() const
{
    return !toObject().attribute("_object").isNull();
}

bool AbstractMutation::isArrayMutation() const
{
    return !toObject().attribute("_array").isNull();
}

bool AbstractMutation::isTextMutation() const
{
    return !toObject().attribute("_text").isNull();
}

bool AbstractMutation::isRichTextMutation() const
{
    return !toObject().attribute("_richtext").isNull();
}

bool AbstractMutation::isSkipMutation() const
{
    return !toObject().attribute("_skip").isNull();
}

bool AbstractMutation::isDeleteMutation() const
{
    return !toObject().attribute("_delete").isNull();
}

bool AbstractMutation::isLiftMutation() const
{
    return !toObject().attribute("_lift").isNull();
}

bool AbstractMutation::isSqueezeMutation() const
{
    return !toObject().attribute("_squeeze").isNull();
}

bool AbstractMutation::isInsertMutation() const
{
    if ( isConstant() )
        return true;
    if ( isArray() )
        return true;

    JSONObject obj = toObject();
    if ( obj.isNull() )
        return false;
    if ( !obj.attribute("_object").isNull() )
        return false;
    if ( !obj.attribute("_array").isNull() )
        return false;
    if ( !obj.attribute("_text").isNull() )
        return false;
    if ( !obj.attribute("_richtext").isNull() )
        return false;
    if ( !obj.attribute("_skip").isNull() )
        return false;
    if ( !obj.attribute("_delete").isNull() )
        return false;
    if ( !obj.attribute("_lift").isNull() )
        return false;
    if ( !obj.attribute("_squeeze").isNull() )
        return false;
    return true;
}

ObjectMutation AbstractMutation::toObjectMutation() const
{
    return ObjectMutation(*this);
}

ArrayMutation AbstractMutation::toArrayMutation() const
{
    return ArrayMutation(*this);
}

TextMutation AbstractMutation::toTextMutation() const
{
    return TextMutation(*this);
}

RichTextMutation AbstractMutation::toRichTextMutation() const
{
    return RichTextMutation(*this);
}

SkipMutation AbstractMutation::toSkipMutation() const
{
    return SkipMutation(*this);
}

DeleteMutation AbstractMutation::toDeleteMutation() const
{
    return DeleteMutation(*this);
}

LiftMutation AbstractMutation::toLiftMutation() const
{
    return LiftMutation(*this);
}

SqueezeMutation AbstractMutation::toSqueezeMutation() const
{
    return SqueezeMutation(*this);
}

InsertMutation AbstractMutation::toInsertMutation() const
{
    return InsertMutation(*this);
}
