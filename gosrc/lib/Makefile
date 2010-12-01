include $(GOROOT)/src/Make.inc

TARG=lightwave
GOFILES=\
	node.go \
	uid.go \
	ot.go \
	otapply.go \
	session.go \
	discovery.go \
	federation_proxy.go \
	federation_gateway.go \
	userid.go \
	history.go \
	capabilities_store.go \
	staticnode.go

GC += -I ../wave/_obj
LD += -L ../wave/_obj

include $(GOROOT)/src/Make.pkg
