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
import React, { useEffect, useState } from 'react';
import { Card, Spin, Tabs } from '@douyinfe/semi-ui';
import SettingsGeneral from '../pages/Setting/Operation/SettingsGeneral.js';
import SettingsDrawing from '../pages/Setting/Operation/SettingsDrawing.js';
import SettingsSensitiveWords from '../pages/Setting/Operation/SettingsSensitiveWords.js';
import SettingsLog from '../pages/Setting/Operation/SettingsLog.js';
import SettingsDataDashboard from '../pages/Setting/Operation/SettingsDataDashboard.js';
import SettingsMonitoring from '../pages/Setting/Operation/SettingsMonitoring.js';
import SettingsCheckIn from '../pages/Setting/Operation/SettingsCheckIn.js';
import SettingsCreditLimit from '../pages/Setting/Operation/SettingsCreditLimit.js';
import SettingsRebate from '../pages/Setting/Operation/SettingsRebate.js';
import ModelSettingsVisualEditor from '../pages/Setting/Operation/ModelSettingsVisualEditor.js';
import GroupRatioSettings from '../pages/Setting/Operation/GroupRatioSettings.js';
import ModelCommonRatioSettings from '../pages/Setting/Operation/ModelCommonRatioSettings.js';
import ModelRatioSettings from '../pages/Setting/Operation/ModelRatioSettings.js';

import { API, showError, showSuccess } from '../helpers';
import SettingsChats from '../pages/Setting/Operation/SettingsChats.js';
import { useTranslation } from 'react-i18next';
import ModelRatioNotSetEditor from '../pages/Setting/Operation/ModelRationNotSetEditor.js';

