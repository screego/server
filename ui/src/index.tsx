import React from 'react';
import ReactDOM from 'react-dom/client';
import './global.css';
import {Button, createTheme, CssBaseline, ThemeProvider, StyledEngineProvider} from '@mui/material';
import {Router} from './Router';
import {SnackbarProvider} from 'notistack';

const theme = createTheme({
    components: {
        MuiSelect: {
            styleOverrides: {
                icon: {position: 'relative'},
            },
        },
        MuiLink: {
            styleOverrides: {
                root: {
                    color: '#458588',
                },
            },
        },
        MuiIconButton: {
            styleOverrides: {
                root: {
                    color: 'inherit',
                },
            },
        },
        MuiListItemIcon: {
            styleOverrides: {
                root: {
                    color: 'inherit',
                },
            },
        },
        MuiToolbar: {
            styleOverrides: {
                root: {
                    background: '#a89984',
                },
            },
        },
        MuiTooltip: {
            styleOverrides: {
                tooltip: {
                    fontSize: '1.6em',
                },
            },
        },
    },
    palette: {
        background: {
            default: '#282828',
            paper: '#32302f',
        },
        text: {
            primary: '#fbf1d4',
        },
        primary: {
            main: '#a89984',
        },
        secondary: {
            main: '#f44336',
        },
        mode: 'dark',
    },
});

const Snackbar: React.FC<React.PropsWithChildren> = ({children}) => {
    const notistackRef = React.createRef<any>();
    const onClickDismiss = (key: unknown) => () => {
        notistackRef.current?.closeSnackbar(key);
    };

    return (
        <SnackbarProvider
            maxSnack={3}
            ref={notistackRef}
            action={(key) => (
                <Button onClick={onClickDismiss(key)} size="small">
                    Dismiss
                </Button>
            )}
        >
            {children}
        </SnackbarProvider>
    );
};

ReactDOM.createRoot(document.getElementById('root')!!).render(
    <StyledEngineProvider injectFirst>
        <ThemeProvider theme={theme}>
            <Snackbar>
                <CssBaseline />
                <Router />
            </Snackbar>
        </ThemeProvider>
    </StyledEngineProvider>
);
