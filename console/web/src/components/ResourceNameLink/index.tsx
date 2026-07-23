import React from 'react';

import { Link } from 'components/Fluent';
import style from './index.module.less';

interface ResourceNameLinkProps {
    name?: React.ReactNode;
    onClick: () => void;
    className?: string;
    ariaLabel?: string;
}

const ResourceNameLink: React.FC<ResourceNameLinkProps> = ({ name, onClick, className, ariaLabel }) => {
    if (name === undefined || name === null || name === '') {
        return <span>-</span>;
    }

    const title = typeof name === 'string' || typeof name === 'number' ? String(name) : undefined;

    return (
        <Link
            className={`${style.link}${className ? ` ${className}` : ''}`}
            title={title}
            aria-label={ariaLabel || (title ? `查看 ${title}` : '查看资源')}
            onClick={(event: React.MouseEvent<HTMLElement>) => {
                event.stopPropagation();
                onClick();
            }}
        >
            {name}
        </Link>
    );
};

export default React.memo(ResourceNameLink);
