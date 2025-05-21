export const getFromURL = (key: string, search: string = window.location.search): string | undefined =>
    search
        .slice(1)
        .split('&')
        .find((param) => param.startsWith(`${key}=`))
        ?.split('=')[1];

export const getRoomFromURL = (): string | undefined => getFromURL('room');

export const useRoomID = (): [string | undefined, (v?: string) => void] => {
    const [RoomId, setRoomId] = React.useState<string | undefined>(() => getRoomFromURL());
    React.useEffect(() => {
        const onChange = (): void => setRoomId(getRoomFromURL());
        window.addEventListener('popstate', onChange);
        return () => window.removeEventListener('popstate', onChange);
    }, [RoomId]);
    return [
        RoomId,
        React.useCallback(
            (id) =>
                setRoomId((oldId?: string) => {
                    if (oldId !== id) {
                        window.history.pushState({roomId: id}, '', id ? `?room=${id}` : '?');
                    }
                    return id;
                }),
            [RoomId]
        ),
    ];
};
