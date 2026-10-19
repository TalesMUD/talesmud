import http from 'node:http';
import https from 'node:https';
import fs from 'node:fs';
import path from 'node:path';
import {createRequire} from 'node:module';
import {fixtureHttp, fixtureSocket} from './uxTopFixtureBackend.mjs';

const require = createRequire(process.env.UX_PROXY_PACKAGE || new URL('../package.json',import.meta.url));
const {WebSocket, WebSocketServer} = require('ws');
const fixture = process.env.UX_FIXTURE === '1';
const upstreamOrigin = fixture ? null : new URL(process.env.UX_UPSTREAM);
const root = path.resolve(process.env.UX_PUBLIC || 'public/mud-client/public');
const port = Number(process.env.UX_PORT || 8016);
const mime = {'.html':'text/html', '.js':'text/javascript', '.css':'text/css', '.json':'application/json', '.png':'image/png', '.woff2':'font/woff2'};
const server = http.createServer((request,response) => {
  if (fixture && fixtureHttp(request,response)) return;
  if (request.url.startsWith('/api/')) {
    const upstream = https.request(new URL(request.url, upstreamOrigin), {
      method:request.method,
      headers:{...request.headers, host:upstreamOrigin.host, origin:upstreamOrigin.origin},
    }, reply => {
      response.writeHead(reply.statusCode, reply.headers);
      reply.pipe(response);
    });
    upstream.on('error', () => { response.writeHead(502); response.end(); });
    request.pipe(upstream);
    return;
  }
  const name = decodeURIComponent(new URL(request.url,'http://localhost').pathname.replace(/^\/play\/?/,'')) || 'index.html';
  const file = path.resolve(root,name);
  if (!file.startsWith(root+path.sep) || !fs.existsSync(file) || !fs.statSync(file).isFile()) {
    response.writeHead(404); response.end(); return;
  }
  response.writeHead(200, {'content-type':mime[path.extname(file)] || 'application/octet-stream', 'cache-control':'no-store'});
  fs.createReadStream(file).pipe(response);
});
const sockets = new WebSocketServer({noServer:true});
server.on('upgrade',(request,socket,head) => {
  if (!request.url.startsWith('/ws')) { socket.destroy(); return; }
  if (fixture) { sockets.handleUpgrade(request,socket,head,fixtureSocket); return; }
  const upstream = new WebSocket(`${upstreamOrigin.protocol === 'https:' ? 'wss:' : 'ws:'}//${upstreamOrigin.host}${request.url}`, {origin:upstreamOrigin.origin});
  upstream.on('error', () => socket.destroy());
  upstream.once('open', () => sockets.handleUpgrade(request,socket,head, downstream => {
    upstream.on('message',(data,binary) => { if(downstream.readyState===WebSocket.OPEN) downstream.send(data,{binary}); });
    downstream.on('message',(data,binary) => { if(upstream.readyState===WebSocket.OPEN) upstream.send(data,{binary}); });
    upstream.on('close',()=>downstream.close());
    downstream.on('close',()=>upstream.close());
    downstream.on('error',()=>upstream.close());
  }));
});
server.listen(port,'127.0.0.1',()=>console.log(`Local client on http://127.0.0.1:${port}/play/ with ${fixture ? 'local protocol fixture' : 'live guest backend'}`));
