#include "liftmutation.h"
#include "json/jsonobject.h"

LiftMutation::LiftMutation()
{
}

LiftMutation::LiftMutation(const QString& id)
{
    becomeObject();
    toObject().setAttribute("_lift", id);
}

LiftMutation::LiftMutation(const JSONAbstractObject& mutation)
    : AbstractMutation(mutation)
{
    if ( !isLiftMutation() )
        clear();
}

LiftMutation::LiftMutation(const LiftMutation& mutation)
    : AbstractMutation(mutation)
{
}

QString LiftMutation::id() const
{
    return toObject().attributeString("_lift");
}

AbstractMutation LiftMutation::mutation() const
{
    return AbstractMutation(toObject().attribute("_mutation"));
}

void LiftMutation::setMutation( const AbstractMutation& mutation)
{
    if ( isNull() )
        becomeObject();
    toObject().setAttribute( "_mutation", mutation );
}

bool LiftMutation::hasMutation() const
{
    return !toObject().attribute("_mutation").isNull();
}
