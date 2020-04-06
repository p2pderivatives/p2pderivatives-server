# $1 host parameter ex : localhost
# $2 port parameter
for i in `seq 1 10`;
do
  nc -z $1 $2 && echo Success && exit 0
  echo -n .
  sleep 1
  done
echo Failed waiting for $1:$2 && exit 1