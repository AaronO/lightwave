export GOROOT=/home/weis/go
cd lib; make; cd ..
cd wave; make; cd ..
rm server/lightwave
cd server; make; cd ..
