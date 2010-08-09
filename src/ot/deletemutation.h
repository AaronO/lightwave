#ifndef DELETEMUTATION_H
#define DELETEMUTATION_H

#include "abstractmutation.h"

class DeleteMutation : public AbstractMutation
{
public:
    DeleteMutation();
    DeleteMutation(int count);
    DeleteMutation( const JSONAbstractObject& mutation );
    DeleteMutation( const DeleteMutation& mutation );

    int count() const;
    void setCount( int count );
};

#endif // DELETEMUTATION_H
