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
import React from 'react';
import { useTranslation } from 'react-i18next';
import { Message } from 'semantic-ui-react';

const NotFound = () => {
  const { t } = useTranslation();
  return (
    <>
      <Message negative>
        <Message.Header>{t('页面不存在')}</Message.Header>
        <p>{t('请检查你的浏览器地址是否正确')}</p>
      </Message>
    </>
  );
};

export default NotFound;
