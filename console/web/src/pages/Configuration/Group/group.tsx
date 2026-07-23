import React, { } from 'react';
import { Button, Input, PrimaryTableProps, Select, Space, Table, TableRowData, Tag, Tooltip } from 'components/Fluent';
import { AddIcon, RefreshIcon } from 'components/Fluent/icons';

import Text from 'components/Text';
import { ConfirmOperationButton, OperationButton } from 'components/OperationButton';
import { ResourceHeader, ResourceToolbar } from 'components/ResourceLayout';
import { useAppDispatch, useAppSelector } from 'modules/store';
import { useNavigate } from 'react-router-dom';
import ConfigGroupEditor from './ConfigGroupEditor';

import style from './index.module.less';
import { cleanConfigGroupPage, editorConfigGroup, listConfigGroups, removeConfigGroups, resetConfigGroup, selectConfigGroup } from 'modules/configuration/group';
import { ConfigFileGroupView } from 'services/config_group';
import { describeAllConfigFiles } from 'services/config_files';
import AuthorizeInput from 'components/Authorize';
import { PolicySourceType } from 'services/auth_policy';
import { Op } from 'services/types';
import { openErrNotification, openInfoNotification } from 'utils/notifition';
import ResourceNameLink from 'components/ResourceNameLink';

interface IConfigGroupTableProps {
}

type FilterState = {
	keyword: string;
	namespace: string;
	publishStatus: string;
};

const defaultFilterState: FilterState = {
	keyword: '',
	namespace: '',
	publishStatus: '',
};

const metadataValue = (row: TableRowData, key: string) => {
	const metadata = row.metadata as Record<string, string> | undefined;
	return metadata?.[key] || '';
};

const hasPendingRelease = (row: TableRowData) => {
	const pending = Number(metadataValue(row, 'pendingReleaseCount') || 0);
	return pending > 0 || row.status === 'to-be-released';
};

const groupKey = (namespace?: string, name?: string) => `${namespace || ''}/${name || ''}`;

const columns = (
	operateConfigGroup: (op: Op, row: TableRowData) => void,
	redirect: (row: TableRowData) => void,
	encryptedFileCounts: Record<string, number>,
): PrimaryTableProps['columns'] => [
	{
		colKey: 'name',
		title: '分组名称',
		cell: ({ row }) => <ResourceNameLink name={row.name} onClick={() => redirect(row)} />,
	},
	{
		colKey: 'namespace',
		title: '命名空间',
		cell: ({ row: { namespace } }) => (<Text>{namespace}</Text>),
	},
	{
		colKey: 'fileCount',
		title: '文件数',
		cell: ({ row: { fileCount } }) => (<Text>{fileCount}</Text>),
	},
	{
		colKey: 'encryptedFileCount',
		title: '加密数',
		cell: ({ row }) => (<Text>{encryptedFileCounts[groupKey(row.namespace, row.name)] || 0}</Text>),
	},
	{
		colKey: 'pendingRelease',
		title: '待发布',
		cell: ({ row }) => hasPendingRelease(row)
			? <Tag theme="warning" variant="light">有变更</Tag>
			: <Tag theme="success" variant="light">无</Tag>,
	},
	{
		colKey: 'action',
		title: '操作',
		cell: ({ row }) => {
			const fileCount = Number(row.fileCount || 0);
			const deleteBlockedByFiles = fileCount > 0;
			return (
				<Space>
					<OperationButton action="view" disabled={row.editable === false} disabledLabel="无权限操作" onClick={() => operateConfigGroup('edit', row)} />
					<OperationButton action="authorize" disabled={row.editable === false} disabledLabel="无权限操作" onClick={() => operateConfigGroup('authorize', row)} />
					<ConfirmOperationButton
						action="delete"
						disabled={row.deleteable === false || fileCount > 0}
						disabledLabel={deleteBlockedByFiles ? `请先删除 ${fileCount} 个配置文件` : '无权限操作'}
						confirmContent="确认删除吗"
						onConfirm={() => operateConfigGroup('delete', row)}
					/>
				</Space>
			)
		},
	},
];


