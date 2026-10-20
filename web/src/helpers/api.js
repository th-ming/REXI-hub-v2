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
import { getUserIdFromLocalStorage, showError, translateBackend } from './utils';
import axios from 'axios';

export let API = axios.create({
  baseURL: import.meta.env.VITE_REACT_APP_SERVER_URL
    ? import.meta.env.VITE_REACT_APP_SERVER_URL
    : '',
  headers: {
    'REXI-hub-v2-User': getUserIdFromLocalStorage(),
    'Cache-Control': 'no-store',
  },
});

export function updateAPI() {
  API = axios.create({
    baseURL: import.meta.env.VITE_REACT_APP_SERVER_URL
      ? import.meta.env.VITE_REACT_APP_SERVER_URL
      : '',
    headers: {
      'REXI-hub-v2-User': getUserIdFromLocalStorage(),
      'Cache-Control': 'no-store',
    },
  });
}

API.interceptors.response.use(
  (response) => {
    const data = response.data;
    if (
      data &&
      typeof data === 'object' &&
      data.success === false &&
      typeof data.message === 'string'
    ) {
      data.message = translateBackend(data.message);
    }
    return response;
  },
  (error) => {
    const data = error.response && error.response.data;
    if (data && typeof data.message === 'string') {
      data.message = translateBackend(data.message);
    }
    showError(error);
  },
);
