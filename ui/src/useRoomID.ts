import React from 'react';

export const getRoomFromURL = (): string | undefined => getFromURL('room');

export const getFromURL = (
    key: string,
    search: string = window.location.search
): string | undefined => new URLSearchParams(search).get(key) ?? undefined;

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
                        const params = new URLSearchParams();
                        if (id) {
                            params.set('room', id);
                        }
                        const query = params.toString();
                        window.history.pushState({roomId: id}, '', query ? `?${query}` : '?');
                    }
                    return id;
                }),
            [setState]
        ),
    ];
};
