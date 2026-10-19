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
import React, { useContext, useEffect, useRef, useState } from 'react';
import {
  Banner,
  Button,
  Col,
  Form,
  Row,
  Modal,
  Space,
  Card,
  Switch,
} from '@douyinfe/semi-ui';
import { API, showError, showSuccess, timestamp2string } from '../helpers';
import { marked } from 'marked';
import { useTranslation } from 'react-i18next';
import { StatusContext } from '../context/Status/index.js';
import Text from '@douyinfe/semi-ui/lib/es/typography/text';

import i18n from '../i18n/i18n';
const OtherSetting = () => {
  const { t } = useTranslation();
  let [inputs, setInputs] = useState({
    Notice: '',
    SystemName: '',
    Logo: '',
    Footer: '',
    About: '',
    HomePageContent: '',
    custom_head_html: '',
    global_css: '',
    global_js: '',
    DisplayInCurrencyEnabled: false,
    DisplayTokenStatEnabled: false,
  });
  let [loading, setLoading] = useState(false);
  const [showUpdateModal, setShowUpdateModal] = useState(false);
  const [statusState, statusDispatch] = useContext(StatusContext);
  const [updateData, setUpdateData] = useState({
    tag_name: '',
    content: '',
  });

  const updateOption = async (key, value) => {
    setLoading(true);
    const res = await API.put('/api/option/', {
      key,
      value,
    });
    const { success, message } = res.data;
    if (success) {
      setInputs((inputs) => ({ ...inputs, [key]: value }));
    } else {
      showError(message);
    }
    setLoading(false);
  };

  const [loadingInput, setLoadingInput] = useState({
    Notice: false,
    SystemName: false,
    Logo: false,
    HomePageContent: false,
    About: false,
    Footer: false,
    CheckUpdate: false,
    custom_head_html: false,
    global_css: false,
    global_js: false,
  });
  const handleInputChange = async (value, e) => {
    const name = e.target.id;
    setInputs((inputs) => ({ ...inputs, [name]: value }));
  };

  // Cài đặt chung
  const formAPISettingGeneral = useRef();
  // Cài đặt chung - Notice
  const submitNotice = async () => {
    try {
      setLoadingInput((loadingInput) => ({ ...loadingInput, Notice: true }));
      await updateOption('Notice', inputs.Notice);
      showSuccess(t('公告已更新'));
    } catch (error) {
      console.error(t('公告更新失败'), error);
      showError(t('公告更新失败'));
    } finally {
      setLoadingInput((loadingInput) => ({ ...loadingInput, Notice: false }));
    }
  };
  // Cài đặt cá nhân hóa
  const formAPIPersonalization = useRef();
  //  Cài đặt cá nhân hóa - SystemName
  const submitSystemName = async () => {
    try {
      setLoadingInput((loadingInput) => ({
        ...loadingInput,
        SystemName: true,
      }));
      await updateOption('SystemName', inputs.SystemName);
      showSuccess(t('系统名称已更新'));
    } catch (error) {
      console.error(t('系统名称更新失败'), error);
      showError(t('系统名称更新失败'));
    } finally {
      setLoadingInput((loadingInput) => ({
        ...loadingInput,
        SystemName: false,
      }));
    }
  };

  // Cài đặt cá nhân hóa - Logo
  const submitLogo = async () => {
    try {
      setLoadingInput((loadingInput) => ({ ...loadingInput, Logo: true }));
      await updateOption('Logo', inputs.Logo);
      showSuccess(i18n.t('Logo updated'));
    } catch (error) {
      console.error(i18n.t('Failed to update logo'), error);
      showError(i18n.t('Failed to update logo'));
    } finally {
      setLoadingInput((loadingInput) => ({ ...loadingInput, Logo: false }));
    }
  };
  // Cài đặt cá nhân hóa - Nội dung trang chủ
  const submitOption = async (key) => {
    try {
      setLoadingInput((loadingInput) => ({
        ...loadingInput,
        HomePageContent: true,
      }));
      await updateOption(key, inputs[key]);
      showSuccess(i18n.t('Homepage content updated'));
    } catch (error) {
      console.error(i18n.t('Failed to update homepage content'), error);
      showError(i18n.t('Failed to update homepage content'));
    } finally {
      setLoadingInput((loadingInput) => ({
        ...loadingInput,
        HomePageContent: false,
      }));
    }
  };
  // Cài đặt cá nhân hóa - Giới thiệu
  const submitAbout = async () => {
    try {
      setLoadingInput((loadingInput) => ({ ...loadingInput, About: true }));
      await updateOption('About', inputs.About);
      showSuccess(i18n.t('About content updated'));
    } catch (error) {
      console.error(i18n.t('Failed to update About content'), error);
      showError(i18n.t('Failed to update About content'));
    } finally {
      setLoadingInput((loadingInput) => ({ ...loadingInput, About: false }));
    }
  };
  // Cài đặt cá nhân hóa - Footer
  const submitFooter = async () => {
    try {
      setLoadingInput((loadingInput) => ({ ...loadingInput, Footer: true }));
      await updateOption('Footer', inputs.Footer);
      showSuccess(i18n.t('Footer content updated'));
    } catch (error) {
      console.error(i18n.t('Failed to update footer content'), error);
      showError(i18n.t('Failed to update footer content'));
    } finally {
      setLoadingInput((loadingInput) => ({ ...loadingInput, Footer: false }));
    }
  };

  // Cài đặt cá nhân hóa - HTML header tùy chỉnh
  const submitCustomHeadHtml = async () => {
    try {
      setLoadingInput((loadingInput) => ({ ...loadingInput, custom_head_html: true }));
      await updateOption('custom_head_html', inputs.custom_head_html);
      showSuccess(i18n.t('Custom header HTML updated'));
    } catch (error) {
      console.error(i18n.t('Failed to update custom header HTML'), error);
      showError(i18n.t('Failed to update custom header HTML'));
    } finally {
      setLoadingInput((loadingInput) => ({ ...loadingInput, custom_head_html: false }));
    }
  };

  // Cài đặt cá nhân hóa - CSS toàn cục
  const submitGlobalCss = async () => {
    try {
      setLoadingInput((loadingInput) => ({ ...loadingInput, global_css: true }));
      await updateOption('global_css', inputs.global_css);
      showSuccess(i18n.t('Global CSS updated'));
    } catch (error) {
      console.error(i18n.t('Failed to update global CSS'), error);
      showError(i18n.t('Failed to update global CSS'));
    } finally {
      setLoadingInput((loadingInput) => ({ ...loadingInput, global_css: false }));
    }
  };

  // Cài đặt cá nhân hóa - Mã JavaScript toàn cục
  const submitGlobalJs = async () => {
    try {
      setLoadingInput((loadingInput) => ({ ...loadingInput, global_js: true }));
      await updateOption('global_js', inputs.global_js);
      showSuccess(i18n.t('Global JavaScript updated'));
    } catch (error) {
      console.error(i18n.t('Failed to update global JavaScript'), error);
      showError(i18n.t('Failed to update global JavaScript'));
    } finally {
      setLoadingInput((loadingInput) => ({ ...loadingInput, global_js: false }));
    }
  };

  const checkUpdate = async () => {
    try {
      setLoadingInput((loadingInput) => ({
        ...loadingInput,
        CheckUpdate: true,
      }));
      // Use a CORS proxy to avoid direct cross-origin requests to GitHub API
      // Option 1: Use a public CORS proxy service
      // const proxyUrl = 'https://cors-anywhere.herokuapp.com/';
      // const res = await API.get(
      //   `${proxyUrl}https://api.github.com/repos/Calcium-Ion/new-api/releases/latest`,
      // );

      // Option 2: Use the JSON proxy approach which often works better with GitHub API
      const res = await fetch(
        'https://api.github.com/repos/REXI-hub v2/REXI-hub v2/releases/latest',
        {
          headers: {
            Accept: 'application/json',
            'Content-Type': 'application/json',
            // Adding User-Agent which is often required by GitHub API
            'User-Agent': 'rexi-hub-v2-update-checker',
          },
        },
      ).then((response) => response.json());

      // Option 3: Use a local proxy endpoint
      // Create a cached version of the response to avoid frequent GitHub API calls
      // const res = await API.get('/api/status/github-latest-release');

      const { tag_name, body } = res;
      if (tag_name === statusState?.status?.version) {
        showSuccess(`Already the latest version: ${tag_name}`);
      } else {
        setUpdateData({
          tag_name: tag_name,
          content: marked.parse(body),
        });
        setShowUpdateModal(true);
      }
    } catch (error) {
      console.error('Failed to check for updates:', error);
      showError('Update check failed, please try again later');
    } finally {
      setLoadingInput((loadingInput) => ({
        ...loadingInput,
        CheckUpdate: false,
      }));
    }
  };
  const getOptions = async () => {
    const res = await API.get('/api/option/');
    const { success, message, data } = res.data;
    if (success) {
      let newInputs = {};
      data.forEach((item) => {
        if (item.key in inputs) {
          if (item.key === 'DisplayInCurrencyEnabled' || item.key === 'DisplayTokenStatEnabled') {
            newInputs[item.key] = item.value === 'true';
          } else {
            newInputs[item.key] = item.value;
          }
        }
      });
      setInputs(newInputs);
      formAPISettingGeneral.current.setValues(newInputs);
      formAPIPersonalization.current.setValues(newInputs);
    } else {
      showError(message);
    }
  };

  useEffect(() => {
    getOptions();
  }, []);

  // Function to open GitHub release page
  const openGitHubRelease = () => {
    window.open(
      `https://github.com/rexi-hub-v2/rexi-hub-v2/releases/tag/${updateData.tag_name}`,
      '_blank',
    );
  };

  const getStartTimeString = () => {
    const timestamp = statusState?.status?.start_time;
    return statusState.status ? timestamp2string(timestamp) : '';
  };

  return (
    <Row>
      <Col
        span={24}
        style={{
          marginTop: '10px',
          display: 'flex',
          flexDirection: 'column',
          gap: '10px',
        }}
      >
        {/* Thông tin phiên bản */}
        <Form>
          <Card>
            <Form.Section text={t('系统信息')}>
              <Row>
                <Col span={16}>
                  <Space>
                    <Text>
                      {t('当前版本')}：
                      {statusState?.status?.version || t('未知')}
                    </Text>
                    <Button
                      type='primary'
                      onClick={checkUpdate}
                      loading={loadingInput['CheckUpdate']}
                    >
                      {t('检查更新')}
                    </Button>
                  </Space>
                </Col>
              </Row>
              <Row>
                <Col span={16}>
                  <Text>
                    {t('启动时间')}：{getStartTimeString()}
                  </Text>
                </Col>
              </Row>
            </Form.Section>
          </Card>
        </Form>
        {/* Cài đặt chung */}
        <Form
          values={inputs}
          getFormApi={(formAPI) => (formAPISettingGeneral.current = formAPI)}
        >
          <Card>
            <Form.Section text={t('通用设置')}>
              <Form.TextArea
                label={t('公告')}
                placeholder={t(
                  '在此输入新的公告内容，支持 Markdown & HTML 代码',
                )}
                field={'Notice'}
                onChange={handleInputChange}
                style={{ fontFamily: 'JetBrains Mono, Consolas' }}
                autosize={{ minRows: 6, maxRows: 12 }}
              />
              <Button onClick={submitNotice} loading={loadingInput['Notice']}>
                {t('设置公告')}
              </Button>
            </Form.Section>
          </Card>
        </Form>
        {/* Cài đặt cá nhân hóa */}
        <Form
          values={inputs}
          getFormApi={(formAPI) => (formAPIPersonalization.current = formAPI)}
        >
          <Card>
            <Form.Section text={t('个性化设置')}>
              <Form.Input
                label={t('系统名称')}
                placeholder={t('在此输入系统名称')}
                field={'SystemName'}
                onChange={handleInputChange}
              />
              <Button
                onClick={submitSystemName}
                loading={loadingInput['SystemName']}
              >
                {t('设置系统名称')}
              </Button>
              <Form.Input
                label={t('Logo 图片地址')}
                placeholder={t('在此输入 Logo 图片地址')}
                field={'Logo'}
                onChange={handleInputChange}
              />
              <Button onClick={submitLogo} loading={loadingInput['Logo']}>
                {t('设置 Logo')}
              </Button>
              <Form.TextArea
                label={t('首页内容')}
                placeholder={t(
                  '在此输入首页内容，支持 Markdown & HTML 代码，设置后首页的状态信息将不再显示。如果输入的是一个链接，则会使用该链接作为 iframe 的 src 属性，这允许你设置任意网页作为首页',
                )}
                field={'HomePageContent'}
                onChange={handleInputChange}
                style={{ fontFamily: 'JetBrains Mono, Consolas' }}
                autosize={{ minRows: 6, maxRows: 12 }}
              />
              <Button
                onClick={() => submitOption('HomePageContent')}
                loading={loadingInput['HomePageContent']}
              >
                {t('设置首页内容')}
              </Button>
              <Form.TextArea
                label={t('关于')}
                placeholder={t(
                  '在此输入新的关于内容，支持 Markdown & HTML 代码。如果输入的是一个链接，则会使用该链接作为 iframe 的 src 属性，这允许你设置任意网页作为关于页面',
                )}
                field={'About'}
                onChange={handleInputChange}
                style={{ fontFamily: 'JetBrains Mono, Consolas' }}
                autosize={{ minRows: 6, maxRows: 12 }}
              />
              <Button onClick={submitAbout} loading={loadingInput['About']}>
                {t('设置关于')}
              </Button>
              {/*  */}
              <Banner
                fullMode={false}
                type='info'
                description={t(
                  '移除 REXI-hub 的版权标识不可在此处进行，且移除必须首先获得授权，项目维护需要花费大量精力，如果本项目对你有意义，请主动支持本项目',
                )}
                closeIcon={null}
                style={{ marginTop: 15 }}
              />
              <Form.Input
                label={t('页脚')}
                placeholder={t(
                  '在此输入新的页脚，留空则使用默认页脚，支持 HTML 代码',
                )}
                field={'Footer'}
                onChange={handleInputChange}
              />
              <Button onClick={submitFooter} loading={loadingInput['Footer']}>
                {t('设置页脚')}
              </Button>
              <Form.TextArea
                label={t('自定义头部 HTML')}
                placeholder={t(
                  '在此输入自定义 HTML 头部内容，将替换页面中的占位符'
                )}
                field={'custom_head_html'}
                onChange={handleInputChange}
                style={{ fontFamily: 'JetBrains Mono, Consolas' }}
                autosize={{ minRows: 6, maxRows: 12 }}
              />
              <Button onClick={submitCustomHeadHtml} loading={loadingInput['custom_head_html']}>
                {t('设置自定义头部 HTML')}
              </Button>
              <Form.TextArea
                label={t('全局 CSS 样式')}
                placeholder={t(
                  '在此输入自定义 CSS 样式代码'
                )}
                field={'global_css'}
                onChange={handleInputChange}
                style={{ fontFamily: 'JetBrains Mono, Consolas' }}
                autosize={{ minRows: 6, maxRows: 12 }}
              />
              <Button onClick={submitGlobalCss} loading={loadingInput['global_css']}>
                {t('设置全局 CSS 样式')}
              </Button>
              <Form.TextArea
                label={t('全局 JavaScript 代码')}
                placeholder={t(
                  '在此输入自定义 JavaScript 代码'
                )}
                field={'global_js'}
                onChange={handleInputChange}
                style={{ fontFamily: 'JetBrains Mono, Consolas' }}
                autosize={{ minRows: 6, maxRows: 12 }}
              />
              <Button onClick={submitGlobalJs} loading={loadingInput['global_js']}>
                {t('设置全局 JavaScript 代码')}
              </Button>
            </Form.Section>
            <Form.Section text={t('显示设置')}>
              <Space vertical align='start' style={{ width: '100%' }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                  <Switch
                    checked={inputs.DisplayInCurrencyEnabled}
                    onChange={async (checked) => {
                      setInputs({ ...inputs, DisplayInCurrencyEnabled: checked });
                      await updateOption('DisplayInCurrencyEnabled', checked.toString());
                    }}
                  />
                  <Text>{t('以货币形式显示配额')}</Text>
                </div>
                <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                  <Switch
                    checked={inputs.DisplayTokenStatEnabled}
                    onChange={async (checked) => {
                      setInputs({ ...inputs, DisplayTokenStatEnabled: checked });
                      await updateOption('DisplayTokenStatEnabled', checked.toString());
                    }}
                  />
                  <Text>{t('显示 Token 统计信息')}</Text>
                </div>
              </Space>
            </Form.Section>
          </Card>
        </Form>
      </Col>
      <Modal
        title={t('新版本') + '：' + updateData.tag_name}
        visible={showUpdateModal}
        onCancel={() => setShowUpdateModal(false)}
        footer={[
          <Button
            key='details'
            type='primary'
            onClick={() => {
              setShowUpdateModal(false);
              openGitHubRelease();
            }}
          >
            {t('详情')}
          </Button>,
        ]}
      >
        <div dangerouslySetInnerHTML={{ __html: updateData.content }}></div>
      </Modal>
    </Row>
  );
};

export default OtherSetting;
