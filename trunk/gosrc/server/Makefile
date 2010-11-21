include $(GOROOT)/src/Make.inc

TARG=lightwave
GOFILES=\
	main.go

GC += -I ../lib/_obj
LD += -L ../lib/_obj

GC += -I ../wave/_obj
LD += -L ../wave/_obj

include $(GOROOT)/src/Make.cmd
