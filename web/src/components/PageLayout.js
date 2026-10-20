/*
Copyright (c) 2026 REXI-hub team

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
import HeaderBar from './HeaderBar.js';
import { Layout } from '@douyinfe/semi-ui';
import SiderBar from './SiderBar.js';
import App from '../App.js';
import FooterBar from './Footer.js';
import { ToastContainer } from 'react-toastify';
import React, { useContext, useEffect } from 'react';
import { StyleContext } from '../context/Style/index.js';
import { useTranslation } from 'react-i18next';
import { LocaleProvider } from '@douyinfe/semi-ui';
import semiEnUS from '@douyinfe/semi-ui/lib/es/locale/source/en_US';
import semiViVN from '@douyinfe/semi-ui/lib/es/locale/source/vi_VN';
import semiZhCN from '@douyinfe/semi-ui/lib/es/locale/source/zh_CN';
import { API, getLogo, getSystemName, showError } from '../helpers/index.js';
import { setStatusData } from '../helpers/data.js';
import { UserContext } from '../context/User/index.js';
import { StatusContext } from '../context/Status/index.js';
import i18n from '../i18n/i18n';
const { Sider, Content, Header, Footer } = Layout;

const PageLayout = () => {
  const [userState, userDispatch] = useContext(UserContext);
  const [statusState, statusDispatch] = useContext(StatusContext);
  const [styleState, styleDispatch] = useContext(StyleContext);
  const { i18n } = useTranslation();

  const semiLocale =
    i18n.language && i18n.language.startsWith('vi')
      ? semiViVN
      : i18n.language && i18n.language.startsWith('zh')
        ? semiZhCN
        : semiEnUS;

  const loadUser = () => {
    let user = localStorage.getItem('user');
    if (user) {
      let data = JSON.parse(user);
      userDispatch({ type: 'login', payload: data });
    }
  };

  const loadStatus = async () => {
    try {
      const res = await API.get('/api/status');
      const { success, data } = res.data;
      if (success) {
        statusDispatch({ type: 'set', payload: data });
        setStatusData(data);
        // Notify other components that status has been updated
        window.dispatchEvent(new Event('statusUpdated'));
      } else {
        showError(i18n.t('Unable to connect to server'));
      }
    } catch (error) {
      showError(i18n.t('Failed to load status'));
    }
  };

  useEffect(() => {
    loadUser();
    loadStatus().catch(console.error);
    let systemName = getSystemName();
    if (systemName) {
      document.title = systemName;
    }
    let logo = getLogo();
    if (logo) {
      let linkElement = document.querySelector("link[rel~='icon']");
      if (linkElement) {
        linkElement.href = logo;
      }
    }
    // Lấy ngôn ngữ dùng lần trước từ localStorage
    const savedLang = localStorage.getItem('i18nextLng');
    if (savedLang) {
      i18n.changeLanguage(savedLang);
    }

    // Mặc định hiển thị thanh bên
    styleDispatch({ type: 'SET_SIDER', payload: true });
  }, [i18n]);

  // Lấy trạng thái thu gọn thanh bên
  const isSidebarCollapsed =
    localStorage.getItem('default_collapse_sidebar') === 'true';

  return (
    <LocaleProvider locale={semiLocale}>
    <Layout
      style={{
        height: '100vh',
        display: 'flex',
        flexDirection: 'column',
        overflow: styleState.isMobile ? 'visible' : 'hidden',
      }}
    >
      <Header
        style={{
          padding: 0,
          height: 'auto',
          lineHeight: 'normal',
          position: styleState.isMobile ? 'sticky' : 'fixed',
          width: '100%',
          top: 0,
          zIndex: 100,
          boxShadow: '0 1px 6px rgba(0, 0, 0, 0.08)',
        }}
      >
        <HeaderBar />
      </Header>
      <Layout
        style={{
          marginTop: styleState.isMobile ? '0' : '56px',
          height: styleState.isMobile ? 'auto' : 'calc(100vh - 56px)',
          overflow: styleState.isMobile ? 'visible' : 'auto',
          display: 'flex',
          flexDirection: 'column',
        }}
      >
        {styleState.showSider && (
          <Sider
            style={{
              position: 'fixed',
              left: 0,
              top: '56px',
              zIndex: 99,
              background: 'var(--semi-color-bg-1)',
              boxShadow: '0 2px 8px rgba(0, 0, 0, 0.15)',
              border: 'none',
              paddingRight: '0',
              height: 'calc(100vh - 56px)',
            }}
          >
            <SiderBar />
          </Sider>
        )}
        <Layout
          style={{
            marginLeft: styleState.isMobile
              ? '0'
              : styleState.showSider
                ? styleState.siderCollapsed
                  ? '60px'
                  : '200px'
                : '0',
            transition: 'margin-left 0.3s ease',
            flex: '1 1 auto',
            display: 'flex',
            flexDirection: 'column',
          }}
        >
          <Content
            style={{
              flex: '1 0 auto',
              overflowY: styleState.isMobile ? 'visible' : 'auto',
              WebkitOverflowScrolling: 'touch',
              padding: styleState.shouldInnerPadding ? '24px' : '0',
              position: 'relative',
              marginTop: styleState.isMobile ? '2px' : '0',
            }}
          >
            <App />
          </Content>
          <Layout.Footer
            style={{
              flex: '0 0 auto',
              width: '100%',
            }}
          >
            <FooterBar />
          </Layout.Footer>
        </Layout>
      </Layout>
      <ToastContainer />
    </Layout>
    </LocaleProvider>
  );
};

export default PageLayout;
