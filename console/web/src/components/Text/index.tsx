import React, { CSSProperties } from 'react';

// 封装成组件
const Text = ({ children, style }: { children: React.ReactNode; style?: CSSProperties }) => {
    return <span style={style}>{children}</span>;
};

export default React.memo(Text);