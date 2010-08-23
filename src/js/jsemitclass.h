#ifndef JSEMITCLASS_H
#define JSEMITCLASS_H

#include <QScriptClass>

class JSEngine;

class JSEmitClass : public QScriptClass
{
public:
    JSEmitClass(JSEngine* engine);

    QVariant extension( Extension extension, const QVariant & argument = QVariant() );
    bool supportsExtension ( Extension extension ) const;
};

#endif // JSEMITCLASS_H
