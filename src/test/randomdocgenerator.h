#ifndef RANDOMDOCGENERATOR_H
#define RANDOMDOCGENERATOR_H

#include "json/jsonobject.h"

class FormatMutation;

class RandomDocGenerator
{
public:
    RandomDocGenerator(int depth);

    JSONObject createObject(int depth);
    QString createString();
    JSONArray createArray(int depth);
    JSONObject createRichText(int depth);

private:
    int m_depth;
    FormatMutation createFormatMutation();
};

#endif // RANDOMDOCGENERATOR_H
