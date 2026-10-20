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
import React, { useContext, useEffect, useMemo, useState } from 'react';
import { Link, useNavigate, useLocation } from 'react-router-dom';
import { UserContext } from '../context/User';
import { StatusContext } from '../context/Status';
import { useTranslation } from 'react-i18next';

import {
  API,
  getLogo,
  getSystemName,
  isAdmin,
  isMobile,
  showError,
} from '../helpers';
import '../index.css';

import {
  IconCalendarClock,
  IconChecklistStroked,
  IconComment,
  IconCommentStroked,
  IconCreditCard,
  IconGift,
  IconHistogram,
  IconImage,
  IconKey,
  IconLayers,
  IconMail,
  IconSetting,
  IconUser,
} from '@douyinfe/semi-icons';
import {
  Nav,
  Divider,
} from '@douyinfe/semi-ui';
import { useSetTheme, useTheme } from '../context/Theme/index.js';
import { StyleContext } from '../context/Style/index.js';
import Text from '@douyinfe/semi-ui/lib/es/typography/text';

import i18n from '../i18n/i18n';
// Tùy chỉnh kiểu nút thanh bên
const navItemStyle = {
  borderRadius: '6px',
  margin: '4px 8px',
};

// Tùy chỉnh kiểu hover của nút thanh bên
const navItemHoverStyle = {
  backgroundColor: 'var(--semi-color-primary-light-default)',
  color: 'var(--semi-color-primary)',
};

// Tùy chỉnh kiểu nút thanh bên khi được chọn
const navItemSelectedStyle = {
  backgroundColor: 'var(--semi-color-primary-light-default)',
  color: 'var(--semi-color-primary)',
  fontWeight: '600',
};

// Tùy chỉnh kiểu icon
const iconStyle = (itemKey, selectedKeys) => {
  return {
    fontSize: '18px',
    color: selectedKeys.includes(itemKey)
      ? 'var(--semi-color-primary)'
      : 'var(--semi-color-text-2)',
  };
};

// Define routerMap as a constant outside the component
const routerMap = {
  home: '/',
  channel: '/admin/channels',
  token: '/app/tokens',
  redemption: '/admin/coupons',
  topup: '/app/wallet',
  user: '/admin/users',
  log: '/app/logs/api-usage',
  midjourney: '/app/logs/drawing',
  setting: '/admin/settings',
  about: '/about',
  detail: '/app/dashboard',
  pricing: '/pricing',
  task: '/app/logs/tasks',
  playground: '/playground',
  personal: '/app/me',
  message: '/admin/messages',
};

