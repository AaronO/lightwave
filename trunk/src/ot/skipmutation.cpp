#include "skipmutation.h"
#include "json/jsonobject.h"

SkipMutation::SkipMutation()
{
}

SkipMutation::SkipMutation(int skip)
{
    setCount(skip);
}


SkipMutation::SkipMutation(const JSONAbstractObject& mutation)
    : AbstractMutation(mutation)
{
    if ( !isSkipMutation() )
        clear();
}

SkipMutation::SkipMutation(const SkipMutation& mutation)
    : AbstractMutation(mutation)
{
}

int SkipMutation::count() const
{
    return toObject().attribute("$skip").toInt();
}

void SkipMutation::setCount( int count )
{
    if ( isNull() )
        becomeObject();
    toObject().setAttribute("$skip", count);
}