const ConfigGroupTable: React.FC<IConfigGroupTableProps> = ({ }) => {
	const dispatch = useAppDispatch();
	const navigate = useNavigate()

	const groupState = useAppSelector(selectConfigGroup);
	const { datas, total, page, limit, loading } = groupState;
	const [filterState, setFilterState] = React.useState<FilterState>(defaultFilterState);
	const [encryptedFileCounts, setEncryptedFileCounts] = React.useState<Record<string, number>>({});

	// 合并编辑相关状态
	const [editorState, setEditorState] = React.useState<{
		visible: boolean;
		mode: 'create' | 'edit' | 'delete';
		data?: TableRowData;
		authorizeVisible: boolean;
	}>({ visible: false, mode: 'create', data: undefined, authorizeVisible: false });


	// 编辑、新建事件
	const operateConfigGroup = (op: Op, row?: TableRowData) => {
		switch (op) {
			case 'create':
				dispatch(resetConfigGroup());
				setEditorState(prev => ({ ...prev, visible: true, mode: 'create', data: undefined }));
				return;
			case 'edit':
				dispatch(editorConfigGroup(row as ConfigFileGroupView));
				setEditorState(prev => ({ ...prev, visible: true, mode: 'edit', data: { ...row } }));
				return;
			case 'delete':
				dispatch(removeConfigGroups({ param: { id: row?.id, namespace: row?.namespace, name: row?.name } }))
				.then((res) => {
					if (res.meta.requestStatus === 'rejected') {
						// 删除失败
						openErrNotification('请求失败', res.payload);
					} else {
						// 删除成功
						openInfoNotification('请求成功', `删除分组 ${row?.name} 成功`);
						refreshTable();
					}
				});
				return;
			case 'authorize':
				setEditorState(prev => ({ ...prev, authorizeVisible: true, data: { ...row } }));
				return;
		}
	}

	React.useEffect(() => {
		refreshTable(1, limit, defaultFilterState);
		return () => {
			dispatch(cleanConfigGroupPage());
		}
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, []);

	React.useEffect(() => {
		let cancelled = false;
		if (!datas.length) {
			setEncryptedFileCounts({});
			return () => {
				cancelled = true;
			};
		}
		Promise.all(datas.map(async (item) => {
			const key = groupKey(item.namespace, item.name);
			try {
				const files = await describeAllConfigFiles({
					namespace: item.namespace,
					group: item.name,
				});
				return [key, files.filter((file) => Boolean(file.encrypted)).length] as const;
			} catch {
				return [key, 0] as const;
			}
		})).then((items) => {
			if (!cancelled) {
				setEncryptedFileCounts(Object.fromEntries(items));
			}
		});
		return () => {
			cancelled = true;
		};
	}, [datas]);

	const refreshTable = (page = 1, limit = 10, filters = filterState) => {
		dispatch(listConfigGroups({
			param: {
				offset: (page - 1) * limit,
				limit: limit,
				namespace: filters.namespace || filters.keyword,
				group: filters.keyword,
			}
		}))
	}

	const namespaceOptions = React.useMemo(() => {
		const namespaces = Array.from(new Set(datas.map((item) => item.namespace).filter(Boolean)));
		return [
			{ label: '全部命名空间', value: '' },
			...namespaces.map((item) => ({ label: item, value: item })),
		];
	}, [datas]);

	const filteredDatas = React.useMemo(() => {
		return datas.filter((item) => {
			if (filterState.publishStatus === 'pending' && !hasPendingRelease(item)) {
				return false;
			}
			if (filterState.publishStatus === 'clean' && hasPendingRelease(item)) {
				return false;
			}
			return true;
		});
	}, [datas, filterState.publishStatus]);

	const summaryStats = React.useMemo(() => {
		const groupCount = datas.length;
		const fileCount = datas.reduce((acc, item) => acc + Number(item.fileCount || 0), 0);
		const encryptedFileCount = datas.reduce((acc, item) => acc + (encryptedFileCounts[groupKey(item.namespace, item.name)] || 0), 0);
		const pendingCount = datas.filter(hasPendingRelease).length;
		return {
			groupCount,
			fileCount,
			encryptedFileCount,
			pendingCount,
		};
	}, [datas, encryptedFileCounts]);

	const updateFilters = (next: Partial<FilterState>) => {
		const merged = { ...filterState, ...next };
		setFilterState(merged);
	};

	const submitFilter = () => {
		refreshTable(1, limit, filterState);
	};

	const resetFilter = () => {
		setFilterState(defaultFilterState);
		refreshTable(1, limit, defaultFilterState);
	};

	const table = (
		<>
			<ResourceHeader
				eyebrow="配置中心 / 配置分组"
				title="配置分组"
				description="同名分组及其同名文件表示同一配置资源的不同环境实例，各命名空间独立维护内容与发布状态。"
				actions={(
					<>
						<Tooltip content="刷新列表">
							<Button shape="square" variant="outline" icon={<RefreshIcon />} onClick={() => refreshTable(page, limit)} />
						</Tooltip>
						<Button theme="primary" icon={<AddIcon />} onClick={() => operateConfigGroup('create')}>新建配置分组</Button>
					</>
				)}
			/>
			<div className={style.summaryBar}>
				<div className={style.summaryItem}>
					<span>分组数</span>
					<strong>{summaryStats.groupCount}</strong>
				</div>
				<div className={style.summaryItem}>
					<span>配置文件数</span>
					<strong>{summaryStats.fileCount}</strong>
				</div>
				<div className={style.summaryItem}>
					<span>配置加密数</span>
					<strong>{summaryStats.encryptedFileCount}</strong>
				</div>
				<div className={`${style.summaryItem} ${style.warningStat}`}>
					<span>待发布数</span>
					<strong>{summaryStats.pendingCount}</strong>
				</div>
			</div>
			<ResourceToolbar
				title="配置分组列表"
				count={loading ? '正在同步列表' : `当前显示 ${filteredDatas.length} 条`}
				filters={(
					<>
						<Input
							className={style.filterKeyword}
							clearable
							label="关键字"
							placeholder="分组名称"
							value={filterState.keyword}
							onChange={(value) => updateFilters({ keyword: value as string })}
							onEnter={submitFilter}
						/>
						<Select
							className={style.filterSelect}
							label="命名空间"
							options={namespaceOptions}
							value={filterState.namespace}
							onChange={(value) => updateFilters({ namespace: value as string })}
						/>
						<Select
							className={style.filterSelect}
							label="发布状态"
							options={[
								{ label: '全部', value: '' },
								{ label: '待发布', value: 'pending' },
								{ label: '无待发布', value: 'clean' },
							]}
							value={filterState.publishStatus}
							onChange={(value) => updateFilters({ publishStatus: value as string })}
						/>
						<Button variant="outline" onClick={submitFilter}>查询</Button>
						<Button variant="text" onClick={resetFilter}>重置</Button>
					</>
				)}
			/>
			{editorState.visible && (
				<ConfigGroupEditor
					key={editorState.mode + (editorState.data?.name || 'new') + (editorState.visible ? '1' : '0')}
					op={editorState.mode}
					visible={editorState.visible}
					closeDrawer={() => {
						// 关闭后重置编辑器状态
						dispatch(resetConfigGroup());
						setEditorState(s => ({ ...s, visible: false }));
						refreshTable();
					}}
				/>
			)}
			{editorState.authorizeVisible && editorState.data?.id && (
                <AuthorizeInput
                    resource_type={PolicySourceType.ConfigGroups}
                    resource_id={editorState.data.id}
                    resource_name={`${editorState.data.namespace}/${editorState.data.name}`}
                    visible={editorState.authorizeVisible}
                    onClose={() => {
                        setEditorState(s => ({ ...s, authorizeVisible: false }));
                    }}
                />
            )}
			<div className={style.groupTableSurface}>
				<Table
					data={filteredDatas}
					columns={columns(operateConfigGroup, (row: TableRowData) => {
						// 发送行信息给到配置文件列表页面，需要限制某些操作
						dispatch(editorConfigGroup({ ...row as ConfigFileGroupView }));
						navigate(`files?namespace=${row.namespace}&group=${row.name}`);
					}, encryptedFileCounts)}
					loading={loading}
					rowKey="id"
					size={"large"}
					tableLayout={'auto'}
					cellEmptyContent={'-'}
					pagination={{
						current: page,
						pageSize: limit,
						total: filterState.publishStatus ? filteredDatas.length : total,
						showJumper: true,
						onChange(pageInfo) {
							refreshTable(pageInfo.current, pageInfo.pageSize);
						},
					}}
					onPageChange={(pageInfo) => {
						refreshTable(pageInfo.current, pageInfo.pageSize);
					}}
				/>
			</div>
		</>
	)

	return (
		<>
			{table}
		</>
	);
}

export default React.memo(ConfigGroupTable);
