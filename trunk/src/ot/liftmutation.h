#ifndef LIFTMUTATION_H
#define LIFTMUTATION_H

#include "abstractmutation.h"

class LiftMutation : public AbstractMutation
{
public:
    LiftMutation();
    LiftMutation(const QString& id);
    LiftMutation(const JSONAbstractObject& mutation);
    LiftMutation(const LiftMutation& mutation);

    QString id() const;
    AbstractMutation mutation() const;
    void setMutation( const AbstractMutation& mutation);
    bool hasMutation() const;
};

#endif // LIFTMUTATION_H
