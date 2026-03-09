/**
 * BotTestHarness - manages fake Telegram server and bot container lifecycle.
 *
 * Architecture:
 * - telegram-test-api runs on host as a fake Telegram API server (internal port)
 * - A proxy server sits in front to fix Go telebot.v3 quirks (bot-facing port)
 * - The Go bot runs as a Docker container, pointing at the proxy
 * - Backend (postgres + Go backend) are already running in Docker
 * - Tests send messages via TelegramClient (talks directly to telegram-test-api)
 */

const { createServer } = require('./patched-server');
const { execSync, spawn } = require('child_process');

const BOT_TOKEN = 'test-bot-token-12345';
const TG_INTERNAL_PORT = 9877; // telegram-test-api
const TG_PROXY_PORT = 9876;   // proxy for Go bot
const BACKEND_URL = 'http://127.0.0.1:8080';
const JWT_SECRET = 'dev-secret-change-me';

class BotTestHarness {
  constructor() {
    this.server = null;
    this.botProcess = null;
    this.clients = new Map();
    this.nextUserId = 100;
    this.nextChatId = 100;
  }

  async start() {
    // Start patched Telegram server (telegram-test-api + proxy)
    this.server = createServer(TG_INTERNAL_PORT, TG_PROXY_PORT);
    await this.server.start();
    console.log(`Fake TG server started (proxy: ${this.server.apiURL})`);

    // Stop existing tgbot container if running
    try {
      execSync('docker stop gotogether-tgbot-1 2>/dev/null', { stdio: 'ignore' });
    } catch (e) { /* ignore */ }
    try {
      execSync('docker stop tgbot-test 2>/dev/null', { stdio: 'ignore' });
    } catch (e) { /* ignore */ }

    // Start bot container pointing at our proxy
    this.botProcess = spawn('docker', [
      'run', '--rm',
      '--name', 'tgbot-test',
      '--network', 'host',
      '-e', `TELEGRAM_BOT_TOKEN=${BOT_TOKEN}`,
      '-e', `TELEGRAM_API_URL=${this.server.apiURL}`,
      '-e', `BACKEND_URL=${BACKEND_URL}`,
      '-e', `JWT_SECRET=${JWT_SECRET}`,
      'gotogether-tgbot:latest',
    ], { stdio: ['ignore', 'pipe', 'pipe'] });

    let botOutput = '';
    this.botProcess.stdout.on('data', (d) => {
      botOutput += d.toString();
      if (process.env.DEBUG) console.log('[bot]', d.toString().trim());
    });
    this.botProcess.stderr.on('data', (d) => {
      botOutput += d.toString();
      if (process.env.DEBUG) console.error('[bot]', d.toString().trim());
    });

    // Wait for bot to start and verify it's running
    await this.sleep(3000);

    // Check if bot is still running
    try {
      execSync('docker inspect tgbot-test', { stdio: 'ignore' });
    } catch (e) {
      throw new Error(`Bot failed to start. Output:\n${botOutput}`);
    }

    console.log('Bot container started successfully');
  }

  async stop() {
    try {
      execSync('docker stop tgbot-test 2>/dev/null', { stdio: 'ignore' });
    } catch (e) { /* ignore */ }

    if (this.botProcess) {
      this.botProcess.kill('SIGTERM');
      this.botProcess = null;
    }

    if (this.server) {
      await this.server.stop();
      this.server = null;
    }

    this.clients.clear();
    await this.sleep(500);
  }

  /**
   * Create a test user client with unique IDs.
   */
  createUser(name) {
    const userId = this.nextUserId++;
    const chatId = this.nextChatId++;
    const userName = name.toLowerCase().replace(/\s+/g, '_');

    const client = this.server.getClient(BOT_TOKEN, {
      userId,
      chatId,
      firstName: name,
      userName,
    });

    const botClient = new BotTestClient(client, this.server, userId, chatId, name);
    this.clients.set(name, botClient);
    return botClient;
  }

  sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
  }
}

/**
 * BotTestClient - high-level wrapper for a single test user.
 */
class BotTestClient {
  constructor(client, server, userId, chatId, name) {
    this.client = client;
    this.server = server;
    this.userId = userId;
    this.chatId = chatId;
    this.name = name;
  }

  async sendCommand(command) {
    const msg = this.client.makeCommand(command);
    await this.client.sendCommand(msg);
    await this.sleep(1500);
  }

  async sendMessage(text) {
    const msg = this.client.makeMessage(text);
    await this.client.sendMessage(msg);
    await this.sleep(1500);
  }

  async pressButton(callbackData) {
    const cb = this.client.makeCallbackQuery(callbackData);
    await this.client.sendCallback(cb);
    await this.sleep(1500);
  }

