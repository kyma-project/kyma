const http = require('http');
const server = http.createServer();

server.on('request', (request, response) => {
    let body = [];
    request.on('data', (chunk) => {
        body.push(chunk);
    }).on('end', () => {
        body = Buffer.concat(body).toString();

	console.log(`==== ${request.method} ${request.url}`);
	console.log('> Headers');
        console.log(request.headers);

	console.log('> Body');
	console.log(body);
        response.writeHead(200, {"Content-Type": "application/json"});
        response.end(JSON.stringify({
    		"token_type": "Bearer",
    		"access_token": "122-b012b9bd-0073-4415-0b2b-f06c36cc4031",
    		"expires_in": 3600,
    		"scope": ""
	}));
    });
}).listen(8084);
