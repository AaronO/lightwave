#ifndef WAVEROOTDOCUMENT_H
#define WAVEROOTDOCUMENT_H

#include "wavedocument.h"
#include "fcgi/fcgirequest.h"
#include <QString>

class WaveContainer;

class WaveMetaDocument : public WaveDocument
{
public:
    WaveMetaDocument(WaveContainer* container, const QString& docId);

    bool addDocument(WaveDocument* wdoc);
    bool addContainer(WaveContainer* c);
};

#endif // WAVEROOTDOCUMENT_H
