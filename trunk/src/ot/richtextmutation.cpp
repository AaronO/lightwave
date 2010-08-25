#include "richtextmutation.h"
#include "json/jsonobject.h"
#include "json/jsonarray.h"

RichTextMutation::RichTextMutation()
{
}

RichTextMutation::RichTextMutation(bool create_empty)
{
    if ( create_empty )
    {
        becomeObject();
        toObject().setAttribute("$richtext", JSONArray(true));
    }
}

RichTextMutation::RichTextMutation(const JSONAbstractObject& mutation)
    : AbstractMutation(mutation)
{
    if ( !isRichTextMutation() )
        clear();
}

RichTextMutation::RichTextMutation(const RichTextMutation& mutation)
    : AbstractMutation(mutation)
{
}

JSONArray RichTextMutation::content() const
{
    return toObject().attributeArray("$richtext");
}

