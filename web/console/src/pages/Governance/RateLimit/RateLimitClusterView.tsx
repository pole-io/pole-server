import React, { useState } from 'react';
import { describeLimitClusters, RateLimitCluster } from 'services/ratelimit';
import { Card, Space } from 'components/Fluent';
import { openErrNotification } from 'utils/notifition';

export interface IRateLimitClusterProps {
}

const RateLimitClusterView: React.FC<IRateLimitClusterProps> = (props) => {

    const [clusters, setClusters] = useState<RateLimitCluster[]>([]);

    React.useEffect(() => {
        const fetchClusters = async () => {
            try {
                const ret = await describeLimitClusters();
                setClusters(ret.list || []);
            } catch (err) {
                console.error('获取限流集群列表失败', err);
                openErrNotification('请求失败', '获取限流集群列表失败，请稍后重试');
            }
        };
        fetchClusters();
    }, [])

    return (
        <>
            <div style={{ margin: 20 }}>
                <h3>限流集群列表</h3>
                {clusters.length > 0 ? (
                    <Space>
                        {clusters.map((cluster) => (
                            <Card

                            />
                        ))}
                    </Space>
                ) : (
                    <p>暂无限流集群</p>
                )}
            </div>
            <div style={{ margin: 20 }}>
                <h3>限流集群操作</h3>
                <p>目前仅支持查看限流集群列表，后续将添加更多功能。</p>
            </div>
        </>
    )
}

export default React.memo(RateLimitClusterView);