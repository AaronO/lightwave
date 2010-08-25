#include "objectmutation.h"
#include "json/jsonobject.h"

ObjectMutation::ObjectMutation()
{
}

ObjectMutation::ObjectMutation(bool create_empty)
{
    if ( create_empty )
    {
        becomeObject();
        toObject().setAttribute("$object", true);
    }
}

ObjectMutation::ObjectMutation(const JSONAbstractObject& mutation)
    : AbstractMutation(mutation)
{
    if ( !isObjectMutation() )
        clear();
}

ObjectMutation::ObjectMutation(const ObjectMutation& mutation)
    : AbstractMutation(mutation)
{
}

void ObjectMutation::setMutation( const QString& name, const AbstractMutation& m )
{
    if ( isNull() )
    {
        becomeObject();
        toObject().setAttribute("$object", true);
    }
    toObject().setAttribute(name, m);
}

AbstractMutation ObjectMutation::mutation(const QString& name)
{
    return AbstractMutation(toObject().attribute(name));
}
