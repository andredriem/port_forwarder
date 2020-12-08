#docker build -t port_forwarder .
docker run -it -p 3000:3000 --expose=7000-8000 -e SERVER_PORT=3000 -e ALLOW_PERMANENT_RULES=false -e EXPOSED_PORT_START_RANGE=7000 -e EXPOSED_PORT_END_RANGE=8000 --cap-add=NET_ADMIN --cap-add=NET_RAW test-container 

