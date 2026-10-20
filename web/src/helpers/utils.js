/*
Copyright (c) 2025 REXI-hub team

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
import i18n from '../i18n/i18n';
import { Toast } from '@douyinfe/semi-ui';
import { toastConstants } from '../constants';
import React from 'react';
import { toast } from 'react-toastify';
import backendZh from '../i18n/locales/backend-zh.json';
import backendVi from '../i18n/locales/backend-vi.json';

const backendMaps = { zh: backendZh, vi: backendVi };
const backendTemplates = {};

function buildBackendTemplates(lang) {
  if (backendTemplates[lang]) {
    return backendTemplates[lang];
  }
  const map = backendMaps[lang];
  if (!map) {
    return null;
  }
  const list = [];
  for (const key of Object.keys(map)) {
    if (!/%[sdvwqf]/.test(key)) {
      continue;
    }
    const source = key
      .replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
      .replace(/%[sdvwqf]/g, '(.*?)');
    list.push([new RegExp('^' + source + '$'), map[key]]);
  }
  list.sort((a, b) => b[0].source.length - a[0].source.length);
  backendTemplates[lang] = list;
  return list;
}

function fillTemplate(translated, captures) {
  let i = 0;
  return translated.replace(/%[sdvwqf]/g, () =>
    i < captures.length ? captures[i++] : '',
  );
}

export function translateBackend(text) {
  if (typeof text !== 'string' || text.length === 0 || text.length > 500) {
    return text;
  }
  const lang = (i18n.language || 'en').split('-')[0];
  const map = backendMaps[lang];
  if (!map) {
    return text;
  }
  if (Object.prototype.hasOwnProperty.call(map, text)) {
    return map[text];
  }
  const idx = text.indexOf(': ');
  if (idx > 0) {
    const head = text.slice(0, idx);
    if (Object.prototype.hasOwnProperty.call(map, head)) {
      return map[head] + text.slice(idx);
    }
  }
  const templates = buildBackendTemplates(lang);
  if (templates) {
    for (const [re, translated] of templates) {
      const m = text.match(re);
      if (m) {
        return fillTemplate(translated, m.slice(1));
      }
    }
  }
  return text;
}

const HTMLToastContent = ({ htmlContent }) => {
  return <div dangerouslySetInnerHTML={{ __html: htmlContent }} />;
};
export default HTMLToastContent;
export function isAdmin() {
  let user = localStorage.getItem('user');
  if (!user) return false;
  user = JSON.parse(user);
  return user.role >= 10;
}

export function isRoot() {
  let user = localStorage.getItem('user');
  if (!user) return false;
  user = JSON.parse(user);
  return user.role >= 100;
}

export function getSystemName() {
  let system_name = localStorage.getItem('system_name');
  if (!system_name) return 'REXI-hub';
  return system_name;
}

export function getLogo() {
  let logo = localStorage.getItem('logo');
  if (!logo) return '/logo.png';
  return logo;
}

export function getUserIdFromLocalStorage() {
  let user = localStorage.getItem('user');
  if (!user) return -1;
  user = JSON.parse(user);
  return user.id;
}

export function getFooterHTML() {
  return localStorage.getItem('footer_html');
}

export async function copy(text) {
  let okay = true;
  try {
    await navigator.clipboard.writeText(text);
  } catch (e) {
    try {
      // Dựng input, thực thi lệnh sao chép
      var _input = window.document.createElement('input');
      _input.value = text;
      window.document.body.appendChild(_input);
      _input.select();
      window.document.execCommand('Copy');
      window.document.body.removeChild(_input);
    } catch (e) {
      okay = false;
      console.error(e);
    }
  }
  return okay;
}

export function isMobile() {
  return window.innerWidth <= 600;
}

let showErrorOptions = { autoClose: toastConstants.ERROR_TIMEOUT };
let showWarningOptions = { autoClose: toastConstants.WARNING_TIMEOUT };
let showSuccessOptions = { autoClose: toastConstants.SUCCESS_TIMEOUT };
let showInfoOptions = { autoClose: toastConstants.INFO_TIMEOUT };
let showNoticeOptions = { autoClose: false };

if (isMobile()) {
  showErrorOptions.position = 'top-center';
  // showErrorOptions.transition = 'flip';

  showSuccessOptions.position = 'top-center';
  // showSuccessOptions.transition = 'flip';

  showInfoOptions.position = 'top-center';
  // showInfoOptions.transition = 'flip';

  showNoticeOptions.position = 'top-center';
  // showNoticeOptions.transition = 'flip';
}

export function showError(error) {
  console.error(error);
  if (error.message) {
    if (error.name === 'AxiosError') {
      switch (error.response.status) {
        case 401:
          window.localStorage.removeItem('user')
          // toast.error('Lỗi: chưa đăng nhập hoặc phiên đăng nhập đã hết hạn, vui lòng đăng nhập lại!', showErrorOptions);
          window.location.href = `/login?expired=true&returnTo=${encodeURIComponent(window.location.pathname + window.location.search)}`;
          break;
        case 429:
          Toast.error(i18n.t('错误：请求次数过多，请稍后再试！'));
          break;
        case 403:
          const cfMitigated = error.response.headers['cf-mitigated'];
          if (cfMitigated === 'challenge') {
            Toast.warning(i18n.t('Please refresh the page! Cloudflare needs to verify your browser.'));
          } else {
            Toast.error(i18n.t('Error: Insufficient permissions, please contact the administrator!'));
          }
          break;
        case 500:
          Toast.error(i18n.t('错误：服务器内部错误，请联系管理员！'));
          break;
        case 405:
          Toast.info(i18n.t('This site is for demo purposes only, no server!'));
          break;
        default: {
          const data = error.response && error.response.data;
          const msg =
            data && (data.message || (data.error && data.error.message));
          if (msg) {
            Toast.error(translateBackend(msg));
          } else {
            Toast.error('Error: ' + error.message);
          }
        }
      }
      return;
    }
    Toast.error('Error: ' + error.message);
  } else if (typeof error === 'string') {
    Toast.error(translateBackend(error));
  } else {
    Toast.error('Error: ' + error);
  }
}

export function showWarning(message) {
  Toast.warning(translateBackend(message));
}

export function showSuccess(message) {
  Toast.success(translateBackend(message));
}

export function showInfo(message) {
  Toast.info(translateBackend(message));
}

export function showNotice(message, isHTML = false) {
  if (isHTML) {
    toast(<HTMLToastContent htmlContent={message} />, showNoticeOptions);
  } else {
    Toast.info(translateBackend(message));
  }
}

export function openPage(url) {
  window.open(url);
}

export function removeTrailingSlash(url) {
  if (url.endsWith('/')) {
    return url.slice(0, -1);
  } else {
    return url;
  }
}

export function getTodayStartTimestamp() {
  var now = new Date();
  now.setHours(0, 0, 0, 0);
  return Math.floor(now.getTime() / 1000);
}

export function timestamp2string(timestamp) {
  let date = new Date(timestamp * 1000);
  let year = date.getFullYear().toString();
  let month = (date.getMonth() + 1).toString();
  let day = date.getDate().toString();
  let hour = date.getHours().toString();
  let minute = date.getMinutes().toString();
  let second = date.getSeconds().toString();
  if (month.length === 1) {
    month = '0' + month;
  }
  if (day.length === 1) {
    day = '0' + day;
  }
  if (hour.length === 1) {
    hour = '0' + hour;
  }
  if (minute.length === 1) {
    minute = '0' + minute;
  }
  if (second.length === 1) {
    second = '0' + second;
  }
  return (
    year + '-' + month + '-' + day + ' ' + hour + ':' + minute + ':' + second
  );
}

export function timestamp2string1(timestamp, dataExportDefaultTime = 'hour') {
  let date = new Date(timestamp * 1000);
  // let year = date.getFullYear().toString();
  let month = (date.getMonth() + 1).toString();
  let day = date.getDate().toString();
  let hour = date.getHours().toString();
  if (day === '24') {
    console.log('timestamp', timestamp);
  }
  if (month.length === 1) {
    month = '0' + month;
  }
  if (day.length === 1) {
    day = '0' + day;
  }
  if (hour.length === 1) {
    hour = '0' + hour;
  }
  let str = month + '-' + day;
  if (dataExportDefaultTime === 'hour') {
    str += ' ' + hour + ':00';
  } else if (dataExportDefaultTime === 'week') {
    let nextWeek = new Date(timestamp * 1000 + 6 * 24 * 60 * 60 * 1000);
    let nextMonth = (nextWeek.getMonth() + 1).toString();
    let nextDay = nextWeek.getDate().toString();
    if (nextMonth.length === 1) {
      nextMonth = '0' + nextMonth;
    }
    if (nextDay.length === 1) {
      nextDay = '0' + nextDay;
    }
    str += ' - ' + nextMonth + '-' + nextDay;
  }
  return str;
}

export function downloadTextAsFile(text, filename) {
  let blob = new Blob([text], { type: 'text/plain;charset=utf-8' });
  let url = URL.createObjectURL(blob);
  let a = document.createElement('a');
  a.href = url;
  a.download = filename;
  a.click();
}

export const verifyJSON = (str) => {
  try {
    JSON.parse(str);
  } catch (e) {
    return false;
  }
  return true;
};

export function verifyJSONPromise(value) {
  try {
    JSON.parse(value);
    return Promise.resolve();
  } catch (e) {
    return Promise.reject(i18n.t('Invalid JSON string'));
  }
}

export function shouldShowPrompt(id) {
  let prompt = localStorage.getItem(`prompt-${id}`);
  return !prompt;
}

export function setPromptShown(id) {
  localStorage.setItem(`prompt-${id}`, 'true');
}

/**
 * So sánh thuộc tính của hai đối tượng, tìm ra thuộc tính đã thay đổi và trả về mảng chứa thông tin thuộc tính thay đổi
 * @param {Object} oldObject - Đối tượng cũ
 * @param {Object} newObject - Đối tượng mới
 * @return {Array} Mảng chứa thông tin thuộc tính đã thay đổi, mỗi phần tử là một đối tượng gồm key, oldValue và newValue
 */
export function compareObjects(oldObject, newObject) {
  const changedProperties = [];

  // So sánh thuộc tính của hai đối tượng
  for (const key in oldObject) {
    if (oldObject.hasOwnProperty(key) && newObject.hasOwnProperty(key)) {
      if (oldObject[key] !== newObject[key]) {
        changedProperties.push({
          key: key,
          oldValue: oldObject[key],
          newValue: newObject[key],
        });
      }
    }
  }

  return changedProperties;
}
