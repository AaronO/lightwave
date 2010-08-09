#include "deletemutation.h"
#include "json/jsonobject.h"

DeleteMutation::DeleteMutation()
{
}

DeleteMutation::DeleteMutation(int count)
{
    setCount(count);
}

DeleteMutation::DeleteMutation( const JSONAbstractObject& mutation )
    : AbstractMutation(mutation)
{
    if ( !isDeleteMutation() )
        clear();
}

DeleteMutation::DeleteMutation( const DeleteMutation& mutation )
    : AbstractMutation(mutation)
{
}

int DeleteMutation::count() const
{
    return toObject().attribute("_delete").toInt();
}

void DeleteMutation::setCount( int count )
{
    if ( isNull() )
        becomeObject();
    toObject().setAttribute("_delete", count);
}
