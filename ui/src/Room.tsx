import React from 'react';
import {setPermanentName} from './name';
import {
    Badge,
    Button,
    Checkbox,
    Dialog,
    DialogActions,
    DialogContent,
    DialogTitle,
    FormControlLabel,
    IconButton,
    Menu,
    MenuItem,
    Paper,
    TextField,
    Theme,
    Tooltip,
    Typography,
} from '@material-ui/core';
import CancelPresentationIcon from '@material-ui/icons/CancelPresentation';
import PresentToAllIcon from '@material-ui/icons/PresentToAll';
import FullScreenIcon from '@material-ui/icons/Fullscreen';
import PeopleIcon from '@material-ui/icons/People';
import ShowMoreIcon from '@material-ui/icons/MoreVert';
import {Video} from './Video';
import {makeStyles} from '@material-ui/core/styles';
import {ConnectedRoom} from './useRoom';
import {useSnackbar} from 'notistack';
import {RoomUser} from './message';

const HostStream: unique symbol = Symbol('mystream');

const flags = (user: RoomUser) => {
    const result: string[] = [];
    if (user.you) {
        result.push('You');
    }
    if (user.owner) {
        result.push('Owner');
    }
    if (user.streaming) {
        result.push('Streaming');
    }
    if (!result.length) {
        return '';
    }
    return ` (${result.join(', ')})`;
};

export const Room = ({
    state,
    share,
    stopShare,
    setName,
}: {
    state: ConnectedRoom;
    share: () => void;
    stopShare: () => void;
    setName: (name: string) => void;
}) => {
    const classes = useStyles();
    const [open, setOpen] = React.useState(false);
    const {enqueueSnackbar} = useSnackbar();
    const [nameInput, setNameInput] = React.useState('');
    const [permanent, setPermanent] = React.useState(false);
    const [showControl, setShowControl] = React.useState(true);
    const [hoverControl, setHoverControl] = React.useState();
    const [showMore, setShowMore] = React.useState<Element>();
    const [selectedStream, setSelectedStream] = React.useState<string | typeof HostStream>();
    const [videoElement, setVideoElement] = React.useState<HTMLVideoElement | null>(null);
    useShowOnMouseMovement(setShowControl);

    React.useEffect(() => {
        if (selectedStream === HostStream && state.hostStream) {
            return;
        }
        if (state.clientStreams.some(({id}) => id === selectedStream)) {
            return;
        }
        if (state.clientStreams.length === 0 && selectedStream) {
            setSelectedStream(undefined);
            return;
        }
        setSelectedStream(state.clientStreams[0]?.id);
    }, [state.clientStreams, selectedStream, state.hostStream]);

    const stream =
        selectedStream === HostStream
            ? state.hostStream
            : state.clientStreams.find(({id}) => selectedStream === id)?.stream;

    const submitName = () => {
        if (permanent) {
            setPermanentName(nameInput);
        }
        setName(nameInput);
        setOpen(false);
    };

    React.useEffect(() => {
        if (videoElement && stream) {
            videoElement.srcObject = stream;
            videoElement.play();
        }
    }, [videoElement, stream]);

    const copyLink = () => {
        navigator?.clipboard?.writeText(window.location.href)?.then(
            () => enqueueSnackbar('Link Copied', {variant: 'success'}),
            (err) => enqueueSnackbar('Copy Failed ' + err, {variant: 'error'})
        );
    };

    const setHoverState = React.useMemo(
        () => ({
            onMouseLeave: () => setHoverControl(false),
            onMouseEnter: () => setHoverControl(true),
        }),
        [setHoverControl]
    );

    const controlVisible = showControl || open || showMore || hoverControl;

    return (
        <div className={classes.videoContainer}>
            {controlVisible && (
                <Paper className={classes.title} elevation={10} {...setHoverState}>
                    <Tooltip title="Copy Link">
                        <Typography
                            variant="h4"
                            component="h4"
                            style={{cursor: 'pointer'}}
                            onClick={copyLink}>
                            {state.id}
                        </Typography>
                    </Tooltip>
                </Paper>
            )}

            {stream ? (
                <video muted ref={setVideoElement} className={classes.video} />
            ) : (
                <Typography
                    variant="h4"
                    align="center"
                    component="div"
                    style={{
                        top: '50%',
                        left: '50%',
                        position: 'absolute',
                        transform: 'translate(-50%, -50%)',
                    }}>
                    no stream available
                </Typography>
            )}

            {controlVisible && (
                <Paper className={classes.control} elevation={10} {...setHoverState}>
                    {state.hostStream ? (
                        <Tooltip title="Cancel Presentation" arrow>
                            <IconButton onClick={stopShare}>
                                <CancelPresentationIcon fontSize="large" />
                            </IconButton>
                        </Tooltip>
                    ) : (
                        <Tooltip title="Start Presentation" arrow>
                            <IconButton onClick={share}>
                                <PresentToAllIcon fontSize="large" />
                            </IconButton>
                        </Tooltip>
                    )}

                    <Tooltip
                        classes={{tooltip: classes.noMaxWidth}}
                        title={
                            <div>
                                <Typography variant="h5">Member List</Typography>
                                {state.users.map((user) => (
                                    <Typography>
                                        {user.name} {flags(user)}
                                    </Typography>
                                ))}
                            </div>
                        }
                        arrow>
                        <Badge badgeContent={state.users.length} color="primary">
                            <PeopleIcon fontSize="large" />
                        </Badge>
                    </Tooltip>
                    <Tooltip title="Fullscreen" arrow>
                        <IconButton
                            onClick={() => videoElement?.requestFullscreen()}
                            disabled={!selectedStream}>
                            <FullScreenIcon fontSize="large" />
                        </IconButton>
                    </Tooltip>
                    <Tooltip title="More" arrow>
                        <IconButton onClick={(e) => setShowMore(e.currentTarget)}>
                            <ShowMoreIcon fontSize="large" />
                        </IconButton>
                    </Tooltip>

                    <Menu
                        anchorEl={showMore}
                        keepMounted
                        open={Boolean(showMore)}
                        onClose={(e) => setShowMore(undefined)}>
                        <MenuItem
                            onClick={() => {
                                setShowMore(undefined);
                                setOpen(true);
                            }}>
                            Change Name
                        </MenuItem>
                    </Menu>
                </Paper>
            )}

            <div className={classes.bottomContainer}>
                {state.clientStreams
                    .filter(({id}) => id !== selectedStream)
                    .map((client) => {
                        return (
                            <Paper
                                elevation={4}
                                className={classes.smallVideoContainer}
                                onClick={() => setSelectedStream(client.id)}>
                                <Video
                                    key={client.id}
                                    src={client.stream}
                                    className={classes.smallVideo}
                                />
                                <Typography
                                    variant="subtitle1"
                                    component="div"
                                    align="center"
                                    className={classes.smallVideoLabel}>
                                    {state.users.find(({id}) => client.peer_id === id)?.name ??
                                        'unknown'}
                                </Typography>
                            </Paper>
                        );
                    })}
                {state.hostStream && selectedStream !== HostStream && (
                    <Paper
                        elevation={4}
                        className={classes.smallVideoContainer}
                        onClick={() => setSelectedStream(HostStream)}>
                        <Video src={state.hostStream} className={classes.smallVideo} />
                        <Typography
                            variant="subtitle1"
                            component="div"
                            align="center"
                            className={classes.smallVideoLabel}>
                            You
                        </Typography>
                    </Paper>
                )}
            </div>

            <Dialog open={open} onClose={() => setOpen(false)}>
                <DialogTitle>Change Name</DialogTitle>
                <DialogContent>
                    <form onSubmit={submitName}>
                        <TextField
                            autoFocus
                            margin="dense"
                            label="Username"
                            value={nameInput}
                            onChange={(e) => setNameInput(e.target.value)}
                            fullWidth
                        />

                        <FormControlLabel
                            control={
                                <Checkbox
                                    checked={permanent}
                                    onChange={(_, checked) => setPermanent(checked)}
                                />
                            }
                            label="Remember"
                        />
                    </form>
                </DialogContent>
                <DialogActions>
                    <Button onClick={() => setOpen(false)} color="primary">
                        Cancel
                    </Button>
                    <Button onClick={submitName} color="primary">
                        Change
                    </Button>
                </DialogActions>
            </Dialog>
        </div>
    );
};

