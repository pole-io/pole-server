import React from 'react';
import { Input } from 'components/Fluent';
import { SearchIcon } from 'components/Fluent/icons';
import Style from './Search.module.less';

const Search = ({ onChange, placeholder = '请输入搜索内容', value }: { onChange: (value: string) => void, placeholder?: string, value?: string }) => (
    <Input
        className={Style.panel}
        prefixIcon={<SearchIcon />}
        placeholder={placeholder}
        value={value}
        onEnter={(value) => onChange(value)}
        onClear={() => onChange('')}
    />
);

export default React.memo(Search);
