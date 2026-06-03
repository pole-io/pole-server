import React from 'react';
import { Input } from 'tdesign-react';
import { SearchIcon } from 'tdesign-icons-react';
import Style from './Search.module.less';

const Search = ({ onChange }: { onChange: (value: string) => void }) => (
    <Input
        className={Style.panel}
        prefixIcon={<SearchIcon />}
        placeholder='请输入搜索内容'
        onEnter={(value) => onChange(value)}
        onClear={() => onChange('')}
    />
);

export default React.memo(Search);
