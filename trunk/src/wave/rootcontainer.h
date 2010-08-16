#ifndef ROOTCONTAINER_H
#define ROOTCONTAINER_H

#include "wavecontainer.h"
#include "waveid.h"
#include "fcgi/fcgirequest.h"
#include "json/jsonobject.h"

class RootContainer : public WaveContainer
{
public:
    RootContainer();

    JSONObject put(FCGI::FCGIRequest* req, const WaveId& waveId);
    JSONObject putFromHostingServer(JSONObject data, const WaveId& waveId);
    JSONObject putFromRemoteServer(JSONObject data, const WaveId& waveId);
    JSONObject get(FCGI::FCGIRequest* req, const WaveId& waveId);

    WaveContainer* container(const WaveId& waveId) const { return const_cast<RootContainer*>(this)->getOrCreateContainer(waveId, false); }

    bool isRemote() const { return false; }

protected:
    virtual WaveContainer* createWaveContainer(const QString& name);

private:
    WaveContainer* getOrCreateContainer(const WaveId& waveId, bool allow_creation);
};

#endif // ROOTCONTAINER_H
