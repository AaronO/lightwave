#ifndef INSERTMUTATION_H
#define INSERTMUTATION_H

#include "abstractmutation.h"

class InsertMutation : public AbstractMutation
{
public:
    InsertMutation();
    InsertMutation(const QString& text);
    InsertMutation(const JSONAbstractObject& mutation);
    InsertMutation(const InsertMutation& mutation);
};

#endif // INSERTMUTATION_H
