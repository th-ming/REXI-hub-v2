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
import i18n from '../../i18n/i18n';
import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { API, isMobile, showError, showSuccess } from '../../helpers';
import Title from '@douyinfe/semi-ui/lib/es/typography/title';
import { Button, Input, SideSheet, Space, Spin } from '@douyinfe/semi-ui';

const AddUser = (props) => {
  const { t } = useTranslation();
  const originInputs = {
    username: '',
    display_name: '',
    password: '',
  };
  const [inputs, setInputs] = useState(originInputs);
  const [loading, setLoading] = useState(false);
  const { username, display_name, password } = inputs;

  const handleInputChange = (name, value) => {
    setInputs((inputs) => ({ ...inputs, [name]: value }));
  };

  const submit = async () => {
    setLoading(true);
    if (inputs.username === '' || inputs.password === '') {
      setLoading(false);
      showError('Username and password cannot be empty!');
      return;
    }
    const res = await API.post(`/api/user/`, inputs);
    const { success, message } = res.data;
    if (success) {
      showSuccess('User account created successfully!');
      setInputs(originInputs);
      props.refresh();
      props.handleClose();
    } else {
      showError(message);
    }
    setLoading(false);
  };

  const handleCancel = () => {
    props.handleClose();
  };

  return (
    <>
      <SideSheet
        placement={'left'}
        title={<Title level={3}>{t('添加用户')}</Title>}
        headerStyle={{ borderBottom: '1px solid var(--semi-color-border)' }}
        bodyStyle={{ borderBottom: '1px solid var(--semi-color-border)' }}
        visible={props.visible}
        footer={
          <div style={{ display: 'flex', justifyContent: 'flex-end' }}>
            <Space>
              <Button theme='solid' size={'large'} onClick={submit}>
                {t('提交')}
              </Button>
              <Button
                theme='solid'
                size={'large'}
                type={'tertiary'}
                onClick={handleCancel}
              >
                {t('取消')}
              </Button>
            </Space>
          </div>
        }
        closeIcon={null}
        onCancel={() => handleCancel()}
        width={isMobile() ? '100%' : 600}
      >
        <Spin spinning={loading}>
          <Input
            style={{ marginTop: 20 }}
            label={i18n.t('用户名')}
            name='username'
            addonBefore={t('用户名')}
            placeholder={t('请输入用户名')}
            onChange={(value) => handleInputChange('username', value)}
            value={username}
            autoComplete='off'
          />
          <Input
            style={{ marginTop: 20 }}
            addonBefore={t('显示名')}
            label={i18n.t('显示名')}
            name='display_name'
            autoComplete='off'
            placeholder={t('请输入显示名称')}
            onChange={(value) => handleInputChange('display_name', value)}
            value={display_name}
          />
          <Input
            style={{ marginTop: 20 }}
            label={i18n.t('密码')}
            name='password'
            type={'password'}
            addonBefore={t('密码')}
            placeholder={t('请输入密码')}
            onChange={(value) => handleInputChange('password', value)}
            value={password}
            autoComplete='off'
          />
        </Spin>
      </SideSheet>
    </>
  );
};

export default AddUser;
