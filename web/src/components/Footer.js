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
import React, { useEffect, useState } from 'react';
import { getFooterHTML } from '../helpers';
import i18n from '../i18n/i18n';

const FooterBar = () => {
  const [footer, setFooter] = useState(getFooterHTML());
  const [poweredByText, setPoweredByText] = useState('');
  let remainCheckTimes = 5;

  const poweredByKeys = [
    'Powered by {{brand}}',
    'Features & speed powered by {{brand}}',
    'Made with ♥️ by {{brand}}',
    'Powerfully built on {{brand}}',
    'Built on {{brand}}',
    'Driven by {{brand}}',
    '{{brand}} Powered',
  ];

  const loadFooter = () => {
    let footer_html = localStorage.getItem('footer_html');
    if (footer_html) {
      setFooter(footer_html);
    }
  };

  useEffect(() => {
    // Set random powered by text on component mount
    const randomIndex = Math.floor(Math.random() * poweredByKeys.length);
    setPoweredByText(
      i18n.t(poweredByKeys[randomIndex], { brand: 'REXI-hub' }),
    );

    const timer = setInterval(() => {
      if (remainCheckTimes <= 0) {
        clearInterval(timer);
        return;
      }
      remainCheckTimes--;
      loadFooter();
    }, 200);
    return () => clearTimeout(timer);
  }, []);

  const PoweredByLink = (
    <a 
      href={'/about'} 
      rel='noreferrer'
      style={{ 
        textDecoration: 'none', 
        color: 'inherit',
        fontSize: '14px'
      }}
      onMouseEnter={(e) => e.target.style.textDecoration = 'underline'}
      onMouseLeave={(e) => e.target.style.textDecoration = 'none'}
    >
      {poweredByText}
    </a>
  );

  let content;
  if (footer) {
    content = (
      <div style={{ 
        display: 'flex', 
        alignItems: 'center', 
        justifyContent: 'space-between',
        gap: '20px',
        maxWidth: '1200px',
        margin: '0 auto',
        padding: '0 20px'
      }}>
        <div className='custom-footer' dangerouslySetInnerHTML={{ __html: footer }}></div>
        <div>{PoweredByLink}</div>
      </div>
    );
  } else {
    content = (
      <div style={{ textAlign: 'right', padding: '0 20px' }}>
        {PoweredByLink}
      </div>
    );
  }

  return (
    <div style={{ paddingBottom: '28px', marginLeft: '10px', marginRight: '10px' }}>
      {content}
    </div>
  );
};

export default FooterBar;