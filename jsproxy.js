"use strict";


var HOST_PORT = parseInt(process.argv[2]);
var REMOTE_IP = process.argv[3];
var REMOTE_PORT = parseInt(process.argv[4]);
console.log(process.argv);


var url = require("url");
var http = require("http");

// Function where we filter the required header to mitigate server blocks
function filterRequestHeaders(request){
    request.headers.host = REMOTE_IP;
    if(request.headers.referer != undefined){
        try{
            let url =  new URL(request.headers.referer);
            url.hostname = REMOTE_IP;
            request.headers.referer = url.href;
        }catch(err){}
    }

    return request;
}

// Function applies filter and forward request
function handleServerRequests(request, response){
    let copyRequestOptions = url.parse(`http://${REMOTE_IP}:${REMOTE_PORT}${request.url}`);
    copyRequestOptions.method = request.method;
    copyRequestOptions.headers = request.headers;
    copyRequestOptions = filterRequestHeaders(copyRequestOptions);

    let copyRequest = http.request(copyRequestOptions);
    copyRequest.addListener("response", function (localResponse) {
        localResponse.pipe(response);
        response.writeHead(
            localResponse.statusCode,
            localResponse.headers,
        );
    });

    request.pipe(copyRequest);
}


var server = http.createServer(handleServerRequests);
server.listen(HOST_PORT);
// Guarantee server is closed if program crashes or is finished
process.on('uncaughtException',(()=>{}));
process.on('SIGTERM', server.close);


