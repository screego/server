import * as gen from 'unique-names-generator';

const roomConfig: gen.Config = {
    dictionaries: [gen.adjectives, gen.colors, gen.animals],
    length: 3,
    separator: '-',
};
export const randomRoomName = () => gen.uniqueNamesGenerator(roomConfig);
