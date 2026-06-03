import { Notification, NotificationPlugin, NotificationInfoOptions } from 'tdesign-react';

export const openErrNotification = (title: string, description: string) => {
    NotificationPlugin.error({
        title: title,
        content: description,
        placement: 'top-right',
        duration: 10000,
        offset: [-30, 30],
        closeBtn: true,
    });
};

export const openInfoNotification = (title: string, description: string) => {
    NotificationPlugin.info({
        title: title,
        content: description,
        placement: 'top-right',
        duration: 10000,
        offset: [-30, 30],
        closeBtn: true,
    });
};