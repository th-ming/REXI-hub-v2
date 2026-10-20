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
import React, { useContext, useEffect, useState } from 'react';
import { Dimmer, Loader, Segment } from 'semantic-ui-react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { API, showError, showSuccess, updateAPI } from '../helpers';
import { UserContext } from '../context/User';
import { setUserData } from '../helpers/data.js';
import i18n from '../i18n/i18n';

const OAuth2Callback = (props) => {
  const [searchParams, setSearchParams] = useSearchParams();

  const [userState, userDispatch] = useContext(UserContext);
  const [prompt, setPrompt] = useState('Processing...');
  const [processing, setProcessing] = useState(true);

  let navigate = useNavigate();

  const sendCode = async (code, state, count) => {
    const res = await API.get(
      `/api/oauth/${props.type}?code=${code}&state=${state}`,
    );
    const { success, message, data } = res.data;
    if (success) {
      if (message === 'bind') {
        showSuccess(i18n.t('Binding successful!'));
        navigate('/admin/settings');
      } else {
        userDispatch({ type: 'login', payload: data });
        localStorage.setItem('user', JSON.stringify(data));
        setUserData(data);
        updateAPI();
        showSuccess(i18n.t('Login successful!'));
        navigate(searchParams.get('returnTo') || '/app/tokens');
      }
    } else {
      showError(message);
      if (count === 0) {
        setPrompt('Operation failed, redirecting to login...');
        navigate('/admin/settings'); // in case this is failed to bind GitHub
        return;
      }
      count++;
      setPrompt(`Error occurred, retry #${count}...`);
      await new Promise((resolve) => setTimeout(resolve, count * 2000));
      await sendCode(code, state, count);
    }
  };

  useEffect(() => {
    let code = searchParams.get('code');
    let state = searchParams.get('state');
    sendCode(code, state, 0).then();
  }, []);

  return (
    <Segment style={{ minHeight: '300px' }}>
      <Dimmer active inverted>
        <Loader size='large'>{prompt}</Loader>
      </Dimmer>
    </Segment>
  );
};

export default OAuth2Callback;
