import React from 'react';
import {
    Dialog,
    DialogTitle,
    DialogContent,
    TextField,
    DialogActions,
    Button,
    Autocomplete,
    Box,
} from '@mui/material';
import {
    CodecBestQuality,
    CodecDefault,
    codecName,
    loadSettings,
    PreferredCodec,
    Settings,
    VideoDisplayMode,
} from './settings';
import {NumberField} from './NumberField';

export interface SettingDialogProps {
    open: boolean;
    setOpen: (open: boolean) => void;
    updateName: (s: string) => void;
    saveSettings: (s: Settings) => void;
}

const getAvailableCodecs = (): PreferredCodec[] => {
    if ('getCapabilities' in RTCRtpSender) {
        return RTCRtpSender.getCapabilities('video')?.codecs ?? [];
    }
    return [];
};

const NativeCodecs = getAvailableCodecs();

export const SettingDialog = ({open, setOpen, updateName, saveSettings}: SettingDialogProps) => {
    const [settingsInput, setSettingsInput] = React.useState(loadSettings);

    const doSubmit = () => {
        saveSettings(settingsInput);
        updateName(settingsInput.name ?? '');
        setOpen(false);
    };

    const {name, preferCodec, displayMode, framerate} = settingsInput;

    return (
        <Dialog open={open} onClose={() => setOpen(false)} maxWidth={'xs'} fullWidth>
            <DialogTitle>Settings</DialogTitle>
            <DialogContent>
                <form onSubmit={doSubmit}>
                    <Box paddingBottom={1}>
                        <TextField
                            autoFocus
                            margin="dense"
                            label="Username"
                            value={name}
                            onChange={(e) =>
                                setSettingsInput((c) => ({...c, name: e.target.value}))
                            }
                            fullWidth
                        />
                    </Box>
                    {NativeCodecs.length > 0 ? (
                        <Box paddingY={1}>
                            <Autocomplete<PreferredCodec>
                                options={[CodecBestQuality, CodecDefault, ...NativeCodecs]}
                                getOptionLabel={({mimeType, sdpFmtpLine}) =>
                                    codecName(mimeType) + (sdpFmtpLine ? ` (${sdpFmtpLine})` : '')
                                }
                                value={preferCodec}
                                isOptionEqualToValue={(a, b) =>
                                    a.mimeType === b.mimeType && a.sdpFmtpLine === b.sdpFmtpLine
                                }
                                fullWidth
                                onChange={(_, value) =>
                                    setSettingsInput((c) => ({
                                        ...c,
                                        preferCodec: value ?? undefined,
                                    }))
                                }
                                renderInput={(params) => (
                                    <TextField {...params} label="Preferred Codec" />
                                )}
                            />
                        </Box>
                    ) : undefined}
                    <Box paddingTop={1}>
                        <Autocomplete<VideoDisplayMode>
                            options={Object.values(VideoDisplayMode)}
                            onChange={(_, value) =>
                                setSettingsInput((c) => ({
                                    ...c,
                                    displayMode: value ?? VideoDisplayMode.FitToWindow,
                                }))
                            }
                            value={displayMode}
                            fullWidth
                            renderInput={(params) => <TextField {...params} label="Display Mode" />}
                        />
                    </Box>
                    <Box paddingTop={1}>
                        <NumberField
                            label="FrameRate"
                            min={1}
                            onChange={(framerate) => setSettingsInput((c) => ({...c, framerate}))}
                            value={framerate}
                            fullWidth
                        />
                    </Box>
                </form>
            </DialogContent>
            <DialogActions>
                <Button onClick={() => setOpen(false)} color="primary">
                    Cancel
                </Button>
                <Button onClick={doSubmit} color="primary">
                    Save
                </Button>
            </DialogActions>
        </Dialog>
    );
};
