include $(GOROOT)/src/Make.inc

TARG=wave
GOFILES=\
	wave.pb.go\
	waveot.go \
	wavenode.go \
	jsonprotobuf.go

GC += -I ../lib/_obj
LD += -L ../lib/_obj

include $(GOROOT)/src/Make.pkg
include $(GOROOT)/src/pkg/goprotobuf.googlecode.com/hg/Make.protobuf