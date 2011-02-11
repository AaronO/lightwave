include $(GOROOT)/src/Make.inc

TARG=lightwave
GOFILES=\
	node.go \
	uri.go \
	ot.go \
	otapply.go \
	session.go \
	discovery.go \
	federation_proxy.go \
	federation_gateway.go \
	user.go \
	history.go \
	capabilities_store.go \
	staticnode.go \
	config.go \
	indexer.go \
	digest.go

GC += -I ../wave/_obj
LD += -L ../wave/_obj

include $(GOROOT)/src/Make.pkg
