docker load -i exchange2_amd64.tar
docker run -p 40102:40102 --name exchange2 -d exchange2


cd exchanges
docker load -i exchange1_amd64.tar
docker load -i exchange2_amd64.tar
docker load -i exchange3_amd64.tar
cd ..


localhost:8080/prices/latest/exchange1/ETHUSDT