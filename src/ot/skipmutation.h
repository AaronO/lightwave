#ifndef SKIPMUTATION_H
#define SKIPMUTATION_H

#include "abstractmutation.h"

class SkipMutation : public AbstractMutation
{
public:
    SkipMutation();
    SkipMutation(int skip);
    SkipMutation(const JSONAbstractObject& mutation);
    SkipMutation(const SkipMutation& mutation);

    int count() const;
    void setCount( int count );
};

#endif // SKIPMUTATION_H