const useShowOnMouseMovement = (doShow: (s: boolean) => void) => {
    const timeoutHandle = React.useRef(0);

    React.useEffect(() => {
        const update = () => {
            if (timeoutHandle.current === 0) {
                doShow(true);
            }

            clearTimeout(timeoutHandle.current);
            timeoutHandle.current = window.setTimeout(() => {
                timeoutHandle.current = 0;
                doShow(false);
            }, 1000);
        };
        window.addEventListener('mousemove', update);
        return () => window.removeEventListener('mousemove', update);
    }, [doShow]);

    React.useEffect(
        () =>
            void (timeoutHandle.current = window.setTimeout(() => {
                timeoutHandle.current = 0;
                doShow(false);
            }, 1000)),
        // eslint-disable-next-line react-hooks/exhaustive-deps
        []
    );
};

const useStyles = makeStyles((theme: Theme) => ({
    title: {
        padding: 15,
        position: 'fixed',
        background: theme.palette.background.paper,
        top: '30px',
        left: '50%',
        transform: 'translateX(-50%)',
        zIndex: 30,
    },
    bottomContainer: {
        position: 'fixed',
        display: 'flex',
        bottom: 0,
        right: 0,
        zIndex: 20,
    },
    control: {
        padding: 15,
        position: 'fixed',
        background: theme.palette.background.paper,
        bottom: '30px',
        left: '50%',
        transform: 'translateX(-50%)',
        zIndex: 30,
    },
    video: {
        maxWidth: '100%',
        maxHeight: '100%',
        width: 'auto',
        height: 'auto',

        position: 'absolute',
        top: '50%',
        left: '50%',
        transform: 'translate(-50%,-50%)',
    },
    smallVideo: {
        minWidth: '100%',
        minHeight: '100%',
        width: 'auto',
        maxWidth: '300px',

        maxHeight: '200px',
    },
    smallVideoLabel: {
        position: 'absolute',
        display: 'block',
        bottom: 0,
        background: 'rgba(0,0,0,.5)',
        padding: '5px 15px',
    },
    noMaxWidth: {
        maxWidth: 'none',
    },
    smallVideoContainer: {
        height: '100%',
        padding: 5,
        maxHeight: 200,
        maxWidth: 400,
        width: '100%',
    },
    videoContainer: {
        position: 'absolute',
        top: 0,
        bottom: 0,
        width: '100%',
        height: '100%',
        overflow: 'hidden',
    },
}));
