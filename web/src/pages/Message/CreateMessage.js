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
import React, { useState, useEffect } from 'react';
import { API, isMobile, showError, showSuccess } from '../../helpers';
import Title from '@douyinfe/semi-ui/lib/es/typography/title';
import { 
  Button, 
  Input, 
  SideSheet, 
  Space, 
  Spin, 
  Select, 
  TextArea,
  RadioGroup,
  Radio,
  Form,
  Table,
  Checkbox,
  Pagination
} from '@douyinfe/semi-ui';
import { useTranslation } from 'react-i18next';

const CreateMessage = (props) => {
  const { t } = useTranslation();
  
  const originInputs = {
    title: '',
    content: '',
    format: 'markdown',
    user_ids: [],
  };
  
  const [inputs, setInputs] = useState(originInputs);
  const [loading, setLoading] = useState(false);
  const [users, setUsers] = useState([]);
  const [selectedUsers, setSelectedUsers] = useState([]);
  const [userLoading, setUserLoading] = useState(false);
  const [userPage, setUserPage] = useState(1);
  const [userPageSize] = useState(20);
  const [userTotal, setUserTotal] = useState(0);
  const [userSearchKeyword, setUserSearchKeyword] = useState('');
  const [selectAll, setSelectAll] = useState(false);
  
  const { title, content, format } = inputs;

  const handleInputChange = (name, value) => {
    setInputs((inputs) => ({ ...inputs, [name]: value }));
  };

  const loadUsers = async (page = 1, keyword = '') => {
    setUserLoading(true);
    try {
      let url = `/api/user/?p=${page}&page_size=${userPageSize}`;
      if (keyword) {
        url = `/api/user/search?keyword=${keyword}&p=${page}&page_size=${userPageSize}`;
      }
      
      const res = await API.get(url);
      const { success, message, data } = res.data;
      if (success) {
        setUsers(data.items || []);
        setUserTotal(data.total || 0);
        setUserPage(data.page || 1);
      } else {
        showError(message);
      }
    } catch (error) {
      showError(error.message);
    }
    setUserLoading(false);
  };

  useEffect(() => {
    if (props.visible) {
      loadUsers();
    }
  }, [props.visible]);

  const handleUserSelect = (userId, checked) => {
    if (checked) {
      setSelectedUsers([...selectedUsers, userId]);
    } else {
      setSelectedUsers(selectedUsers.filter(id => id !== userId));
      setSelectAll(false);
    }
  };

  const handleSelectAll = (checked) => {
    setSelectAll(checked);
    if (checked) {
      const allUserIds = users.map(user => user.id);
      setSelectedUsers([...new Set([...selectedUsers, ...allUserIds])]);
    } else {
      const currentPageUserIds = users.map(user => user.id);
      setSelectedUsers(selectedUsers.filter(id => !currentPageUserIds.includes(id)));
    }
  };

  const handleUserSearch = (keyword) => {
    setUserSearchKeyword(keyword);
    setUserPage(1);
    loadUsers(1, keyword);
  };

  const handleUserPageChange = (page) => {
    setUserPage(page);
    loadUsers(page, userSearchKeyword);
  };

  const userColumns = [
    {
      title: (
        <Checkbox
          checked={selectAll}
          onChange={handleSelectAll}
        />
      ),
      dataIndex: 'select',
      width: 50,
      render: (text, record) => (
        <Checkbox
          checked={selectedUsers.includes(record.id)}
          onChange={(checked) => handleUserSelect(record.id, checked)}
        />
      ),
    },
    {
      title: 'ID',
      dataIndex: 'id',
      width: 80,
    },
    {
      title: t('用户名'),
      dataIndex: 'username',
    },
    {
      title: t('显示名称'),
      dataIndex: 'display_name',
      render: (text) => text || '-',
    },
  ];

  const submit = async () => {
    setLoading(true);
    
    if (!inputs.title.trim()) {
      showError(t('请输入消息标题'));
      setLoading(false);
      return;
    }
    
    if (!inputs.content.trim()) {
      showError(t('请输入消息内容'));
      setLoading(false);
      return;
    }
    
    if (selectedUsers.length === 0) {
      showError(t('请选择至少一个收件人'));
      setLoading(false);
      return;
    }

    const messageData = {
      title: inputs.title.trim(),
      content: inputs.content.trim(),
      format: inputs.format,
      user_ids: selectedUsers,
    };

    try {
      const res = await API.post('/api/admin/messages', messageData);
      const { success, message } = res.data;
      if (success) {
        showSuccess(t('消息创建成功'));
        setInputs(originInputs);
        setSelectedUsers([]);
        setSelectAll(false);
        props.refresh();
        props.handleClose();
      } else {
        showError(message);
      }
    } catch (error) {
      showError(error.message);
    }
    setLoading(false);
  };

  const handleCancel = () => {
    setInputs(originInputs);
    setSelectedUsers([]);
    setSelectAll(false);
    props.handleClose();
  };

  return (
    <>
      <SideSheet
        placement={'right'}
        title={<Title level={3}>{t('创建消息')}</Title>}
        headerStyle={{ borderBottom: '1px solid var(--semi-color-border)' }}
        bodyStyle={{ borderBottom: '1px solid var(--semi-color-border)' }}
        visible={props.visible}
        footer={
          <div style={{ display: 'flex', justifyContent: 'flex-end' }}>
            <Space>
              <Button theme='solid' size={'large'} onClick={submit} loading={loading}>
                {t('发送消息')}
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
        width={isMobile() ? '100%' : 800}
      >
        <Spin spinning={loading}>
          <div style={{ padding: '20px 0' }}>
            <Input
              label={t('消息标题')}
              placeholder={t('请输入消息标题')}
              value={title}
              onChange={(value) => handleInputChange('title', value)}
              style={{ marginBottom: 20 }}
            />

            <div style={{ marginBottom: 20 }}>
              <label style={{ display: 'block', marginBottom: 8, fontWeight: 600 }}>
                {t('内容格式')}
              </label>
              <RadioGroup
                type="button"
                value={format}
                onChange={(e) => handleInputChange('format', e.target.value)}
              >
                <Radio value="markdown">{i18n.t('Markdown')}</Radio>
                <Radio value="html">HTML</Radio>
              </RadioGroup>
            </div>

            <div style={{ marginBottom: 20 }}>
              <label style={{ display: 'block', marginBottom: 8, fontWeight: 600 }}>
                {t('消息内容')}
              </label>
              <TextArea
                placeholder={format === 'html' ? t('请输入HTML格式的消息内容') : t('请输入Markdown格式的消息内容')}
                value={content}
                onChange={(value) => handleInputChange('content', value)}
                rows={8}
                style={{ fontFamily: 'monospace' }}
              />
            </div>

            <div>
              <label style={{ display: 'block', marginBottom: 8, fontWeight: 600 }}>
                {t('选择收件人')} ({selectedUsers.length} {t('人已选择')})
              </label>
              
              <div style={{ marginBottom: 12 }}>
                <Input
                  placeholder={t('搜索用户名或显示名称')}
                  prefix="🔍"
                  value={userSearchKeyword}
                  onChange={handleUserSearch}
                  style={{ width: 300 }}
                />
              </div>

              <Table
                columns={userColumns}
                dataSource={users}
                loading={userLoading}
                pagination={false}
                size="small"
                style={{ maxHeight: 300, overflow: 'auto' }}
              />
              
              <div style={{ marginTop: 12, textAlign: 'center' }}>
                <Pagination
                  total={userTotal}
                  currentPage={userPage}
                  pageSize={userPageSize}
                  onPageChange={handleUserPageChange}
                  showSizeChanger={false}
                  size="small"
                />
              </div>
            </div>
          </div>
        </Spin>
      </SideSheet>
    </>
  );
};

export default CreateMessage;