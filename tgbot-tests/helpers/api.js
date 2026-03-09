/**
 * Direct backend API helper for setting up test state.
 * Used to create/verify data without going through the Telegram bot.
 */

const axios = require('axios');
const crypto = require('crypto');

const BACKEND_URL = 'http://127.0.0.1:8080';
const JWT_SECRET = 'dev-secret-change-me';

class BackendAPI {
  constructor(baseURL = BACKEND_URL) {
    this.baseURL = baseURL;
  }

  /** Derive password the same way the Go bot does: HMAC-SHA256("tgbot:{id}", secret) base64 */
  derivePassword(telegramId) {
    const hmac = crypto.createHmac('sha256', JWT_SECRET);
    hmac.update(`tgbot:${telegramId}`);
    return hmac.digest('base64');
  }

  /** Register a user directly on the backend (mimics bot auth) */
  async register(telegramId, displayName) {
    const email = `tg_${telegramId}@telegram.local`;
    const password = this.derivePassword(telegramId);
    try {
      const resp = await axios.post(`${this.baseURL}/api/auth/register`, {
        email,
        displayName,
        password,
      });
      return resp.data;
    } catch (e) {
      // If already registered, login instead
      if (e.response?.status === 409 || e.response?.data?.error?.includes('already exists')) {
        return this.login(telegramId);
      }
      throw e;
    }
  }

  /** Login a user directly */
  async login(telegramId) {
    const email = `tg_${telegramId}@telegram.local`;
    const password = this.derivePassword(telegramId);
    const resp = await axios.post(`${this.baseURL}/api/auth/login`, {
      email,
      password,
    });
    return resp.data;
  }

  /** Get auth token for a telegram user */
  async getToken(telegramId, displayName = 'Test User') {
    try {
      const resp = await this.login(telegramId);
      return resp.token;
    } catch (e) {
      const resp = await this.register(telegramId, displayName);
      return resp.token;
    }
  }

  /** Create a meeting directly via API */
  async createMeeting(token, input) {
    const resp = await axios.post(`${this.baseURL}/api/meetings/`, input, {
      headers: { Authorization: `Bearer ${token}` },
    });
    return resp.data;
  }

  /** Get a meeting */
  async getMeeting(token, meetingId) {
    const resp = await axios.get(`${this.baseURL}/api/meetings/${meetingId}/`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    return resp.data;
  }

  /** List my meetings */
  async listMyMeetings(token) {
    const resp = await axios.get(`${this.baseURL}/api/meetings/`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    return resp.data;
  }

  /** Confirm a meeting */
  async confirmMeeting(token, meetingId, slotId = null) {
    const body = slotId ? { timeSlotId: slotId } : {};
    const resp = await axios.post(
      `${this.baseURL}/api/meetings/${meetingId}/confirm`,
      body,
      { headers: { Authorization: `Bearer ${token}` } }
    );
    return resp.data;
  }

  /** Vote on time slots */
  async vote(token, meetingId, slotIds) {
    await axios.post(
      `${this.baseURL}/api/meetings/${meetingId}/votes`,
      { timeSlotIds: slotIds },
      { headers: { Authorization: `Bearer ${token}` } }
    );
  }

  /** Delete a meeting */
  async deleteMeeting(token, meetingId) {
    await axios.delete(`${this.baseURL}/api/meetings/${meetingId}/`, {
      headers: { Authorization: `Bearer ${token}` },
    });
  }

  /** Search users */
  async searchUsers(token, query) {
    const resp = await axios.get(
      `${this.baseURL}/api/users/search?q=${encodeURIComponent(query)}&limit=10`,
      { headers: { Authorization: `Bearer ${token}` } }
    );
    return resp.data;
  }

  /** Update RSVP status */
  async updateRSVP(token, meetingId, status) {
    const resp = await axios.put(
      `${this.baseURL}/api/meetings/${meetingId}/participants/rsvp`,
      { status },
      { headers: { Authorization: `Bearer ${token}` } }
    );
    return resp.data;
  }

  /** List all meetings (public + participated) */
  async listAllMeetings(token) {
    const resp = await axios.get(`${this.baseURL}/api/meetings/all`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    return resp.data;
  }

  /**
   * Assert meeting state matches expected properties.
   * Fetches the meeting and checks each key in expectedProps.
   */
  async assertMeetingState(token, meetingId, expectedProps) {
    const meeting = await this.getMeeting(token, meetingId);
    for (const [key, value] of Object.entries(expectedProps)) {
      if (typeof value === 'function') {
        // Custom assertion function
        value(meeting[key], meeting);
      } else if (Array.isArray(value)) {
        expect(meeting[key]).toEqual(expect.arrayContaining(value));
      } else {
        expect(meeting[key]).toBe(value);
      }
    }
    return meeting;
  }
}

module.exports = { BackendAPI };
