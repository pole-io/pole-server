import React, { useEffect, useRef } from 'react';
import { DiffEditor, Editor as MonacoEditor } from '@monaco-editor/react'
import type { DiffOnMount, MonacoDiffEditor, OnChange, OnMount } from '@monaco-editor/react'
import { Icon, LoadingIcon } from 'tdesign-icons-react';
import Style from './index.module.less';
import classNames from 'classnames';

export interface ICodeDiffEditorProps {
    namespace: string;
    group: string;
    filename: string;
    readonly?: boolean;
    allowFullScreen?: boolean;
    curValue?: string;
    nextValue?: string;
    language?: string;
    theme?: string;
    onChange?: (value: string | undefined, event: any) => void;
    onMount?: (editor: any, monaco: any) => void;
}

export function toHighlightLanguage(format?: string) {
    if (!format) {
        return "text"
    }
    if (format === "properties") {
        return 'ini'
    }
    if (format === 'yml') {
        return 'yaml'
    }
    return format
}

const CodeDiffEditor: React.FC<ICodeDiffEditorProps> = props => {
    return (
        <section
            className={classNames({
                [Style.monacoSection]: true,
            })}
            style={{
                height: 'calc(100vh - 370px)',
                width: "100%",
                position: 'relative'
            }}
        >
            <div style={{ display: 'flex', justifyContent: 'space-between', padding: '0 8px 4px 8px', fontSize: 12, color: '#888' }}>
                <span>{'当前版本'}</span>
                <span>{'目标版本'}</span>
            </div>
            <DiffEditor
                loading={<LoadingIcon />}
                theme={'vs'}
                options={{
                    // 控制是否只读
                    readOnly: true,
                    // 控制是否显示行号
                    lineNumbers: 'on',
                    // 控制布局自适应
                    automaticLayout: true,
                    //
                }}
                original={props.curValue}
                modified={props.nextValue}
                originalLanguage={toHighlightLanguage(props.language)}
                modifiedLanguage={toHighlightLanguage(props.language)}
            />
        </section>
    )
}

// 使用React.memo优化组件渲染
export default React.memo(CodeDiffEditor);