  /**
   * Get all bot responses from the history.
   * Bot messages have chat_id as a string (e.g. "999"), while
   * user messages have chat.id as a number. We only want bot messages.
   */
  async getHistory() {
    const history = await this.client.getUpdatesHistory();
    if (!history) return [];

    const chatIdStr = String(this.chatId);
    return history
      .filter(h => {
        const msg = h.message;
        if (!msg) return false;
        // Bot messages have chat_id as a top-level string field
        // User messages have chat.id as nested number field but no chat_id
        return String(msg.chat_id) === chatIdStr && !msg.from;
      })
      .map(h => ({
        text: h.message?.text || '',
        reply_markup: h.message?.reply_markup
          ? (typeof h.message.reply_markup === 'string'
            ? JSON.parse(h.message.reply_markup)
            : h.message.reply_markup)
          : null,
        chat_id: h.message?.chat_id,
      }));
  }

  async getLastResponse(nth = 0) {
    const history = await this.getHistory();
    const relevant = history.filter(h => h.text);
    if (relevant.length === 0) return null;
    const idx = relevant.length - 1 - nth;
    return idx >= 0 ? relevant[idx] : null;
  }

  async getLastResponses(n) {
    const history = await this.getHistory();
    const relevant = history.filter(h => h.text);
    return relevant.slice(-n);
  }

  async expectResponseContaining(substring, nth = 0) {
    const resp = await this.getLastResponse(nth);
    if (!resp) {
      throw new Error(`No bot response found (expected "${substring}")`);
    }
    if (!resp.text.includes(substring)) {
      throw new Error(
        `Expected response to contain "${substring}" but got:\n${resp.text}`
      );
    }
    return resp;
  }

  async findButton(textPattern) {
    const history = await this.getHistory();
    for (let i = history.length - 1; i >= 0; i--) {
      const markup = history[i].reply_markup;
      if (!markup?.inline_keyboard) continue;
      for (const row of markup.inline_keyboard) {
        for (const btn of row) {
          if (btn.text && btn.text.includes(textPattern)) {
            return btn.callback_data;
          }
        }
      }
    }
    return null;
  }

  async pressButtonByText(textPattern) {
    const data = await this.findButton(textPattern);
    if (!data) {
      const history = await this.getHistory();
      const buttons = [];
      for (const msg of history) {
        if (msg.reply_markup?.inline_keyboard) {
          for (const row of msg.reply_markup.inline_keyboard) {
            for (const btn of row) {
              buttons.push(`"${btn.text}" -> ${btn.callback_data}`);
            }
          }
        }
      }
      throw new Error(
        `Button "${textPattern}" not found. Available buttons:\n${buttons.join('\n')}`
      );
    }
    await this.pressButton(data);
    return data;
  }

  /**
   * Poll until a condition on the last response is met, or timeout.
   * @param {Function} conditionFn - receives the last response, returns truthy if condition met
   * @param {number} timeout - max time in ms (default 5000)
   * @param {number} interval - polling interval in ms (default 300)
   */
  async waitFor(conditionFn, timeout = 5000, interval = 300) {
    const start = Date.now();
    while (Date.now() - start < timeout) {
      const resp = await this.getLastResponse();
      if (resp && conditionFn(resp)) return resp;
      await this.sleep(interval);
    }
    throw new Error(`waitFor timed out after ${timeout}ms`);
  }

  /**
   * Find a button matching a pattern from the Nth-last response's keyboard.
   * @param {string} textPattern - partial text to match against button labels
   * @param {number} nth - 0 = last response, 1 = second-to-last, etc.
   */
  async findButtonInResponse(textPattern, nth = 0) {
    const resp = await this.getLastResponse(nth);
    if (!resp?.reply_markup?.inline_keyboard) return null;
    for (const row of resp.reply_markup.inline_keyboard) {
      for (const btn of row) {
        if (btn.text && btn.text.includes(textPattern)) {
          return btn.callback_data;
        }
      }
    }
    return null;
  }

  /**
   * Get all buttons from the most recent keyboard.
   */
  async getAllButtons() {
    const history = await this.getHistory();
    const buttons = [];
    for (let i = history.length - 1; i >= 0; i--) {
      const markup = history[i].reply_markup;
      if (markup?.inline_keyboard) {
        for (const row of markup.inline_keyboard) {
          for (const btn of row) {
            buttons.push({ text: btn.text, data: btn.callback_data });
          }
        }
        return buttons;
      }
    }
    return buttons;
  }

  sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
  }
}

module.exports = { BotTestHarness, BotTestClient, BOT_TOKEN, BACKEND_URL, JWT_SECRET };
