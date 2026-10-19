import i18n from '../i18n/i18n';
/*
Copyright (c) 2025 Tethys Plex

This file is part of REXI-hub v2.

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.
*/

// Telegram login fields
export const TELEGRAM_LOGIN_FIELDS = [
  'id',
  'first_name',
  'last_name',
  'username',
  'photo_url',
  'auth_date',
  'hash',
  'lang',
];

// Common form styles
export const FORM_STYLES = {
  container: {
    justifyContent: 'center',
    display: 'flex',
    marginTop: 120,
  },
  card: {
    width: 500,
  },
  linkContainer: {
    display: 'flex',
    justifyContent: 'space-between',
    marginTop: 20,
  },
  thirdPartyContainer: {
    display: 'flex',
    justifyContent: 'center',
    marginTop: 20,
  },
  telegramContainer: {
    display: 'flex',
    justifyContent: 'center',
    marginTop: 5,
  },
  turnstileContainer: {
    display: 'flex',
    justifyContent: 'center',
    marginTop: 20,
  },
  wechatModalContent: {
    display: 'flex',
    alignItem: 'center',
    flexDirection: 'column',
  },
  wechatModalText: {
    textAlign: 'center',
  },
};

// Common validation messages
export const VALIDATION_MESSAGES = {
  TURNSTILE_WAIT: 'Please retry in a few seconds, Turnstile is checking your browser!',
  LOGIN_SUCCESS: 'Login successful!',
  REGISTER_SUCCESS: 'Registration successful!',
  PASSWORD_TOO_SHORT: 'Password must be at least 8 characters!',
  PASSWORD_MISMATCH: i18n.t('The two passwords do not match'),
  MISSING_CREDENTIALS: 'Please enter your username and password!',
  VERIFICATION_CODE_SENT: 'Verification code sent, please check your email!',
  DEFAULT_PASSWORD_WARNING: 'You are using the default password!',
  CHANGE_DEFAULT_PASSWORD: 'Please change the default password immediately!',
  SESSION_EXPIRED: 'Not logged in or session expired, please log in again',
};

// Helper functions
export const getAffCode = () => {
  let statusFromStorage = localStorage.getItem('status');
  if (statusFromStorage) {
    statusFromStorage = JSON.parse(statusFromStorage);
    if (statusFromStorage.aff_enabled === true) {
      let affCode = new URLSearchParams(window.location.search).get('aff');
      if (affCode) {
        localStorage.setItem('aff', affCode);
        return affCode;
      }
      return localStorage.getItem('aff');
    }
  }
  return null;
};

export const getStoredStatus = () => {
  let status = localStorage.getItem('status');
  return status ? JSON.parse(status) : {};
};