#ifndef VIEWCONTAINER_H
#define VIEWCONTAINER_H

#include "wavecontainer.h"
#include <QString>

class RootContainer;
class View;

class ViewContainer : public WaveContainer
{
public:
    ViewContainer(RootContainer* parent, const QString& name);

    bool isRemote() const { return false; }
    bool buildsDigest() const { return false; }

    /**
      * A convenience function for documents().
      */
    QList<View*> views() const;
    /**
      * A convenience function for document().
      */
    View* view(const QString& viewId) const;

protected:
    WaveContainer* createWaveContainer(const QString& name);
    WaveDocument* createDocument(const QString& docId);
    void onDocumentUpdate(WaveDocument* wdoc);
};

#endif // VIEWCONTAINER_H
