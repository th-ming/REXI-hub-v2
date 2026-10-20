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
import i18n from '../../i18n/i18n';
import React from 'react';
import TokensTable from '../../components/TokensTable';
import { Banner, Layout } from '@douyinfe/semi-ui';
import { useTranslation } from 'react-i18next';
import { useContext } from 'react';
import { StatusContext } from '../../context/Status';
const Token = () => {
  const { t } = useTranslation();
  const [statusState, statusDispatch] = useContext(StatusContext);
  return (
    <>
      <Layout>
        <Layout.Header>
          <Banner
            type='warning'
            description={t(
              '令牌无法精确控制使用额度，只允许自用，请勿直接将令牌分发给他人。',
            )}
          />
          {/* Warning banner for chat content logging */}
          {statusState?.status?.log_chat_content_enabled && (
            <Banner
              type='warning'
              description={i18n.t('Administrators of this site can view your conversations')}
              style={{
                margin: '0 0 16px 0',
                borderRadius: '6px',
              }}
            />
          )}
        </Layout.Header>
        <Layout.Content>
          <TokensTable />
        </Layout.Content>
      </Layout>
    </>
  );
};

export default Token;
