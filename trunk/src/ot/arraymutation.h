#ifndef ARRAYMUTATION_H
#define ARRAYMUTATION_H

#include "abstractmutation.h"
#include "json/jsonarray.h"

class ArrayMutation : public AbstractMutation
{
public:
    ArrayMutation();
    ArrayMutation(bool create_empty);
    ArrayMutation(const JSONAbstractObject& mutation);
    ArrayMutation(const ArrayMutation& mutation);

    JSONArray content() const;
    void setContent( const JSONArray& arr );
};

#endif // ARRAYMUTATION_H
