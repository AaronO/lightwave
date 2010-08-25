#include "squeezemutation.h"
#include "json/jsonobject.h"

SqueezeMutation::SqueezeMutation()
{
}

SqueezeMutation::SqueezeMutation(const QString& id)
{
    becomeObject();
    toObject().setAttribute("$squeeze", id);
}

SqueezeMutation::SqueezeMutation(const JSONAbstractObject& mutation)
    : AbstractMutation(mutation)
{
    if ( !isSqueezeMutation() )
        clear();
}

SqueezeMutation::SqueezeMutation(const SqueezeMutation& mutation)
    : AbstractMutation(mutation)
{
}

QString SqueezeMutation::id() const
{
    return toObject().attributeString("$squeeze");
}
