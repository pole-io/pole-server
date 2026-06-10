import React from 'react';
import { Input } from 'tdesign-react';
import { SearchIcon } from 'tdesign-icons-react';
import Style from './Search.module.less';

const Search = ({ onChange, placeholder = '请输入搜索内容' }: { onChange: (value: string) => void, placeholder?: string }) => (
    <Input
        className={Style.panel}
        prefixIcon={<SearchIcon />}
        placeholder={placeholder}
        onEnter={(value) => onChange(value)}
        onClear={() => onChange('')}
    />
);

export default React.memo(Search);
