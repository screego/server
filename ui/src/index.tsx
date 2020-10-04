import React from 'react';
import ReactDOM from 'react-dom';
import './global.css';
import {Button, createMuiTheme, CssBaseline, MuiThemeProvider} from '@material-ui/core';
import {Router} from './Router';
import {SnackbarProvider} from 'notistack';

const theme = createMuiTheme({
    overrides: {
        MuiSelect: {icon: {position: 'relative'}},
        MuiLink: {
            root: {
                color: '#458588',
            },
        },
        MuiIconButton: {
            root: {
                color: 'inherit',
            },
        },
        MuiListItemIcon: {
            root: {
                color: 'inherit',
            },
        },
        MuiToolbar: {
            root: {
                background: '#a89984',
            },
        },
        MuiTooltip: {
            tooltip: {
                fontSize: '1.6em',
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
        type: 'dark',
    },
});

const Snackbar: React.FC = ({children}) => {
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
            )}>
            {children}
        </SnackbarProvider>
    );
};

ReactDOM.render(
    <React.StrictMode>
        <MuiThemeProvider theme={theme}>
            <Snackbar>
                <CssBaseline />
                <Router />
            </Snackbar>
        </MuiThemeProvider>
    </React.StrictMode>,
    document.getElementById('root')
);
