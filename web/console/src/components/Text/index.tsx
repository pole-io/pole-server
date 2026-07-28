import React from 'react';

const Text = ({ children, ...props }: React.HTMLAttributes<HTMLSpanElement>) => {
    return <span {...props}>{children}</span>;
};

export default React.memo(Text);
