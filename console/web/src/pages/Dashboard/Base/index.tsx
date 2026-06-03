import React, { memo } from 'react';
import { Card } from 'tdesign-react';

const DashBoard = () => (
  <div style={{ overflowX: 'hidden' }}>
    <Card></Card>
  </div>
);

export default memo(DashBoard);
