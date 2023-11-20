import {RoomMode, UIConfig} from './message';
import {useSnackbar} from 'notistack';
import React from 'react';
import {urlWithSlash} from './url';

export interface UseConfig extends UIConfig {
    login: (username: string, password: string) => Promise<void>;
    refetch: () => void;
    logout: () => Promise<void>;
    loading: boolean;
}

export const useConfig = (): UseConfig => {
    const {enqueueSnackbar} = useSnackbar();
    const [{loading, ...config}, setConfig] = React.useState<UIConfig & {loading: boolean}>({
        authMode: 'all',
        user: 'guest',
        loggedIn: false,
        loading: true,
        version: 'unknown',
        roomName: 'unknown',
        closeRoomWhenOwnerLeaves: true,
    });

    const refetch = React.useCallback(async () => {
        return fetch(`${urlWithSlash}config`)
            .then((data) => data.json())
            .then(setConfig);
    }, [setConfig]);

    const login = async (username: string, password: string) => {
        const body = new FormData();
        body.set('user', username);
        body.set('pass', password);
        const result = await fetch(`${urlWithSlash}login`, {method: 'POST', body});
        const json = await result.json();
        if (result.status !== 200) {
            enqueueSnackbar('Login Failed: ' + json.message, {variant: 'error'});
        } else {
            await refetch();
            enqueueSnackbar('Logged in!', {variant: 'success'});
        }
    };

    const logout = async () => {
        const result = await fetch(`${urlWithSlash}logout`, {method: 'POST'});
        if (result.status !== 200) {
            enqueueSnackbar('Logout Failed: ' + (await result.text()), {variant: 'error'});
        } else {
            await refetch();
            enqueueSnackbar('Logged Out.', {variant: 'success'});
        }
    };

    // eslint-disable-next-line react-hooks/exhaustive-deps
    React.useEffect(() => void refetch(), []);

    return {...config, refetch, loading, login, logout};
};

export const authModeToRoomMode = (authMode: UIConfig['authMode'], loggedIn: boolean): RoomMode => {
    if (loggedIn) {
        return RoomMode.Turn;
    }
    switch (authMode) {
        case 'all':
            return RoomMode.Turn;
        case 'turn':
            return RoomMode.Stun;
        case 'none':
        default:
            return RoomMode.Turn;
    }
};
