import React from 'react';
import { Descriptions, Drawer, Radio, RadioGroup, Select, Space, Tag } from 'components/Fluent';

import { useAppDispatch, useAppSelector } from 'modules/store';
import { listActiveConfigFileRelease, listOneConfigFileRelease, selectFileRelease } from 'modules/configuration/release';
import CodeEditor from 'components/CodeEditor';
import { resolveFileFormat } from 'utils/path';
import CodeDiffEditor from 'components/CodeDiffEditor';
import Text from 'components/Text';
import { openErrNotification } from 'utils/notifition';

// active: true
// createBy: "HTTP:127.0.0.1"
// createTime: "2025-05-05 14:48:02"
// fileName: "conf/app/server.yaml"
// format: ""
// group: "123"
// id: "3"
// modifyBy: "HTTP:127.0.0.1"
// modifyTime: "2025-05-17 17:09:42"
// name: "v1"
// namespace: "default"
// releaseDescription: "v1"
// releaseType: "normal"
// tags: [{key: "123", value: "123"}, {key: "456", value: "123"}, {key: "789", value: "123"}]
// version: "2"

const { DescriptionsItem } = Descriptions;

interface IReleaseDetailProps {
    namespace: string;
    group: string;
    fileName: string;
    visible: boolean;
    onClose: () => void;
}

const ReleaseDetail: React.FC<IReleaseDetailProps> = ({ namespace, group, fileName, visible, onClose }) => {
    const dispatch = useAppDispatch();

    const releaseState = useAppSelector(selectFileRelease);
    const { versions = [], editFileRelease, viewFileRelease, activeFileRelease } = releaseState;

    const [openDiff, setOpenDiff] = React.useState<boolean>(false);

    React.useEffect(() => {
        if (visible) {
            dispatch(listActiveConfigFileRelease({
                param: {
                    namespace: namespace,
                    group: group,
                    file_name: fileName,
                    release_name: editFileRelease?.name || ''
                }
            })).then((res) => {
                if (res.meta.requestStatus === 'rejected') {
                    openErrNotification('获取版本列表失败', res.payload as string);
                }
            })
        }
    }, [visible]);

    const fetchDiffData = (release_name: string) => {
        dispatch(listOneConfigFileRelease({
            param: {
                namespace: namespace,
                group: group,
                file_name: fileName,
                release_name: release_name
            }
        })).then((res) => {
            if (res.meta.requestStatus === 'rejected') {
                openErrNotification('获取版本详情失败', res.payload as string);
            }
        });
    }

    const renderHeader = (
        <>
            <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                <Space direction='horizontal'>
                    <Text>
                        {editFileRelease?.name} 版本详情
                    </Text>
                    <RadioGroup
                        theme='button'
                        variant='primary-filled'
                        value={openDiff ? 'diff' : 'single'}
                        onChange={(value) => {
                            setOpenDiff(value === 'diff');
                        }}
                    >
                        <Radio.Button value="single">单版本查看</Radio.Button>
                        <Radio.Button value="diff">多版本比对</Radio.Button>
                    </RadioGroup>
                    {openDiff && (
                        <Select
                            label="目标对比版本："
                            options={versions.map((item) => ({
                            label: item.name,
                                value: item.name,
                                disabled: item.name === editFileRelease?.name
                            }))}
                            onChange={(value) => {
                                fetchDiffData(value as string);
                            }}
                        />
                    )}
                </Space>
            </div>
        </>
    )

    return (
        <>
            <Drawer
                visible={visible}
                header={renderHeader}
                size={'60%'}
                cancelBtn={false}
                onClose={onClose}
            >

                <Descriptions column={2} size='small'>
                    <DescriptionsItem label="命名空间">{activeFileRelease?.namespace}</DescriptionsItem>
                    <DescriptionsItem label="配置组">{activeFileRelease?.group}</DescriptionsItem>
                    <DescriptionsItem label="文件名">{activeFileRelease?.fileName}</DescriptionsItem>
                    <DescriptionsItem label="版本">{activeFileRelease?.version}</DescriptionsItem>
                    <DescriptionsItem label="发布类型">{activeFileRelease?.releaseType}</DescriptionsItem>
                    <DescriptionsItem label="配置类型">
                        {activeFileRelease?.configType === 'CONFIG_TEMPLATE' ? '模板渲染' : '普通文本'}
                    </DescriptionsItem>
                    {activeFileRelease?.configType === 'CONFIG_TEMPLATE' && (
                        <>
                            <DescriptionsItem label="Template">{activeFileRelease.templateBinding?.templateId || '-'}</DescriptionsItem>
                            <DescriptionsItem label="Template Release">
                                {activeFileRelease.templateBinding?.templateReleaseId || '-'}
                            </DescriptionsItem>
                            <DescriptionsItem label="Binding Release">
                                {activeFileRelease.templateBinding?.bindingReleaseId || '-'}
                            </DescriptionsItem>
                        </>
                    )}
                    <DescriptionsItem label="发布描述">{activeFileRelease?.releaseDescription}</DescriptionsItem>
                    <DescriptionsItem label="创建时间">{activeFileRelease?.createTime}</DescriptionsItem>
                    <DescriptionsItem label="创建人">{activeFileRelease?.createBy}</DescriptionsItem>
                    <DescriptionsItem label="修改时间">{activeFileRelease?.modifyTime}</DescriptionsItem>
                    <DescriptionsItem label="修改人">{activeFileRelease?.modifyBy}</DescriptionsItem>
                    <DescriptionsItem label="加密状态">{activeFileRelease?.tags?.find(item => item.key === 'internal-encrypted')?.value ? '开启' : '-'}</DescriptionsItem>
                    <DescriptionsItem label="加密算法">{activeFileRelease?.tags?.find(item => item.key === 'internal-encryptalgo')?.value || '-'}</DescriptionsItem>
                    <DescriptionsItem label="文件标签">
                        <Space>
                            {activeFileRelease?.tags?.map((item) => (
                                <Tag variant='outline'>
                                    {item.key} : {item.value}
                                </Tag>
                            ))}
                        </Space>
                    </DescriptionsItem>
                </Descriptions>

                <div style={{ marginTop: '20px' }}>
                    {openDiff ? (
                        <>
                            <CodeDiffEditor
                                namespace={activeFileRelease?.namespace || ''}
                                group={activeFileRelease?.group || ''}
                                filename={activeFileRelease?.fileName || ''}
                                language={resolveFileFormat(activeFileRelease?.fileName)}
                                curValue={activeFileRelease?.content}
                                nextValue={viewFileRelease?.content}
                                readonly={true}
                                allowFullScreen={false}
                            />
                        </>
                    ) : (
                        <CodeEditor
                            language={resolveFileFormat(activeFileRelease?.fileName)}
                            value={activeFileRelease?.content}
                            readonly={true}
                            allowFullScreen={false}
                        />
                    )}
                </div>
            </Drawer>
        </>
    )
}

export default React.memo(ReleaseDetail);
