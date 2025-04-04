import React from 'react';
import { useTranslation } from 'react-i18next';
import { Dropdown, Menu } from 'antd';
import { GlobalOutlined } from '@ant-design/icons';

interface Language {
  code: string;
  name: string;
  flag: string;
}

const languages: Language[] = [
  { code: 'en', name: 'English', flag: '🇺🇸' },
  { code: 'zh', name: '中文', flag: '🇨🇳' },
];

const LanguageSwitcher: React.FC = () => {
  const { i18n } = useTranslation();
  
  const changeLanguage = (langCode: string) => {
    i18n.changeLanguage(langCode);
  };
  
  const menu = (
    <Menu>
      {languages.map((lang) => (
        <Menu.Item 
          key={lang.code} 
          onClick={() => changeLanguage(lang.code)}
          className={i18n.language === lang.code ? 'ant-dropdown-menu-item-active' : ''}
        >
          <span style={{ marginRight: 8 }}>{lang.flag}</span>
          {lang.name}
        </Menu.Item>
      ))}
    </Menu>
  );

  return (
    <Dropdown overlay={menu} trigger={['click']} placement="bottomRight">
      <div className="language-switcher" style={{ cursor: 'pointer', display: 'flex', alignItems: 'center' }}>
        <GlobalOutlined style={{ fontSize: 16, marginRight: 4 }} />
        <span>{languages.find(lang => lang.code === i18n.language)?.name || 'English'}</span>
      </div>
    </Dropdown>
  );
};

export default LanguageSwitcher;