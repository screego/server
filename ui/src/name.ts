import * as gen from 'unique-names-generator';

const roomConfig: gen.Config = {
    dictionaries: [gen.adjectives, gen.colors, gen.animals],
    length: 3,
    separator: '-',
};
export const randomRoomName = () => gen.uniqueNamesGenerator(roomConfig);

export const getPermanentName = () => localStorage.getItem('screego_name') ?? undefined;
export const setPermanentName = (name: string) => localStorage.setItem('screego_name', name);
