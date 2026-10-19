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
import i18n from '../i18n/i18n';
import React, { useEffect, useState, useRef } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Button,
  Form,
  Row,
  Col,
  Typography,
  Modal,
  Banner,
  TagInput,
  Spin,
  Card,
  Select,
} from '@douyinfe/semi-ui';
const { Text } = Typography;
import {
  removeTrailingSlash,
  showError,
  showSuccess,
  verifyJSON,
} from '../helpers/utils';
import { API } from '../helpers/api';

const SystemSetting = () => {
  let [inputs, setInputs] = useState({
    PasswordLoginEnabled: '',
    PasswordRegisterEnabled: '',
    EmailVerificationEnabled: '',
    GitHubOAuthEnabled: '',
    GitHubClientId: '',
    GitHubClientSecret: '',
    'oidc.enabled': '',
    'oidc.client_id': '',
    'oidc.client_secret': '',
    'oidc.well_known': '',
    'oidc.authorization_endpoint': '',
    'oidc.token_endpoint': '',
    'oidc.user_info_endpoint': '',
    Notice: '',
    SMTPServer: '',
    SMTPPort: '',
    SMTPAccount: '',
    SMTPFrom: '',
    SMTPToken: '',
    ServerAddress: '',
    WorkerUrl: '',
    WorkerValidKey: '',
    EpayId: '',
    EpayKey: '',
    Price: 7.3,
    MinTopUp: 1,
    TopupGroupRatio: '',
    PayAddress: '',
    CustomCallbackAddress: '',
    Footer: '',
    WeChatAuthEnabled: '',
    WeChatServerAddress: '',
    WeChatServerToken: '',
    WeChatAccountQRCodeImageURL: '',
    TurnstileCheckEnabled: '',
    TurnstileSiteKey: '',
    TurnstileSecretKey: '',
    RegisterEnabled: '',
    EmailDomainRestrictionEnabled: '',
    EmailAliasRestrictionEnabled: '',
    SMTPSSLEnabled: '',
    EmailDomainWhitelist: [],
    // telegram login
    TelegramOAuthEnabled: '',
    TelegramBotToken: '',
    TelegramBotName: '',
    LinuxDOOAuthEnabled: '',
    LinuxDOClientId: '',
    LinuxDOClientSecret: '',
    LinuxDOMinimumTrustLevel: '',
    IDCFlareOAuthEnabled: '',
    IDCFlareClientId: '',
    IDCFlareClientSecret: '',
    IDCFlareMinimumTrustLevel: '',
    // reverse proxy settings
    ReverseProxyEnabled: '',
    ReverseProxyProvider: '',
  });

  const [originInputs, setOriginInputs] = useState({});
  const [loading, setLoading] = useState(false);
  const [isLoaded, setIsLoaded] = useState(false);
  const formApiRef = useRef(null);
  const [emailDomainWhitelist, setEmailDomainWhitelist] = useState([]);
  const [showPasswordLoginConfirmModal, setShowPasswordLoginConfirmModal] =
    useState(false);
  const [linuxDOOAuthEnabled, setLinuxDOOAuthEnabled] = useState(false);
  const [idcFlareOAuthEnabled, setIDCFlareOAuthEnabled] = useState(false);
  const [emailToAdd, setEmailToAdd] = useState('');

  const getOptions = async () => {
    setLoading(true);
    const res = await API.get('/api/option/');
    const { success, message, data } = res.data;
    if (success) {
      let newInputs = {};
      data.forEach((item) => {
        switch (item.key) {
          case 'TopupGroupRatio':
            item.value = JSON.stringify(JSON.parse(item.value), null, 2);
            break;
          case 'EmailDomainWhitelist':
            setEmailDomainWhitelist(item.value ? item.value.split(',') : []);
            break;
          case 'PasswordLoginEnabled':
          case 'PasswordRegisterEnabled':
          case 'EmailVerificationEnabled':
          case 'GitHubOAuthEnabled':
          case 'WeChatAuthEnabled':
          case 'TelegramOAuthEnabled':
          case 'RegisterEnabled':
          case 'TurnstileCheckEnabled':
          case 'EmailDomainRestrictionEnabled':
          case 'EmailAliasRestrictionEnabled':
          case 'SMTPSSLEnabled':
          case 'LinuxDOOAuthEnabled':
          case 'IDCFlareOAuthEnabled':
          case 'ReverseProxyEnabled':
          case 'oidc.enabled':
            item.value = item.value === 'true';
            break;
          case 'Price':
          case 'MinTopUp':
            item.value = parseFloat(item.value);
            break;
          default:
            break;
        }
        newInputs[item.key] = item.value;
      });
      setInputs(newInputs);
      setOriginInputs(newInputs);
      if (formApiRef.current) {
        formApiRef.current.setValues(newInputs);
      }
      setIsLoaded(true);
    } else {
      showError(message);
    }
    setLoading(false);
  };

  useEffect(() => {
    getOptions();
  }, []);

  const updateOptions = async (options) => {
    setLoading(true);
    try {
      // Tách các tùy chọn kiểu checkbox khỏi các tùy chọn khác
      const checkboxOptions = options.filter((opt) =>
        opt.key.toLowerCase().endsWith('enabled'),
      );
      const otherOptions = options.filter(
        (opt) => !opt.key.toLowerCase().endsWith('enabled'),
      );

      // Xử lý các tùy chọn kiểu checkbox
      for (const opt of checkboxOptions) {
        const res = await API.put('/api/option/', {
          key: opt.key,
          value: opt.value.toString(),
        });
        if (!res.data.success) {
          showError(res.data.message);
          return;
        }
      }

      // Xử lý các tùy chọn khác
      if (otherOptions.length > 0) {
        const requestQueue = otherOptions.map((opt) =>
          API.put('/api/option/', {
            key: opt.key,
            value:
              typeof opt.value === 'boolean' ? opt.value.toString() : opt.value,
          }),
        );

        const results = await Promise.all(requestQueue);

        // Kiểm tra tất cả yêu cầu có thành công không
        const errorResults = results.filter((res) => !res.data.success);
        errorResults.forEach((res) => {
          showError(res.data.message);
        });
      }

      showSuccess(i18n.t('Updated successfully'));
      // Cập nhật trạng thái cục bộ
      const newInputs = { ...inputs };
      options.forEach((opt) => {
        newInputs[opt.key] = opt.value;
      });
      setInputs(newInputs);
    } catch (error) {
      showError(i18n.t('Update failed'));
    }
    setLoading(false);
  };

  const handleFormChange = (values) => {
    setInputs(values);
  };

  const submitServerAddress = async () => {
    let ServerAddress = removeTrailingSlash(inputs.ServerAddress);
    await updateOptions([{ key: 'ServerAddress', value: ServerAddress }]);
  };

  const submitWorker = async () => {
    let WorkerUrl = removeTrailingSlash(inputs.WorkerUrl);
    await updateOptions([
      { key: 'WorkerUrl', value: WorkerUrl },
      { key: 'WorkerValidKey', value: inputs.WorkerValidKey },
    ]);
  };

  const submitPayAddress = async () => {
    if (inputs.ServerAddress === '') {
      showError('Please fill in the server address first');
      return;
    }
    if (originInputs['TopupGroupRatio'] !== inputs.TopupGroupRatio) {
      if (!verifyJSON(inputs.TopupGroupRatio)) {
        showError(i18n.t('Top-up group ratio is not a valid JSON string'));
        return;
      }
    }

    const options = [
      { key: 'PayAddress', value: removeTrailingSlash(inputs.PayAddress) },
    ];

    if (inputs.EpayId !== '') {
      options.push({ key: 'EpayId', value: inputs.EpayId });
    }
    if (inputs.EpayKey !== undefined && inputs.EpayKey !== '') {
      options.push({ key: 'EpayKey', value: inputs.EpayKey });
    }
    if (inputs.Price !== '') {
      options.push({ key: 'Price', value: inputs.Price.toString() });
    }
    if (inputs.MinTopUp !== '') {
      options.push({ key: 'MinTopUp', value: inputs.MinTopUp.toString() });
    }
    if (inputs.CustomCallbackAddress !== '') {
      options.push({
        key: 'CustomCallbackAddress',
        value: inputs.CustomCallbackAddress,
      });
    }
    if (originInputs['TopupGroupRatio'] !== inputs.TopupGroupRatio) {
      options.push({ key: 'TopupGroupRatio', value: inputs.TopupGroupRatio });
    }

    await updateOptions(options);
  };

  const submitSMTP = async () => {
    const options = [];

    if (originInputs['SMTPServer'] !== inputs.SMTPServer) {
      options.push({ key: 'SMTPServer', value: inputs.SMTPServer });
    }
    if (originInputs['SMTPAccount'] !== inputs.SMTPAccount) {
      options.push({ key: 'SMTPAccount', value: inputs.SMTPAccount });
    }
    if (originInputs['SMTPFrom'] !== inputs.SMTPFrom) {
      options.push({ key: 'SMTPFrom', value: inputs.SMTPFrom });
    }
    if (
      originInputs['SMTPPort'] !== inputs.SMTPPort &&
      inputs.SMTPPort !== ''
    ) {
      options.push({ key: 'SMTPPort', value: inputs.SMTPPort });
    }
    if (
      originInputs['SMTPToken'] !== inputs.SMTPToken &&
      inputs.SMTPToken !== ''
    ) {
      options.push({ key: 'SMTPToken', value: inputs.SMTPToken });
    }

    if (options.length > 0) {
      await updateOptions(options);
    }
  };

  const submitEmailDomainWhitelist = async () => {
    if (Array.isArray(emailDomainWhitelist)) {
      await updateOptions([
        {
          key: 'EmailDomainWhitelist',
          value: emailDomainWhitelist.join(','),
        },
      ]);
    } else {
      showError('Email domain whitelist format is incorrect');
    }
  };

  const handleAddEmail = () => {
    if (emailToAdd && emailToAdd.trim() !== '') {
      const domain = emailToAdd.trim();

      // Kiểm tra định dạng tên miền
      const domainRegex = /^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$/;
      if (!domainRegex.test(domain)) {
        showError('Email domain format is incorrect, please enter a valid domain such as gmail.com');
        return;
      }

      // Kiểm tra đã tồn tại chưa
      if (emailDomainWhitelist.includes(domain)) {
        showError(i18n.t('This domain is already in the whitelist'));
        return;
      }

      setEmailDomainWhitelist([...emailDomainWhitelist, domain]);
      setEmailToAdd('');
      showSuccess(i18n.t('Added to the whitelist'));
    }
  };

  const submitWeChat = async () => {
    const options = [];

    if (originInputs['WeChatServerAddress'] !== inputs.WeChatServerAddress) {
      options.push({
        key: 'WeChatServerAddress',
        value: removeTrailingSlash(inputs.WeChatServerAddress),
      });
    }
    if (
      originInputs['WeChatAccountQRCodeImageURL'] !==
      inputs.WeChatAccountQRCodeImageURL
    ) {
      options.push({
        key: 'WeChatAccountQRCodeImageURL',
        value: inputs.WeChatAccountQRCodeImageURL,
      });
    }
    if (
      originInputs['WeChatServerToken'] !== inputs.WeChatServerToken &&
      inputs.WeChatServerToken !== ''
    ) {
      options.push({
        key: 'WeChatServerToken',
        value: inputs.WeChatServerToken,
      });
    }

    if (options.length > 0) {
      await updateOptions(options);
    }
  };

  const submitGitHubOAuth = async () => {
    const options = [];

    if (originInputs['GitHubClientId'] !== inputs.GitHubClientId) {
      options.push({ key: 'GitHubClientId', value: inputs.GitHubClientId });
    }
    if (
      originInputs['GitHubClientSecret'] !== inputs.GitHubClientSecret &&
      inputs.GitHubClientSecret !== ''
    ) {
      options.push({
        key: 'GitHubClientSecret',
        value: inputs.GitHubClientSecret,
      });
    }

    if (options.length > 0) {
      await updateOptions(options);
    }
  };

  const submitOIDCSettings = async () => {
    if (inputs['oidc.well_known'] !== '') {
      if (
        !inputs['oidc.well_known'].startsWith('http://') &&
        !inputs['oidc.well_known'].startsWith('https://')
      ) {
        showError(t('Well-Known URL 必须以 http:// 或 https:// 开头'));
        return;
      }
      try {
        const res = await API.get(inputs['oidc.well_known']);
        inputs['oidc.authorization_endpoint'] =
          res.data['authorization_endpoint'];
        inputs['oidc.token_endpoint'] = res.data['token_endpoint'];
        inputs['oidc.user_info_endpoint'] = res.data['userinfo_endpoint'];
        showSuccess('OIDC config fetched successfully!');
      } catch (err) {
        console.error(err);
        showError(
          'Failed to fetch OIDC config, please check the network and Well-Known URL',
        );
        return;
      }
    }

    const options = [];

    if (originInputs['oidc.well_known'] !== inputs['oidc.well_known']) {
      options.push({
        key: 'oidc.well_known',
        value: inputs['oidc.well_known'],
      });
    }
    if (originInputs['oidc.client_id'] !== inputs['oidc.client_id']) {
      options.push({ key: 'oidc.client_id', value: inputs['oidc.client_id'] });
    }
    if (
      originInputs['oidc.client_secret'] !== inputs['oidc.client_secret'] &&
      inputs['oidc.client_secret'] !== ''
    ) {
      options.push({
        key: 'oidc.client_secret',
        value: inputs['oidc.client_secret'],
      });
    }
    if (
      originInputs['oidc.authorization_endpoint'] !==
      inputs['oidc.authorization_endpoint']
    ) {
      options.push({
        key: 'oidc.authorization_endpoint',
        value: inputs['oidc.authorization_endpoint'],
      });
    }
    if (originInputs['oidc.token_endpoint'] !== inputs['oidc.token_endpoint']) {
      options.push({
        key: 'oidc.token_endpoint',
        value: inputs['oidc.token_endpoint'],
      });
    }
    if (
      originInputs['oidc.user_info_endpoint'] !==
      inputs['oidc.user_info_endpoint']
    ) {
      options.push({
        key: 'oidc.user_info_endpoint',
        value: inputs['oidc.user_info_endpoint'],
      });
    }

    if (options.length > 0) {
      await updateOptions(options);
    }
  };

  const submitTelegramSettings = async () => {
    const options = [
      { key: 'TelegramBotToken', value: inputs.TelegramBotToken },
      { key: 'TelegramBotName', value: inputs.TelegramBotName },
    ];
    await updateOptions(options);
  };

  const submitTurnstile = async () => {
    const options = [];

    if (originInputs['TurnstileSiteKey'] !== inputs.TurnstileSiteKey) {
      options.push({ key: 'TurnstileSiteKey', value: inputs.TurnstileSiteKey });
    }
    if (
      originInputs['TurnstileSecretKey'] !== inputs.TurnstileSecretKey &&
      inputs.TurnstileSecretKey !== ''
    ) {
      options.push({
        key: 'TurnstileSecretKey',
        value: inputs.TurnstileSecretKey,
      });
    }

    if (options.length > 0) {
      await updateOptions(options);
    }
  };

  const submitLinuxDOOAuth = async () => {
    const options = [];

    if (originInputs['LinuxDOClientId'] !== inputs.LinuxDOClientId) {
      options.push({ key: 'LinuxDOClientId', value: inputs.LinuxDOClientId });
    }
    if (
      originInputs['LinuxDOClientSecret'] !== inputs.LinuxDOClientSecret &&
      inputs.LinuxDOClientSecret !== ''
    ) {
      options.push({
        key: 'LinuxDOClientSecret',
        value: inputs.LinuxDOClientSecret,
      });
    }
    if (originInputs['LinuxDOMinimumTrustLevel'] !== inputs.LinuxDOMinimumTrustLevel) {
      options.push({
        key: 'LinuxDOMinimumTrustLevel',
        value: inputs.LinuxDOMinimumTrustLevel.toString(),
      });
    }

    if (options.length > 0) {
      await updateOptions(options);
    }
  };

  const submitIDCFlareOAuth = async () => {
    const options = [];

    if (originInputs['IDCFlareClientId'] !== inputs.IDCFlareClientId) {
      options.push({ key: 'IDCFlareClientId', value: inputs.IDCFlareClientId });
    }
    if (
      originInputs['IDCFlareClientSecret'] !== inputs.IDCFlareClientSecret &&
      inputs.IDCFlareClientSecret !== ''
    ) {
      options.push({
        key: 'IDCFlareClientSecret',
        value: inputs.IDCFlareClientSecret,
      });
    }
    if (originInputs['IDCFlareMinimumTrustLevel'] !== inputs.IDCFlareMinimumTrustLevel) {
      options.push({
        key: 'IDCFlareMinimumTrustLevel',
        value: inputs.IDCFlareMinimumTrustLevel.toString(),
      });
    }

    if (options.length > 0) {
      await updateOptions(options);
    }
  };

  const submitReverseProxy = async () => {
    const options = [];

    if (originInputs['ReverseProxyEnabled'] !== inputs.ReverseProxyEnabled) {
      options.push({ key: 'ReverseProxyEnabled', value: inputs.ReverseProxyEnabled });
    }
    if (originInputs['ReverseProxyProvider'] !== inputs.ReverseProxyProvider) {
      options.push({ key: 'ReverseProxyProvider', value: inputs.ReverseProxyProvider });
    }

    if (options.length > 0) {
      await updateOptions(options);
    }
  };

  const handleCheckboxChange = async (optionKey, event) => {
    const value = event.target.checked;

    if (optionKey === 'PasswordLoginEnabled' && !value) {
      setShowPasswordLoginConfirmModal(true);
    } else {
      await updateOptions([{ key: optionKey, value }]);
    }
    if (optionKey === 'LinuxDOOAuthEnabled') {
      setLinuxDOOAuthEnabled(value);
    }
    if (optionKey === 'IDCFlareOAuthEnabled') {
      setIDCFlareOAuthEnabled(value);
    }
  };

  const handlePasswordLoginConfirm = async () => {
    await updateOptions([{ key: 'PasswordLoginEnabled', value: false }]);
    setShowPasswordLoginConfirmModal(false);
  };

  return (
    <div>
      {isLoaded ? (
        <Form
          initValues={inputs}
          onValueChange={handleFormChange}
          getFormApi={(api) => (formApiRef.current = api)}
        >
          {({ formState, values, formApi }) => (
            <div
              style={{
                display: 'flex',
                flexDirection: 'column',
                gap: '10px',
                marginTop: '10px',
              }}
            >
              <Card>
                <Form.Section text={i18n.t('General Settings')}>
                  <Form.Input
                    field='ServerAddress'
                    label={i18n.t('服务器地址')}
                    placeholder={t('例如：https://yourdomain.com')}
                    style={{ width: '100%' }}
                  />
                  <Button onClick={submitServerAddress}>{t('更新服务器地址')}</Button>
                </Form.Section>
              </Card>
              
              <Card>
                <Form.Section text={i18n.t('Reverse Proxy Settings')}>
                  <Text>{t('用以支持系统在反向代理后运行时正确识别客户端IP地址')}</Text>
                  <Form.Checkbox
                    field='ReverseProxyEnabled'
                    noLabel
                    onChange={(e) =>
                      handleCheckboxChange('ReverseProxyEnabled', e)
                    }
                  >
                    {t('系统在反向代理后运行')}
                  </Form.Checkbox>
                  {inputs.ReverseProxyEnabled && (
                    <Row
                      gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}
                      style={{ marginTop: 16 }}
                    >
                      <Col xs={24} sm={24} md={12} lg={12} xl={12}>
                        <Form.Select
                          field='ReverseProxyProvider'
                          label={i18n.t('Reverse proxy provider')}
                          placeholder={i18n.t('Select reverse proxy provider')}
                          style={{ width: '100%' }}
                        >
                          <Select.Option value='nginx'>{t('Nginx / OpenResty (通用)')}</Select.Option>
                          <Select.Option value='cloudflare'>{i18n.t('Cloudflare')}</Select.Option>
                        </Form.Select>
                      </Col>
                    </Row>
                  )}
                  <Button onClick={submitReverseProxy} style={{ marginTop: 16 }}>
                    {t('保存反向代理设置')}
                  </Button>
                </Form.Section>
              </Card>
              <Card>
                <Form.Section text={i18n.t('Proxy Settings')}>
                  <Text>
                    {t('（支持')}{' '}
                    <a
                      href='https://github.com/Calcium-Ion/new-api-worker'
                      target='_blank'
                      rel='noreferrer'
                    >
                      new-api-worker
                    </a>
                    {t('）')}
                  </Text>
                  <Row
                    gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}
                  >
                    <Col xs={24} sm={24} md={12} lg={12} xl={12}>
                      <Form.Input
                        field='WorkerUrl'
                        label={i18n.t('Worker Address')}
                        placeholder={t('例如：https://workername.yourdomain.workers.dev')}
                      />
                    </Col>
                    <Col xs={24} sm={24} md={12} lg={12} xl={12}>
                      <Form.Input
                        field='WorkerValidKey'
                        label={i18n.t('Worker Key')}
                        placeholder={i18n.t('Sensitive info is not sent to the frontend')}
                        type='password'
                      />
                    </Col>
                  </Row>
                  <Button onClick={submitWorker}>{t('更新Worker设置')}</Button>
                </Form.Section>
              </Card>

              <Card>
                <Form.Section text={i18n.t('Payment Settings')}>
                  <Text>
                    {t('（当前仅支持易支付接口，默认使用上方服务器地址作为回调地址！）')}
                  </Text>
                  <Row
                    gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}
                  >
                    <Col xs={24} sm={24} md={8} lg={8} xl={8}>
                      <Form.Input
                        field='PayAddress'
                        label={i18n.t('Payment address')}
                        placeholder={t('例如：https://yourdomain.com')}
                      />
                    </Col>
                    <Col xs={24} sm={24} md={8} lg={8} xl={8}>
                      <Form.Input
                        field='EpayId'
                        label={i18n.t('Epay merchant ID')}
                        placeholder={i18n.t('e.g. 0001')}
                      />
                    </Col>
                    <Col xs={24} sm={24} md={8} lg={8} xl={8}>
                      <Form.Input
                        field='EpayKey'
                        label={i18n.t('Epay merchant key')}
                        placeholder={i18n.t('Sensitive info is not sent to the frontend')}
                        type='password'
                      />
                    </Col>
                  </Row>
                  <Row
                    gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}
                    style={{ marginTop: 16 }}
                  >
                    <Col xs={24} sm={24} md={8} lg={8} xl={8}>
                      <Form.Input
                        field='CustomCallbackAddress'
                        label={i18n.t('Callback URL')}
                    placeholder={t('例如：https://yourdomain.com')}
                      />
                    </Col>
                    <Col xs={24} sm={24} md={8} lg={8} xl={8}>
                      <Form.InputNumber
                        field='Price'
                        precision={2}
                        label={i18n.t('Top-up price (x CNY per USD)')}
                        placeholder={i18n.t('e.g. 7 means 7 CNY per USD')}
                      />
                    </Col>
                    <Col xs={24} sm={24} md={8} lg={8} xl={8}>
                      <Form.InputNumber
                        field='MinTopUp'
                        label={i18n.t('Minimum top-up amount (USD)')}
                        placeholder={i18n.t('e.g. 2 means minimum $2')}
                      />
                    </Col>
                  </Row>
                  <Form.TextArea
                    field='TopupGroupRatio'
                    label={i18n.t('Top-up group ratio')}
                    placeholder={i18n.t('JSON text where keys are group names and values are ratios')}
                    autosize
                  />
                  <Button onClick={submitPayAddress}>{t('更新支付设置')}</Button>
                </Form.Section>
              </Card>

              <Card>
                <Form.Section text={i18n.t('Login & Registration Settings')}>
                  <Row
                    gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}
                  >
                    <Col xs={24} sm={24} md={12} lg={12} xl={12}>
                      <Form.Checkbox
                        field='PasswordLoginEnabled'
                        noLabel
                        onChange={(e) =>
                          handleCheckboxChange('PasswordLoginEnabled', e)
                        }
                      >
                        {t('允许通过密码进行登录')}
                      </Form.Checkbox>
                      <Form.Checkbox
                        field='PasswordRegisterEnabled'
                        noLabel
                        onChange={(e) =>
                          handleCheckboxChange('PasswordRegisterEnabled', e)
                        }
                      >
                        {t('允许通过密码进行注册')}
                      </Form.Checkbox>
                      <Form.Checkbox
                        field='EmailVerificationEnabled'
                        noLabel
                        onChange={(e) =>
                          handleCheckboxChange('EmailVerificationEnabled', e)
                        }
                      >
                        {t('通过密码注册时需要进行邮箱验证')}
                      </Form.Checkbox>
                      <Form.Checkbox
                        field='RegisterEnabled'
                        noLabel
                        onChange={(e) =>
                          handleCheckboxChange('RegisterEnabled', e)
                        }
                      >
                        {t('允许新用户注册')}
                      </Form.Checkbox>
                      <Form.Checkbox
                        field='TurnstileCheckEnabled'
                        noLabel
                        onChange={(e) =>
                          handleCheckboxChange('TurnstileCheckEnabled', e)
                        }
                      >
                        {t('启用 Turnstile 用户校验')}
                      </Form.Checkbox>
                    </Col>
                    <Col xs={24} sm={24} md={12} lg={12} xl={12}>
                      <Form.Checkbox
                        field='GitHubOAuthEnabled'
                        noLabel
                        onChange={(e) =>
                          handleCheckboxChange('GitHubOAuthEnabled', e)
                        }
                      >
                        {t('允许通过 GitHub 账户登录 & 注册')}
                      </Form.Checkbox>
                      <Form.Checkbox
                        field='LinuxDOOAuthEnabled'
                        noLabel
                        onChange={(e) =>
                          handleCheckboxChange('LinuxDOOAuthEnabled', e)
                        }
                      >
                        {t('允许通过 Linux DO 账户登录 & 注册')}
                      </Form.Checkbox>
                      <Form.Checkbox
                        field='IDCFlareOAuthEnabled'
                        noLabel
                        onChange={(e) =>
                          handleCheckboxChange('IDCFlareOAuthEnabled', e)
                        }
                      >
                        {t('允许通过 IDC Flare 账户登录 & 注册')}
                      </Form.Checkbox>
                      <Form.Checkbox
                        field='WeChatAuthEnabled'
                        noLabel
                        onChange={(e) =>
                          handleCheckboxChange('WeChatAuthEnabled', e)
                        }
                      >
                        {t('允许通过微信登录 & 注册')}
                      </Form.Checkbox>
                      <Form.Checkbox
                        field='TelegramOAuthEnabled'
                        noLabel
                        onChange={(e) =>
                          handleCheckboxChange('TelegramOAuthEnabled', e)
                        }
                      >
                        {t('允许通过 Telegram 进行登录')}
                      </Form.Checkbox>
                      <Form.Checkbox
                        field="['oidc.enabled']"
                        noLabel
                        onChange={(e) =>
                          handleCheckboxChange('oidc.enabled', e)
                        }
                      >
                        {t('允许通过 OIDC 进行登录')}
                      </Form.Checkbox>
                    </Col>
                  </Row>
                </Form.Section>
              </Card>

              <Card>
                <Form.Section text='Email Domain Whitelist'>
                  <Text>{t('用以防止恶意用户利用临时邮箱批量注册')}</Text>
                  <Row
                    gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}
                  >
                    <Col xs={24} sm={24} md={12} lg={12} xl={12}>
                      <Form.Checkbox
                        field='EmailDomainRestrictionEnabled'
                        noLabel
                        onChange={(e) =>
                          handleCheckboxChange(
                            'EmailDomainRestrictionEnabled',
                            e,
                          )
                        }
                      >
                        {t('启用邮箱域名白名单')}
                      </Form.Checkbox>
                    </Col>
                    <Col xs={24} sm={24} md={12} lg={12} xl={12}>
                      <Form.Checkbox
                        field='EmailAliasRestrictionEnabled'
                        noLabel
                        onChange={(e) =>
                          handleCheckboxChange(
                            'EmailAliasRestrictionEnabled',
                            e,
                          )
                        }
                      >
                        {t('启用邮箱别名限制')}
                      </Form.Checkbox>
                    </Col>
                  </Row>
                  <TagInput
                    value={emailDomainWhitelist}
                    onChange={setEmailDomainWhitelist}
                    placeholder={i18n.t('Enter domain and press Enter')}
                    style={{ width: '100%', marginTop: 16 }}
                  />
                  <Form.Input
                    placeholder={i18n.t('Enter email domain to add')}
                    value={emailToAdd}
                    onChange={(value) => setEmailToAdd(value)}
                    style={{ marginTop: 16 }}
                    suffix={
                      <Button theme="solid" type="primary" onClick={handleAddEmail}>{t('添加')}</Button>
                    }
                    onEnterPress={handleAddEmail}
                  />
                  <Button
                    onClick={submitEmailDomainWhitelist}
                    style={{ marginTop: 10 }}
                  >
                    {t('保存邮箱域名白名单设置')}
                  </Button>
                </Form.Section>
              </Card>
              <Card>
                <Form.Section text={i18n.t('SMTP Settings')}>
                  <Text>{t('用以支持系统的邮件发送')}</Text>
                  <Row
                    gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}
                  >
                    <Col xs={24} sm={24} md={8} lg={8} xl={8}>
                      <Form.Input field='SMTPServer' label={i18n.t('SMTP server address')} />
                    </Col>
                    <Col xs={24} sm={24} md={8} lg={8} xl={8}>
                      <Form.Input field='SMTPPort' label={i18n.t('SMTP port')} />
                    </Col>
                    <Col xs={24} sm={24} md={8} lg={8} xl={8}>
                      <Form.Input field='SMTPAccount' label={i18n.t('SMTP account')} />
                    </Col>
                  </Row>
                  <Row
                    gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}
                    style={{ marginTop: 16 }}
                  >
                    <Col xs={24} sm={24} md={8} lg={8} xl={8}>
                      <Form.Input field='SMTPFrom' label={i18n.t('SMTP sender email')} />
                    </Col>
                    <Col xs={24} sm={24} md={8} lg={8} xl={8}>
                      <Form.Input
                        field='SMTPToken'
                        label={i18n.t('SMTP access credential')}
                        type='password'
                        placeholder={i18n.t('Sensitive info is not sent to the frontend')}
                      />
                    </Col>
                    <Col xs={24} sm={24} md={8} lg={8} xl={8}>
                      <Form.Checkbox
                        field='SMTPSSLEnabled'
                        noLabel
                        onChange={(e) =>
                          handleCheckboxChange('SMTPSSLEnabled', e)
                        }
                      >
                        {t('启用SMTP SSL')}
                      </Form.Checkbox>
                    </Col>
                  </Row>
                  <Button onClick={submitSMTP}>{t('保存 SMTP 设置')}</Button>
                </Form.Section>
              </Card>
              <Card>
                <Form.Section text={i18n.t('OIDC Settings')}>
                  <Text>
                    {t('用以支持通过 OIDC 登录，例如 Okta、Auth0 等兼容 OIDC 协议的 IdP')}
                  </Text>
                  <Banner
                    type='info'
                    description={`Homepage URL: ${inputs.ServerAddress ? inputs.ServerAddress : i18n.t('site address')} | Redirect URL: ${inputs.ServerAddress ? inputs.ServerAddress : i18n.t('site address')}/oauth/oidc`}
                    style={{ marginBottom: 20, marginTop: 16 }}
                  />
                  <Text>
                    {t('若你的 OIDC Provider 支持 Discovery Endpoint，你可以仅填写 OIDC Well-Known URL，系统会自动获取 OIDC 配置')}
                  </Text>
                  <Row
                    gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}
                  >
                    <Col xs={24} sm={24} md={12} lg={12} xl={12}>
                      <Form.Input
                        field="['oidc.well_known']"
                        label={i18n.t('Well-Known URL')}
                        placeholder={i18n.t('Enter OIDC Well-Known URL')}
                      />
                    </Col>
                    <Col xs={24} sm={24} md={12} lg={12} xl={12}>
                      <Form.Input
                        field="['oidc.client_id']"
                        label={i18n.t('Client ID')}
                        placeholder={i18n.t('Enter OIDC Client ID')}
                      />
                    </Col>
                  </Row>
                  <Row
                    gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}
                  >
                    <Col xs={24} sm={24} md={12} lg={12} xl={12}>
                      <Form.Input
                        field="['oidc.client_secret']"
                        label={i18n.t('Client Secret')}
                        type='password'
                        placeholder={i18n.t('Sensitive info is not sent to the frontend')}
                      />
                    </Col>
                    <Col xs={24} sm={24} md={12} lg={12} xl={12}>
                      <Form.Input
                        field="['oidc.authorization_endpoint']"
                        label={i18n.t('Authorization Endpoint')}
                        placeholder={i18n.t('Enter OIDC Authorization Endpoint')}
                      />
                    </Col>
                  </Row>
                  <Row
                    gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}
                  >
                    <Col xs={24} sm={24} md={12} lg={12} xl={12}>
                      <Form.Input
                        field="['oidc.token_endpoint']"
                        label={i18n.t('Token Endpoint')}
                        placeholder={i18n.t('Enter OIDC Token Endpoint')}
                      />
                    </Col>
                    <Col xs={24} sm={24} md={12} lg={12} xl={12}>
                      <Form.Input
                        field="['oidc.user_info_endpoint']"
                        label={i18n.t('User Info Endpoint')}
                        placeholder={i18n.t('Enter OIDC Userinfo Endpoint')}
                      />
                    </Col>
                  </Row>
                  <Button onClick={submitOIDCSettings}>{t('保存 OIDC 设置')}</Button>
                </Form.Section>
              </Card>

              <Card>
                <Form.Section text={i18n.t('GitHub OAuth App Settings')}>
                  <Text>{t('用以支持通过 GitHub 进行登录注册')}</Text>
                  <Banner
                    type='info'
                    description={`Homepage URL: ${inputs.ServerAddress ? inputs.ServerAddress : i18n.t('site address')} | Authorization callback URL: ${inputs.ServerAddress ? inputs.ServerAddress : i18n.t('site address')}/oauth/github`}
                    style={{ marginBottom: 20, marginTop: 16 }}
                  />
                  <Row
                    gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}
                  >
                    <Col xs={24} sm={24} md={12} lg={12} xl={12}>
                      <Form.Input
                        field='GitHubClientId'
                        label={i18n.t('GitHub Client ID')}
                      />
                    </Col>
                    <Col xs={24} sm={24} md={12} lg={12} xl={12}>
                      <Form.Input
                        field='GitHubClientSecret'
                        label={i18n.t('GitHub Client Secret')}
                        type='password'
                        placeholder={i18n.t('Sensitive info is not sent to the frontend')}
                      />
                    </Col>
                  </Row>
                  <Button onClick={submitGitHubOAuth}>
                    {t('保存 GitHub OAuth 设置')}
                  </Button>
                </Form.Section>
              </Card>
              <Card>
                <Form.Section text={i18n.t('LinuxDO OAuth Settings')}>
                  <Text>
                    {t('用以支持通过 Linux DO 进行登录注册')}
                    <a
                      href='https://connect.linux.do/'
                      target='_blank'
                      rel='noreferrer'
                      style={{
                        display: 'inline-block',
                        marginLeft: 4,
                        marginRight: 4,
                      }}
                    >
                      {t('点击此处')}
                    </a>
                    {t('管理你的 LinuxDO OAuth App')}
                  </Text>
                  <Banner
                    type='info'
                    description={`Callback URL: ${inputs.ServerAddress ? inputs.ServerAddress : i18n.t('site address')}/oauth/linuxdo`}
                    style={{ marginBottom: 20, marginTop: 16 }}
                  />
                  <Row
                    gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}
                  >
                    <Col xs={24} sm={24} md={10} lg={10} xl={10}>
                      <Form.Input
                        field='LinuxDOClientId'
                        label={i18n.t('Linux DO Client ID')}
                        placeholder={i18n.t('Enter your LinuxDO OAuth App ID')}
                      />
                    </Col>
                    <Col xs={24} sm={24} md={10} lg={10} xl={10}>
                      <Form.Input
                        field='LinuxDOClientSecret'
                        label={i18n.t('Linux DO Client Secret')}
                        type='password'
                        placeholder={i18n.t('Sensitive info is not sent to the frontend')}
                      />
                    </Col>
                    <Col xs={24} sm={24} md={4} lg={4} xl={4}>
                      <Form.InputNumber
                        field='LinuxDOMinimumTrustLevel'
                        label={i18n.t('LinuxDO Minimum Trust Level')}
                        placeholder={i18n.t('Minimum trust level for registration')}
                        min={0}
                        max={4}
                      />
                    </Col>
                  </Row>
                  <Button onClick={submitLinuxDOOAuth}>
                    {t('保存 Linux DO OAuth 设置')}
                  </Button>
                </Form.Section>
              </Card>

              <Card>
                <Form.Section text={i18n.t('IDC Flare OAuth Settings')}>
                  <Text>
                    {t('用以支持通过 IDC Flare 进行登录注册')}
                    <a
                      href='https://connect.idcflare.com/'
                      target='_blank'
                      rel='noreferrer'
                      style={{
                        display: 'inline-block',
                        marginLeft: 4,
                        marginRight: 4,
                      }}
                    >
                      {t('点击此处')}
                    </a>
                    {t('管理你的 IDC Flare OAuth App')}
                  </Text>
                  <Banner
                    type='info'
                    description={`Callback URL: ${inputs.ServerAddress ? inputs.ServerAddress : i18n.t('site address')}/oauth/idcflare`}
                    style={{ marginBottom: 20, marginTop: 16 }}
                  />
                  <Row
                    gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}
                  >
                    <Col xs={24} sm={24} md={10} lg={10} xl={10}>
                      <Form.Input
                        field='IDCFlareClientId'
                        label={i18n.t('IDC Flare Client ID')}
                        placeholder={i18n.t('Enter your IDC Flare OAuth App ID')}
                      />
                    </Col>
                    <Col xs={24} sm={24} md={10} lg={10} xl={10}>
                      <Form.Input
                        field='IDCFlareClientSecret'
                        label={i18n.t('IDC Flare Client Secret')}
                        type='password'
                        placeholder={i18n.t('Sensitive info is not sent to the frontend')}
                      />
                    </Col>
                    <Col xs={24} sm={24} md={4} lg={4} xl={4}>
                      <Form.InputNumber
                        field='IDCFlareMinimumTrustLevel'
                        label={i18n.t('IDC Flare Minimum Trust Level')}
                        placeholder={i18n.t('Minimum trust level for registration')}
                        min={0}
                        max={4}
                      />
                    </Col>
                  </Row>
                  <Button onClick={submitIDCFlareOAuth}>
                    {t('保存 IDC Flare OAuth 设置')}
                  </Button>
                </Form.Section>
              </Card>

              <Card>
                <Form.Section text={i18n.t('WeChat Server Settings')}>
                  <Text>{t('用以支持通过微信进行登录注册')}</Text>
                  <Row
                    gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}
                  >
                    <Col xs={24} sm={24} md={8} lg={8} xl={8}>
                      <Form.Input
                        field='WeChatServerAddress'
                        label={i18n.t('WeChat Server address')}
                      />
                    </Col>
                    <Col xs={24} sm={24} md={8} lg={8} xl={8}>
                      <Form.Input
                        field='WeChatServerToken'
                        label={i18n.t('WeChat Server credential')}
                        type='password'
                        placeholder={i18n.t('Sensitive info is not sent to the frontend')}
                      />
                    </Col>
                    <Col xs={24} sm={24} md={8} lg={8} xl={8}>
                      <Form.Input
                        field='WeChatAccountQRCodeImageURL'
                        label={i18n.t('WeChat official account QR code image URL')}
                      />
                    </Col>
                  </Row>
                  <Button onClick={submitWeChat}>
                    {t('保存 WeChat Server 设置')}
                  </Button>
                </Form.Section>
              </Card>

              <Card>
                <Form.Section text={i18n.t('Telegram Login Settings')}>
                  <Text>{t('用以支持通过 Telegram 进行登录注册')}</Text>
                  <Row
                    gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}
                  >
                    <Col xs={24} sm={24} md={12} lg={12} xl={12}>
                      <Form.Input
                        field='TelegramBotToken'
                        label={i18n.t('Telegram Bot Token')}
                        placeholder={i18n.t('Sensitive info is not sent to the frontend')}
                        type='password'
                      />
                    </Col>
                    <Col xs={24} sm={24} md={12} lg={12} xl={12}>
                      <Form.Input
                        field='TelegramBotName'
                        label={i18n.t('Telegram Bot name')}
                      />
                    </Col>
                  </Row>
                  <Button onClick={submitTelegramSettings}>
                    {t('保存 Telegram 登录设置')}
                  </Button>
                </Form.Section>
              </Card>

              <Card>
                <Form.Section text={i18n.t('Turnstile Settings')}>
                  <Text>{t('用以支持用户校验')}</Text>
                  <Row
                    gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}
                  >
                    <Col xs={24} sm={24} md={12} lg={12} xl={12}>
                      <Form.Input
                        field='TurnstileSiteKey'
                        label={i18n.t('Turnstile Site Key')}
                      />
                    </Col>
                    <Col xs={24} sm={24} md={12} lg={12} xl={12}>
                      <Form.Input
                        field='TurnstileSecretKey'
                        label={i18n.t('Turnstile Secret Key')}
                        type='password'
                        placeholder={i18n.t('Sensitive info is not sent to the frontend')}
                      />
                    </Col>
                  </Row>
                  <Button onClick={submitTurnstile}>{t('保存 Turnstile 设置')}</Button>
                </Form.Section>
              </Card>

              <Modal
                title={i18n.t('Confirm disabling password login')}
                visible={showPasswordLoginConfirmModal}
                onOk={handlePasswordLoginConfirm}
                onCancel={() => {
                  setShowPasswordLoginConfirmModal(false);
                  formApiRef.current.setValue('PasswordLoginEnabled', true);
                }}
                okText={i18n.t('Confirm')}
                cancelText={i18n.t('取消')}
              >
                <p>{t('您确定要取消密码登录功能吗？这可能会影响用户的登录方式。')}</p>
              </Modal>
            </div>
          )}
        </Form>
      ) : (
        <div
          style={{
            display: 'flex',
            justifyContent: 'center',
            alignItems: 'center',
            height: '100vh',
          }}
        >
          <Spin size='large' />
        </div>
      )}
    </div>
  );
};

export default SystemSetting;