const SiderBar = () => {
  const { t } = useTranslation();
  const [styleState, styleDispatch] = useContext(StyleContext);
  const [statusState, statusDispatch] = useContext(StatusContext);
  const defaultIsCollapsed =
    localStorage.getItem('default_collapse_sidebar') === 'true';

  const [selectedKeys, setSelectedKeys] = useState(['home']);
  const [isCollapsed, setIsCollapsed] = useState(defaultIsCollapsed);
  const [chatItems, setChatItems] = useState([]);
  const [openedKeys, setOpenedKeys] = useState([]);
  const theme = useTheme();
  const setTheme = useSetTheme();
  const location = useLocation();
  const [routerMapState, setRouterMapState] = useState(routerMap);

  // Tính trước toàn bộ kiểu icon có thể có
  const allItemKeys = useMemo(() => {
    const keys = [
      'home',
      'channel',
      'token',
      'redemption',
      'topup',
      'user',
      'log',
      'midjourney',
      'setting',
      'about',
      'chat',
      'detail',
      'pricing',
      'task',
      'playground',
      'personal',
      'message',
    ];
    // Thêm keys cho mục trò chuyện
    for (let i = 0; i < chatItems.length; i++) {
      keys.push('chat' + i);
    }
    return keys;
  }, [chatItems]);

  // Dùng useMemo tính một lần toàn bộ kiểu của icon
  const iconStyles = useMemo(() => {
    const styles = {};
    allItemKeys.forEach((key) => {
      styles[key] = iconStyle(key, selectedKeys);
    });
    return styles;
  }, [allItemKeys, selectedKeys]);

  const workspaceItems = useMemo(
    () => [
      {
        text: t('数据看板'),
        itemKey: 'detail',
        to: '/app/dashboard',
        icon: <IconCalendarClock />,
        className:
          localStorage.getItem('enable_data_export') === 'true'
            ? ''
            : 'tableHiddle',
      },
      {
        text: t('API令牌'),
        itemKey: 'token',
        to: '/app/tokens',
        icon: <IconKey />,
      },
      {
        text: t('使用日志'),
        itemKey: 'log',
        to: '/app/logs/api-usage',
        icon: <IconHistogram />,
      },
      {
        text: t('绘图日志'),
        itemKey: 'midjourney',
        to: '/app/logs/drawing',
        icon: <IconImage />,
        className:
          localStorage.getItem('enable_drawing') === 'true'
            ? ''
            : 'tableHiddle',
      },
      {
        text: t('任务日志'),
        itemKey: 'task',
        to: '/app/logs/tasks',
        icon: <IconChecklistStroked />,
        className:
          localStorage.getItem('enable_task') === 'true' ? '' : 'tableHiddle',
      },
    ],
    [
      localStorage.getItem('enable_data_export'),
      localStorage.getItem('enable_drawing'),
      localStorage.getItem('enable_task'),
      t,
    ],
  );

  const financeItems = useMemo(
    () => [
      {
        text: t('计费与订阅'),
        itemKey: 'topup',
        to: '/app/wallet',
        icon: <IconCreditCard />,
      },
      {
        text: t('个人设置'),
        itemKey: 'personal',
        to: '/app/me',
        icon: <IconUser />,
      },
    ],
    [t],
  );

  const adminItems = useMemo(
    () => [
      {
        text: t('渠道'),
        itemKey: 'channel',
        to: '/admin/channels',
        icon: <IconLayers />,
        className: isAdmin() ? '' : 'tableHiddle',
      },
      {
        text: t('兑换码'),
        itemKey: 'redemption',
        to: '/admin/coupons',
        icon: <IconGift />,
        className: isAdmin() ? '' : 'tableHiddle',
      },
      {
        text: t('消息管理'),
        itemKey: 'message',
        to: '/admin/messages',
        icon: <IconMail />,
        className: isAdmin() ? '' : 'tableHiddle',
      },
      {
        text: t('用户管理'),
        itemKey: 'user',
        to: '/admin/users',
        icon: <IconUser />,
      },
      {
        text: t('系统设置'),
        itemKey: 'setting',
        to: '/admin/settings',
        icon: <IconSetting />,
      },
    ],
    [isAdmin(), t],
  );

  const chatMenuItems = useMemo(
    () => [
      {
        text: t('Playground'),
        itemKey: 'playground',
        to: '/playground',
        icon: <IconCommentStroked />,
      },
      // Only include chat menu if there are chat items
      ...(chatItems.length > 0
        ? [
            {
              text: t('聊天'),
              itemKey: 'chat',
              items: chatItems,
              icon: <IconComment />,
            },
          ]
        : []),
    ],
    [chatItems, t],
  );

  // Function to update router map with chat routes
  const updateRouterMapWithChats = (chats) => {
    const newRouterMap = { ...routerMap };

    if (Array.isArray(chats) && chats.length > 0) {
      for (let i = 0; i < chats.length; i++) {
        newRouterMap['chat' + i] = '/chat/' + i;
      }
    }

    setRouterMapState(newRouterMap);
    return newRouterMap;
  };

  // Update the useEffect for chat items
  useEffect(() => {
    let chats = localStorage.getItem('chats');
    if (chats) {
      try {
        chats = JSON.parse(chats);
        if (Array.isArray(chats)) {
          let chatItems = [];
          for (let i = 0; i < chats.length; i++) {
            let chat = {};
            for (let key in chats[i]) {
              chat.text = key;
              chat.itemKey = 'chat' + i;
              chat.to = '/chat/' + i;
            }
            chatItems.push(chat);
          }
          setChatItems(chatItems);

          // Update router map with chat routes
          updateRouterMapWithChats(chats);
        }
      } catch (e) {
        console.error(e);
        showError(i18n.t('Failed to parse chat data'));
      }
    }
  }, []);

  // Update the useEffect for route selection
  useEffect(() => {
    const currentPath = location.pathname;
    let matchingKey = Object.keys(routerMapState).find(
      (key) => routerMapState[key] === currentPath,
    );

    // Handle chat routes
    if (!matchingKey && currentPath.startsWith('/chat/')) {
      const chatIndex = currentPath.split('/').pop();
      if (!isNaN(chatIndex)) {
        matchingKey = 'chat' + chatIndex;
      } else {
        matchingKey = 'chat';
      }
    }

    // If we found a matching key, update the selected keys
    if (matchingKey) {
      setSelectedKeys([matchingKey]);
    }
  }, [location.pathname, routerMapState]);

  useEffect(() => {
    setIsCollapsed(styleState.siderCollapsed);
  }, [styleState.siderCollapsed]);

  // Custom divider style
  const dividerStyle = {
    margin: '8px 0',
    opacity: 0.6,
  };

  // Custom group label style
  const groupLabelStyle = {
    padding: '8px 16px',
    color: 'var(--semi-color-text-2)',
    fontSize: '12px',
    fontWeight: 'bold',
    textTransform: 'uppercase',
    letterSpacing: '0.5px',
  };

  return (
    <>
      <Nav
        className='custom-sidebar-nav'
        style={{
          width: isCollapsed ? '60px' : (openedKeys.includes('chat') ? '240px' : '200px'),
          boxShadow: '0 2px 8px rgba(0, 0, 0, 0.15)',
          borderRight: '1px solid var(--semi-color-border)',
          background: 'var(--semi-color-bg-1)',
          borderRadius: styleState.isMobile ? '0' : '0 8px 8px 0',
          position: 'relative',
          zIndex: 95,
          height: '100%',
          overflow: 'hidden', // Để CSS flexbox xử lý việc cuộn
          display: 'flex',
          flexDirection: 'column',
        }}
        limitIndent={false} // Cho phép điều khiển thụt lề linh hoạt hơn
        defaultIsCollapsed={
          localStorage.getItem('default_collapse_sidebar') === 'true'
        }
        isCollapsed={isCollapsed}
        onCollapseChange={(collapsed) => {
          setIsCollapsed(collapsed);
          // styleDispatch({ type: 'SET_SIDER', payload: true });
          styleDispatch({ type: 'SET_SIDER_COLLAPSED', payload: collapsed });
          localStorage.setItem('default_collapse_sidebar', collapsed);

          // Đảm bảo khi thu gọn thanh bên có mục được chọn, tránh tính toán không cần thiết
          if (selectedKeys.length === 0) {
            const currentPath = location.pathname;
            const matchingKey = Object.keys(routerMapState).find(
              (key) => routerMapState[key] === currentPath,
            );

            if (matchingKey) {
              setSelectedKeys([matchingKey]);
            } else if (currentPath.startsWith('/chat/')) {
              setSelectedKeys(['chat']);
            } else {
              setSelectedKeys(['detail']); // Mặc định chọn trang chủ
            }
          }
        }}
        selectedKeys={selectedKeys}
        itemStyle={navItemStyle}
        hoverStyle={navItemHoverStyle}
        selectedStyle={navItemSelectedStyle}
        renderWrapper={({ itemElement, isSubNav, isInSubNav, props }) => {
          return (
            <Link
              style={{ textDecoration: 'none' }}
              to={routerMapState[props.itemKey] || routerMap[props.itemKey]}
            >
              {itemElement}
            </Link>
          );
        }}
        onSelect={(key) => {
          if (key.itemKey.toString().startsWith('chat')) {
            styleDispatch({ type: 'SET_INNER_PADDING', payload: false });
          } else {
            styleDispatch({ type: 'SET_INNER_PADDING', payload: true });
          }

          // Nếu bấm vào mục cha của submenu đang mở thì thu gọn submenu
          if (openedKeys.includes(key.itemKey)) {
            setOpenedKeys(openedKeys.filter((k) => k !== key.itemKey));
          } else {
            setSelectedKeys([key.itemKey]);
          }
        }}
        openKeys={openedKeys}
        onOpenChange={(data) => {
          setOpenedKeys(data.openKeys);
        }}
      >
        {/* Chat Section - Only show if there are chat items */}
        {chatMenuItems.map((item) => {
          if (item.items && item.items.length > 0) {
            return (
              <Nav.Sub
                key={item.itemKey}
                itemKey={item.itemKey}
                text={item.text}
                icon={React.cloneElement(item.icon, {
                  style: iconStyles[item.itemKey],
                })}
              >
                {item.items.map((subItem) => (
                  <Nav.Item
                    key={subItem.itemKey}
                    itemKey={subItem.itemKey}
                    text={subItem.text}
                    level={1} // Đặt cấp của mục con để tối ưu thụt lề
                  />
                ))}
              </Nav.Sub>
            );
          } else {
            return (
              <Nav.Item
                key={item.itemKey}
                itemKey={item.itemKey}
                text={item.text}
                icon={React.cloneElement(item.icon, {
                  style: iconStyles[item.itemKey],
                })}
              />
            );
          }
        })}

        {/* Divider */}
        <Divider style={dividerStyle} />

        {/* Workspace Section */}
        {!isCollapsed && <Text style={groupLabelStyle}>{t('控制台')}</Text>}
        {workspaceItems.map((item) => (
          <Nav.Item
            key={item.itemKey}
            itemKey={item.itemKey}
            text={item.text}
            icon={React.cloneElement(item.icon, {
              style: iconStyles[item.itemKey],
            })}
            className={item.className}
          />
        ))}

        {isAdmin() && (
          <>
            {/* Divider */}
            <Divider style={dividerStyle} />

            {/* Admin Section */}
            {!isCollapsed && <Text style={groupLabelStyle}>{t('管理员')}</Text>}
            {adminItems.map((item) => (
              <Nav.Item
                key={item.itemKey}
                itemKey={item.itemKey}
                text={item.text}
                icon={React.cloneElement(item.icon, {
                  style: iconStyles[item.itemKey],
                })}
                className={item.className}
              />
            ))}
          </>
        )}

        {/* Divider */}
        <Divider style={dividerStyle} />

        {/* Finance Management Section */}
        {!isCollapsed && <Text style={groupLabelStyle}>{t('个人中心')}</Text>}
        {financeItems.map((item) => (
          <Nav.Item
            key={item.itemKey}
            itemKey={item.itemKey}
            text={item.text}
            icon={React.cloneElement(item.icon, {
              style: iconStyles[item.itemKey],
            })}
            className={item.className}
          />
        ))}

        <Nav.Footer
          collapseButton={true}
          collapseText={(collapsed) => {
            if (collapsed) {
              return t('展开侧边栏');
            }
            return t('收起侧边栏');
          }}
        />
      </Nav>
    </>
  );
};

export default SiderBar;
