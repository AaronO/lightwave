#include "textmutation.h"
#include "json/jsonobject.h"
#include "json/jsonarray.h"

TextMutation::TextMutation()
{
}

TextMutation::TextMutation(bool create_empty)
{
    if ( create_empty )
    {
        becomeObject();
        toObject().setAttribute("_text", JSONArray(true));
    }
}

TextMutation::TextMutation(const JSONAbstractObject& mutation)
    : AbstractMutation(mutation)
{
    if ( !isTextMutation() )
        clear();
}

TextMutation::TextMutation(const TextMutation& mutation)
    : AbstractMutation(mutation)
{
}

JSONArray TextMutation::content() const
{
    return toObject().attributeArray("_text");
}
