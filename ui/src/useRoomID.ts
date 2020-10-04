import React from 'react';

const getRoomFromURL = (search: string): string | undefined =>
    search
        .slice(1)
        .split('&')
        .find((param) => param.startsWith('room='))
        ?.split('=')[1];

export const useRoomID = (): [string | undefined, (v?: string) => void] => {
    const [state, setState] = React.useState<string | undefined>(() =>
        getRoomFromURL(window.location.search)
    );
    React.useEffect(() => {
        const onChange = (): void => setState(getRoomFromURL(window.location.search));
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
