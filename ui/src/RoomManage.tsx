import React from 'react';
import {
    Button,
    Checkbox,
    FormControl,
    FormControlLabel,
    Grid,
    IconButton,
    InputLabel,
    MenuItem,
    Paper,
    Select,
    TextField,
    Typography,
    Link,
} from '@material-ui/core';
import {FCreateRoom, UseRoom} from './useRoom';
import {RoomMode, UIConfig} from './message';
import {randomRoomName} from './name';
import HelpIcon from '@material-ui/icons/Help';
import logo from './logo.svg';
import {UseConfig} from './useConfig';
import {LoginForm} from './LoginForm';

const defaultMode = (authMode: UIConfig['authMode'], loggedIn: boolean): RoomMode => {
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

const CreateRoom = ({room, config}: Pick<UseRoom, 'room'> & {config: UIConfig}) => {
    const [id, setId] = React.useState(randomRoomName);
    const [mode, setMode] = React.useState<RoomMode>(defaultMode(config.authMode, config.loggedIn));
    const [ownerLeave, setOwnerLeave] = React.useState(true);
    const submit = () =>
        room({
            type: 'create',
            payload: {
                mode,
                closeOnOwnerLeave: ownerLeave,
                id: id || undefined,
            },
        });
    return (
        <div>
            <FormControl fullWidth>
                <TextField
                    fullWidth
                    value={id}
                    onChange={(e) => setId(e.target.value)}
                    label="id"
                    margin="dense"
                />
                <FormControlLabel
                    control={
                        <Checkbox
                            checked={ownerLeave}
                            onChange={(_, checked) => setOwnerLeave(checked)}
                        />
                    }
                    label="Close Room after you leave"
                />
                <FormControl margin="dense">
                    <InputLabel>NAT Traversal via</InputLabel>
                    <Select
                        fullWidth
                        value={mode}
                        onChange={(x) => setMode(x.target.value as RoomMode)}
                        endAdornment={
                            <IconButton size="small" href="https://screego.net/#/nat-traversal">
                                <HelpIcon />
                            </IconButton>
                        }>
                        <MenuItem
                            value={RoomMode.Stun}
                            disabled={config.authMode === 'all' && !config.loggedIn}>
                            STUN
                        </MenuItem>
                        <MenuItem
                            value={RoomMode.Turn}
                            disabled={config.authMode !== 'none' && !config.loggedIn}>
                            TURN
                        </MenuItem>
                    </Select>
                </FormControl>
                <Button onClick={submit} fullWidth variant="contained">
                    Create Room
                </Button>
            </FormControl>
        </div>
    );
};

export const RoomManage = ({room, config}: {room: FCreateRoom; config: UseConfig}) => {
    const [showLogin, setShowLogin] = React.useState(false);

    const canCreateRoom = config.authMode !== 'all';
    const loginVisible = !config.loggedIn && (showLogin || !canCreateRoom);

    return (
        <Grid
            container={true}
            justify="center"
            style={{paddingTop: 50, maxWidth: 400, width: '100%', margin: '0 auto'}}
            spacing={4}>
            <Grid item xs={12}>
                <Typography align="center" gutterBottom>
                    <img src={logo} style={{width: 230}} alt="logo" />
                </Typography>
                <Paper elevation={3} style={{padding: 20}}>
                    {loginVisible ? (
                        <LoginForm
                            config={config}
                            hide={canCreateRoom ? () => setShowLogin(false) : undefined}
                        />
                    ) : (
                        <>
                            <Typography style={{display: 'flex', alignItems: 'center'}}>
                                <span style={{flex: 1}}>Hello {config.user}!</span>{' '}
                                {config.loggedIn ? (
                                    <Button variant="outlined" size="small" onClick={config.logout}>
                                        Logout
                                    </Button>
                                ) : (
                                    <Button
                                        variant="outlined"
                                        size="small"
                                        onClick={() => setShowLogin(true)}>
                                        Login
                                    </Button>
                                )}
                            </Typography>

                            <CreateRoom room={room} config={config} />
                        </>
                    )}
                </Paper>
            </Grid>
            <div style={{position: 'absolute', margin: '0 auto', bottom: 0}}>
                Screego {config.version} |{' '}
                <Link href="https://github.com/screego/server/">GitHub</Link>
            </div>
        </Grid>
    );
};
