import React from 'react';

export const getRoomFromURL = (): string | undefined => getFromURL('room');

export const getFromURL = (
    key: string,
    search: string = window.location.search
): string | undefined =>
    search
        .slice(1)
        .split('&')
        .find((param) => param.startsWith(`${key}=`))
        ?.split('=')[1];

export const useRoomID = (): [string | undefined, (v?: string) => void] => {
    const [state, setState] = React.useState<string | undefined>(() => getRoomFromURL());
    React.useEffect(() => {
        const onChange = (): void => setState(getRoomFromURL());
        window.addEventListener('popstate', onChange);
        return () => window.removeEventListener('popstate', onChange);
    }, [setState]);
    return [
        state,
        React.useCallback(
            (id) =>
                setState((oldId?: string) => {
                    if (oldId !== id) {
                        window.history.pushState({roomId: id}, '', id ? `?room=${id}` : '?');
                    }
                    return id;
                }),
            [setState]
        ),
    ];
};
