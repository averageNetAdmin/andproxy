rm /home/kopatich/andproxy/dbmng/andproxy.db
cd /home/kopatich/andproxy/dbmng
go build .
./dbmng &
sleep 2s
cd /home/kopatich/andproxy/proxy
go build .
./proxy &
cd /home/kopatich/andproxy/webint
go build ./
./webint

pkill -KILL dbmng
pkill -KILL proxy
pkill -KILL webint