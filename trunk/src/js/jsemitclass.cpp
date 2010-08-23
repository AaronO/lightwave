#include "jsemitclass.h"
#include "jsengine.h"
#include <QVariant>
#include <QScriptContext>

JSEmitClass::JSEmitClass(JSEngine* engine)
    : QScriptClass(engine)
{
}

QVariant JSEmitClass::extension( Extension extension, const QVariant& argument )
{
    Q_UNUSED(argument);

    if (extension == Callable)
    {
        // QScriptContext *context = qvariant_cast<QScriptContext*>(argument);
        QScriptContext *context = engine()->currentContext();
        if ( context->argumentCount() != 2)
        {
            context->throwError("emit() accepts exactly two parameters");
            return QVariant();
        }
        if ( !context->argument(0).isArray() )
        {
            context->throwError("The first argument to emit() must be an array");
            return QVariant();
        }
        qDebug("emit(%s,%s)", qPrintable(context->argument(0).toString()), qPrintable(context->argument(1).toString()));
        return QVariant(true);
    }
    return QVariant();
}

bool JSEmitClass::supportsExtension ( Extension extension ) const
{
    if ( extension == Callable )
        return true;
    return false;
}
