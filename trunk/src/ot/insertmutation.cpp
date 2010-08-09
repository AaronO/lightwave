#include "insertmutation.h"
#include "json/jsonconstant.h"

InsertMutation::InsertMutation()
{
}

InsertMutation::InsertMutation(const QString& text)
{
    JSONConstant::ConstantData* d = new JSONConstant::ConstantData();
    m_data = d;
    d->variant.setValue(text);
}

InsertMutation::InsertMutation(const JSONAbstractObject& mutation)
    : AbstractMutation(mutation)
{
    if ( !isInsertMutation() )
        clear();
}

InsertMutation::InsertMutation(const InsertMutation& mutation)
    : AbstractMutation(mutation)
{
}
