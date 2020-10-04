import React from 'react';
import {RoomManage} from './RoomManage';
import {useRoom} from './useRoom';
import {Room} from './Room';
import {useConfig} from './useConfig';

export const Router = () => {
    const {room, state, ...other} = useRoom();
    const config = useConfig();

    if (config.loading) {
        // show spinner
        return null;
    }

    if (state) {
        return <Room state={state} {...other} />;
    }

    return <RoomManage room={room} config={config} />;
};
