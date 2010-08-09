#ifndef OBJECTMUTATION_H
#define OBJECTMUTATION_H

#include "abstractmutation.h"

class ObjectMutation : public AbstractMutation
{
public:
    ObjectMutation();
    ObjectMutation(bool create_empty);
    ObjectMutation(const JSONAbstractObject& mutation);
    ObjectMutation(const ObjectMutation& mutation);

    void setMutation( const QString& name, const AbstractMutation& m );
    AbstractMutation mutation(const QString& name);
};

#endif // OBJECTMUTATION_H