import i18n from '../i18n/i18n';
const OperationSetting = () => {
  const { t } = useTranslation();
  let [inputs, setInputs] = useState({
    QuotaForNewUser: 0,
    QuotaForInviter: 0,
    QuotaForInvitee: 0,
    QuotaRemindThreshold: 0,
    PreConsumedQuota: 0,
    StreamCacheQueueLength: 0,
    ModelRatio: '',
    CacheRatio: '',
    CompletionRatio: '',
    ModelPrice: '',
    GroupRatio: '',
    UserUsableGroups: '',
    TopUpLink: '',
    'general_setting.docs_link': '',
    // ChatLink2: '', // Biến trạng thái mới được thêm
    QuotaPerUnit: 0,
    AutomaticDisableChannelEnabled: false,
    AutomaticEnableChannelEnabled: false,
    ChannelDisableThreshold: 0,
    LogConsumeEnabled: false,
    LogChatContentEnabled: false,
    LogErrorEnabled: false,
    DisplayInCurrencyEnabled: false,
    DisplayTokenStatEnabled: false,
    CheckSensitiveEnabled: false,
    CheckSensitiveOnPromptEnabled: false,
    CheckSensitiveOnCompletionEnabled: '',
    StopOnSensitiveEnabled: '',
    SensitiveWords: '',
    MjNotifyEnabled: false,
    MjAccountFilterEnabled: false,
    MjModeClearEnabled: false,
    MjForwardUrlEnabled: false,
    MjActionCheckSuccessEnabled: false,
    DrawingEnabled: false,
    DataExportEnabled: false,
    DataExportDefaultTime: 'hour',
    DataExportInterval: 5,
    DefaultCollapseSidebar: false, // Mặc định thu gọn thanh bên
    RetryTimes: 0,
    Chats: '[]',
    DemoSiteEnabled: false,
    SelfUseModeEnabled: false,
    AutomaticDisableKeywords: '',
    CheckInEnabled: false,
    CheckInQuota: '',
    CheckInMaxQuota: '',
    RebateEnabled: false,
    RebatePercentage: 0,
    AffEnabled: false,
  });

  let [loading, setLoading] = useState(false);

  const getOptions = async () => {
    const res = await API.get('/api/option/');
    const { success, message, data } = res.data;
    if (success) {
      let newInputs = {};
      data.forEach((item) => {
        if (
          item.key === 'ModelRatio' ||
          item.key === 'GroupRatio' ||
          item.key === 'UserUsableGroups' ||
          item.key === 'CompletionRatio' ||
          item.key === 'ModelPrice' ||
          item.key === 'CacheRatio'
        ) {
          item.value = JSON.stringify(JSON.parse(item.value), null, 2);
        }
        if (
          item.key.endsWith('Enabled') ||
          ['DefaultCollapseSidebar'].includes(item.key)
        ) {
          newInputs[item.key] = item.value === 'true' ? true : false;
        } else {
          newInputs[item.key] = item.value;
        }
      });

      setInputs(newInputs);
    } else {
      showError(message);
    }
  };
  async function onRefresh() {
    try {
      setLoading(true);
      await getOptions();
      // showSuccess('Làm mới thành công');
    } catch (error) {
      showError(i18n.t('Refresh failed'));
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    onRefresh();
  }, []);

  return (
    <>
      <Spin spinning={loading} size='large'>
        {/* Cài đặt chung */}
        <Card style={{ marginTop: '10px' }}>
          <SettingsGeneral options={inputs} refresh={onRefresh} />
        </Card>
        {/* Cài đặt vẽ tranh */}
        <Card style={{ marginTop: '10px' }}>
          <SettingsDrawing options={inputs} refresh={onRefresh} />
        </Card>
        {/* Cài đặt lọc từ bị chặn */}
        <Card style={{ marginTop: '10px' }}>
          <SettingsSensitiveWords options={inputs} refresh={onRefresh} />
        </Card>
        {/* Cài đặt log */}
        <Card style={{ marginTop: '10px' }}>
          <SettingsLog options={inputs} refresh={onRefresh} />
        </Card>
        {/* Bảng dữ liệu */}
        <Card style={{ marginTop: '10px' }}>
          <SettingsDataDashboard options={inputs} refresh={onRefresh} />
        </Card>
        {/* Cài đặt giám sát */}
        <Card style={{ marginTop: '10px' }}>
          <SettingsMonitoring options={inputs} refresh={onRefresh} />
        </Card>
        {/* Cài đặt điểm danh */}
        <Card style={{ marginTop: '10px' }}>
          <SettingsCheckIn options={inputs} refresh={onRefresh} />
        </Card>
        {/* Cài đặt hạn mức */}
        <Card style={{ marginTop: '10px' }}>
          <SettingsCreditLimit options={inputs} refresh={onRefresh} />
        </Card>
        {/* Cài đặt hoàn tiền */}
        <Card style={{ marginTop: '10px' }}>
          <SettingsRebate options={inputs} refresh={onRefresh} />
        </Card>
        {/* Cài đặt trò chuyện */}
        <Card style={{ marginTop: '10px' }}>
          <SettingsChats options={inputs} refresh={onRefresh} />
        </Card>
        {/* Cài đặt hệ số nhóm */}
        <Card style={{ marginTop: '10px' }}>
          <GroupRatioSettings options={inputs} refresh={onRefresh} />
        </Card>
        {/* Cài đặt liên quan đến hệ số mô hình */}
        <Card style={{ marginTop: '10px' }}>
          <ModelCommonRatioSettings options={inputs} refresh={onRefresh} />
        </Card>
        {/* Gộp cài đặt hệ số mô hình và cài đặt hệ số trực quan */}
        <Card style={{ marginTop: '10px' }}>
          <Tabs type='line'>
            <Tabs.TabPane tab={t('模型倍率设置')} itemKey='model'>
              <ModelRatioSettings options={inputs} refresh={onRefresh} />
            </Tabs.TabPane>
            <Tabs.TabPane tab={t('可视化倍率设置')} itemKey='visual'>
              <ModelSettingsVisualEditor options={inputs} refresh={onRefresh} />
            </Tabs.TabPane>
            <Tabs.TabPane tab={t('未设置倍率模型')} itemKey='unset_models'>
              <ModelRatioNotSetEditor options={inputs} refresh={onRefresh} />
            </Tabs.TabPane>
          </Tabs>
        </Card>
      </Spin>
    </>
  );
};

export default OperationSetting;
