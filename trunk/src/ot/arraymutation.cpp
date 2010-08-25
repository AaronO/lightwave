#include "arraymutation.h"
#include "json/jsonarray.h"
#include "json/jsonobject.h"

ArrayMutation::ArrayMutation()
{
}

ArrayMutation::ArrayMutation(bool create_empty)
{
    if ( create_empty )
    {
        becomeObject();
        toObject().setAttribute("$array", JSONArray(true));
    }
}

ArrayMutation::ArrayMutation(const JSONAbstractObject& mutation)
    : AbstractMutation(mutation)
{
    if ( !isArrayMutation() )
        clear();
}

ArrayMutation::ArrayMutation(const ArrayMutation& mutation)
    : AbstractMutation(mutation)
{
}

JSONArray ArrayMutation::content() const
{
    return toObject().attributeArray("$array");
}

void ArrayMutation::setContent( const JSONArray& arr )
{
    if ( isNull() )
        becomeObject();
    toObject().setAttribute("$array", arr);
}
