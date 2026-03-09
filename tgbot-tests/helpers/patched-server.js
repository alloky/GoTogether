/**
 * Creates a proxy HTTP server that sits in front of telegram-test-api
 * to handle Go telebot.v3's quirks:
 * 1. Replaces `null` JSON bodies with `{}`
 * 2. Stubs setMyCommands (not supported by telegram-test-api)
 */

const TelegramServer = require('telegram-test-api');
const http = require('http');

function createServer(tgPort, proxyPort) {
  const tgServer = new TelegramServer({ port: tgPort, host: '127.0.0.1' });

  const proxy = http.createServer((req, res) => {
    let body = '';
    req.on('data', chunk => { body += chunk; });
    req.on('end', () => {
      // Handle setMyCommands stub
      if (req.url.includes('/setMyCommands')) {
        res.writeHead(200, { 'Content-Type': 'application/json' });
        res.end(JSON.stringify({ ok: true, result: true }));
        return;
      }

      // Fix null/empty bodies
      let fixedBody = body;
      if (body.trim() === 'null' || body.trim() === '') {
        fixedBody = '{}';
      }

      // Forward to actual telegram-test-api server
      const proxyReq = http.request({
        hostname: '127.0.0.1',
        port: tgPort,
        path: req.url,
        method: req.method,
        headers: {
          ...req.headers,
          'content-length': Buffer.byteLength(fixedBody),
          host: `127.0.0.1:${tgPort}`,
        },
      }, (proxyRes) => {
        res.writeHead(proxyRes.statusCode, proxyRes.headers);
        proxyRes.pipe(res);
      });

      proxyReq.on('error', (err) => {
        res.writeHead(502, { 'Content-Type': 'application/json' });
        res.end(JSON.stringify({ ok: false, error_code: 502, description: err.message }));
      });

      proxyReq.write(fixedBody);
      proxyReq.end();
    });
  });

  return {
    tgServer,
    proxy,
    apiURL: `http://127.0.0.1:${proxyPort}`,

    async start() {
      await tgServer.start();
      await new Promise((resolve, reject) => {
        proxy.listen(proxyPort, '0.0.0.0', () => resolve());
        proxy.on('error', reject);
      });
    },

    async stop() {
      await new Promise(resolve => proxy.close(resolve));
      await tgServer.stop();
    },

    getClient(token, options) {
      return tgServer.getClient(token, options);
    },

    get storage() {
      return tgServer.storage;
    },
  };
}

module.exports = { createServer };
